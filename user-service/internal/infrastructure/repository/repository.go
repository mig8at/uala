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

func (r *repository) Create(ctx context.Context, createUser *dto.CreateUser) (*models.User, error) {
	userModel := &models.User{
		Name:     createUser.Name,
		Nickname: createUser.Nickname,
		Email:    createUser.Email,
		Bio:      createUser.Bio,
		Avatar:   createUser.Avatar,
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
	r.redis.Set(ctx, key, userData, 0).Err()

	return userModel, nil
}

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
