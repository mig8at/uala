package repository

import (
	"context"
	"testing"
	"user_service/internal/application/dto"
	"user_service/internal/domain/models"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRepository_Create(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to in-memory database: %v", err)
	}

	redisTest := redis.NewClient(&redis.Options{})

	// Migrar los modelos
	err = db.AutoMigrate(&models.User{}, &models.Follower{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	repo := NewRepository(db, redisTest)

	input := &dto.CreateUser{
		Name:     "Test User",
		Email:    "test@example.com",
		Nickname: "@testuser",
		Bio:      "Test bio",
		Avatar:   "https://example.com/avatar.png",
	}

	user, err := repo.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, input.Name, user.Name)
	assert.Equal(t, input.Email, user.Email)
	assert.Equal(t, input.Nickname, user.Nickname)
	assert.Equal(t, input.Bio, user.Bio)
	assert.Equal(t, input.Avatar, user.Avatar)
	assert.Equal(t, 0, user.Followers)
	assert.Equal(t, 0, user.Following)

	// Verificar que el usuario se ha guardado en la base de datos
	var dbUser models.User
	err = db.First(&dbUser, "id = ?", user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, dbUser.ID)
}
