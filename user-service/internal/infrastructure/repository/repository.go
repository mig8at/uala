package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"user_service/internal/application/dto"
	"user_service/internal/domain/models"
	"user_service/internal/interfaces"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) interfaces.UserRepository {
	return &repository{db: db, redis: redis}
}

// CreateUser crea un nuevo usuario en la base de datos y lo almacena en Redis.
func (r *repository) Create(ctx context.Context, createUser *dto.CreateUser) (*models.User, error) {
	userModel := &models.User{
		Name:     createUser.Name,
		Nickname: createUser.Nickname,
		Email:    createUser.Email,
	}

	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, fmt.Errorf("error al crear el usuario: %w", err)
	}

	// Serializar el usuario para Redis
	userData, err := json.Marshal(struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}{
		ID:       userModel.ID,
		Name:     userModel.Name,
		Nickname: userModel.Nickname,
		Avatar:   userModel.Avatar,
	})
	if err != nil {
		return nil, fmt.Errorf("error al serializar el usuario: %w", err)
	}

	// Almacenar en Redis
	key := fmt.Sprintf("users:%s", userModel.ID)
	if err := r.redis.Set(ctx, key, userData, 0).Err(); err != nil {
		return nil, fmt.Errorf("error al guardar el usuario en Redis: %w", err)
	}

	return userModel, nil
}

// GetUserByID obtiene un usuario por su ID.
func (r *repository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	if err := r.db.WithContext(ctx).First(user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al obtener el usuario: %w", err)
	}
	return user, nil
}

// PaginateUsers devuelve una lista paginada de usuarios ordenados por seguidores.
func (r *repository) Paginate(ctx context.Context, page, limit int) ([]models.User, error) {
	offset := (page - 1) * limit
	var users []models.User

	if err := r.db.WithContext(ctx).
		Order("followers DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, fmt.Errorf("error al obtener usuarios: %w", err)
	}

	return users, nil
}

// Follow permite que un usuario siga a otro usuario.
func (r *repository) Follow(ctx context.Context, userID, followerID string) error {
	if userID == followerID {
		return fmt.Errorf("un usuario no puede seguirse a sí mismo")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user, follower models.User

		// Verificar que ambos usuarios existan
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("usuario a seguir no encontrado")
			}
			return fmt.Errorf("error al obtener el usuario a seguir: %w", err)
		}

		if err := tx.First(&follower, "id = ?", followerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("usuario seguidor no encontrado")
			}
			return fmt.Errorf("error al obtener el usuario seguidor: %w", err)
		}

		// Verificar si ya sigue al usuario
		var count int64
		if err := tx.Model(&models.Follower{}).
			Where("user_id = ? AND follower_id = ?", userID, followerID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("error al verificar seguimiento: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("el usuario ya sigue a este usuario")
		}

		// Crear la relación de seguimiento
		followerRecord := &models.Follower{
			UserID:     userID,
			FollowerID: followerID,
		}
		if err := tx.Create(followerRecord).Error; err != nil {
			return fmt.Errorf("error al crear el registro de seguimiento: %w", err)
		}

		// Actualizar contadores de seguidores y seguidos
		if err := tx.Model(&user).UpdateColumn("followers", gorm.Expr("followers + ?", 1)).Error; err != nil {
			return fmt.Errorf("error al incrementar los seguidores: %w", err)
		}

		if err := tx.Model(&follower).UpdateColumn("following", gorm.Expr("following + ?", 1)).Error; err != nil {
			return fmt.Errorf("error al incrementar los seguidos: %w", err)
		}

		// Actualizar Redis dentro de la transacción
		pipe := r.redis.TxPipeline()
		pipe.SAdd(ctx, fmt.Sprintf("following:%s", followerID), userID)
		pipe.SAdd(ctx, fmt.Sprintf("followers:%s", userID), followerID)
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("error al actualizar Redis: %w", err)
		}

		return nil
	})

	return err
}

// Unfollow permite que un usuario deje de seguir a otro usuario.
func (r *repository) Unfollow(ctx context.Context, userID, followerID string) error {
	if userID == followerID {
		return fmt.Errorf("un usuario no puede dejar de seguirse a sí mismo")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verificar que la relación de seguimiento exista
		var followerRecord models.Follower
		if err := tx.Where("user_id = ? AND follower_id = ?", userID, followerID).First(&followerRecord).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("el usuario no sigue a este usuario")
			}
			return fmt.Errorf("error al verificar seguimiento: %w", err)
		}

		// Eliminar la relación de seguimiento
		if err := tx.Delete(&followerRecord).Error; err != nil {
			return fmt.Errorf("error al eliminar el registro de seguimiento: %w", err)
		}

		// Actualizar contadores de seguidores y seguidos
		if err := tx.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("followers", gorm.Expr("followers - ?", 1)).Error; err != nil {
			return fmt.Errorf("error al decrementar los seguidores: %w", err)
		}

		if err := tx.Model(&models.User{}).Where("id = ?", followerID).UpdateColumn("following", gorm.Expr("following - ?", 1)).Error; err != nil {
			return fmt.Errorf("error al decrementar los seguidos: %w", err)
		}

		// Actualizar Redis dentro de la transacción
		pipe := r.redis.TxPipeline()
		pipe.SRem(ctx, fmt.Sprintf("following:%s", followerID), userID)
		pipe.SRem(ctx, fmt.Sprintf("followers:%s", userID), followerID)
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("error al actualizar Redis: %w", err)
		}

		return nil
	})

	return err
}

// GetFollowers obtiene una lista paginada de los seguidores de un usuario.
func (r *repository) Followers(ctx context.Context, userID string, page, limit int) ([]models.User, error) {
	offset := (page - 1) * limit
	var users []models.User

	if err := r.db.WithContext(ctx).
		Model(&models.Follower{}).
		Select("users.*").
		Joins("JOIN users ON followers.follower_id = users.id").
		Where("followers.user_id = ?", userID).
		Order("users.followers DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, fmt.Errorf("error al obtener seguidores: %w", err)
	}

	return users, nil
}

// GetFollowing obtiene una lista paginada de los usuarios que sigue un usuario.
func (r *repository) Following(ctx context.Context, userID string, page, limit int) ([]models.User, error) {
	offset := (page - 1) * limit
	var users []models.User

	if err := r.db.WithContext(ctx).
		Model(&models.Follower{}).
		Select("users.*").
		Joins("JOIN users ON followers.user_id = users.id").
		Where("followers.follower_id = ?", userID).
		Order("users.followers DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return nil, fmt.Errorf("error al obtener seguidos: %w", err)
	}

	return users, nil
}
