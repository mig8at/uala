package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"user_service/internal/application/dto"
	"user_service/internal/domain/models"
	"user_service/internal/interfaces"

	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(sqlite *gorm.DB, redis *redis.Client) interfaces.UserRepository {
	return &repository{db: sqlite, redis: redis}
}

func (r *repository) Create(ctx context.Context, createUser *dto.CreateUser) (*models.User, error) {

	userModel := &models.User{}
	if err := copier.Copy(userModel, createUser); err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, err
	}

	rUser, err := redisUser(userModel)
	if err != nil {
		return userModel, nil
	}

	key := fmt.Sprintf("users:%s", userModel.ID)
	err = r.redis.Set(ctx, key, rUser, 0).Err()

	if err != nil {
		fmt.Printf("Error al guardar en Redis: %v\n", err)
	}

	return userModel, nil
}

func (r *repository) GetById(ctx context.Context, id string) (*models.User, error) {

	user := &models.User{}

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(user).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, err
	}

	return user, nil
}

func (r *repository) Paginate(ctx context.Context, page, limit int) ([]models.User, error) {

	offset := (page - 1) * limit

	var users []models.User

	if err := r.db.WithContext(ctx).Order("followers desc").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, err
	}

	return users, nil
}

func (r *repository) Follow(ctx context.Context, id, followerID string) error {

	if id == followerID {
		return fmt.Errorf("un usuario no puede seguirse a sí mismo")
	}

	var user, follower models.User

	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("id = ?", id).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("usuario a seguir no encontrado")
			}
			return fmt.Errorf("error al obtener el usuario a seguir: %v", err)
		}

		if err := tx.Where("id = ?", followerID).First(&follower).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("usuario seguidor no encontrado")
			}
			return fmt.Errorf("error al obtener el usuario seguidor: %v", err)
		}

		var existingFollower models.Follower
		if err := tx.Where("user_id = ? AND follower_id = ?", id, followerID).First(&existingFollower).Error; err == nil {
			return fmt.Errorf("el usuario ya sigue a este usuario")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("error al verificar si el usuario ya sigue: %v", err)
		}

		followerRecord := &models.Follower{
			UserID:     id,
			FollowerID: followerID,
		}
		if err := tx.Create(followerRecord).Error; err != nil {
			return fmt.Errorf("error al crear el registro de seguidor: %v", err)
		}

		if err := tx.Model(&user).UpdateColumn("followers", gorm.Expr("followers + ?", 1)).Error; err != nil {
			return fmt.Errorf("error al incrementar los seguidores del usuario: %v", err)
		}

		if err := tx.Model(&follower).UpdateColumn("following", gorm.Expr("following + ?", 1)).Error; err != nil {
			return fmt.Errorf("error al incrementar los siguiendo del usuario seguidor: %v", err)
		}

		return nil
	})

	err := r.redis.SAdd(ctx, fmt.Sprintf("following:%s", id), followerID).Err()
	if err != nil {
		return err
	}

	err = r.redis.SAdd(ctx, fmt.Sprintf("followers:%s", followerID), id).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) Unfollow(ctx context.Context, id, followerID string) error {

	if id == followerID {
		return fmt.Errorf("un usuario no puede dejar de seguirse a sí mismo")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user, follower models.User

		if err := tx.Where("id = ?", id).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("usuario a dejar de seguir no encontrado")
			}
			return fmt.Errorf("error al obtener el usuario a dejar de seguir: %v", err)
		}

		if err := tx.Where("id = ?", followerID).First(&follower).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("usuario seguidor no encontrado")
			}
			return fmt.Errorf("error al obtener el usuario seguidor: %v", err)
		}

		var existingFollower models.Follower
		if err := tx.Where("user_id = ? AND follower_id = ?", id, followerID).First(&existingFollower).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("el usuario no sigue a este usuario")
			}
			return fmt.Errorf("error al verificar si el usuario ya sigue: %v", err)
		}

		if err := tx.Delete(&existingFollower).Error; err != nil {
			return fmt.Errorf("error al eliminar el registro de seguidor: %v", err)
		}

		if err := tx.Model(&user).UpdateColumn("followers", gorm.Expr("followers - ?", 1)).Error; err != nil {
			return fmt.Errorf("error al decrementar los seguidores del usuario: %v", err)
		}

		if err := tx.Model(&follower).UpdateColumn("following", gorm.Expr("following - ?", 1)).Error; err != nil {
			return fmt.Errorf("error al decrementar los siguiendo del usuario seguidor: %v", err)
		}

		return nil
	})

}

func (r *repository) Followers(ctx context.Context, id string, page, limit int) ([]models.User, error) {

	offset := (page - 1) * limit

	var users []models.User

	if err := r.db.WithContext(ctx).
		Table("followers").
		Select("users.*").
		Joins("JOIN users ON followers.follower_id = users.id").
		Where("followers.user_id = ?", id).
		Order("users.followers desc").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, err
	}

	return users, nil
}

func (r *repository) Following(ctx context.Context, id string, page, limit int) ([]models.User, error) {

	offset := (page - 1) * limit

	var users []models.User

	if err := r.db.WithContext(ctx).
		Table("followers").
		Select("users.*").
		Joins("JOIN users ON followers.user_id = users.id").
		Where("followers.follower_id = ?", id).
		Order("users.followers desc").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, err
	}

	return users, nil
}

func redisUser(u *models.User) ([]byte, error) {
	jsonData, err := json.Marshal(struct {
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}{
		Name:     u.Name,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	})
	if err != nil {
		return nil, fmt.Errorf("error al serializar el Tweet a JSON: %w", err)
	}
	return jsonData, nil
}
