package application

import (
	"context"
	"errors"
	"testing"
	"user_service/internal/application/dto"
	"user_service/internal/domain/models"
	"user_service/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Create(t *testing.T) {
	// Crear un mock del UserRepository
	mockRepo := new(mocks.UserRepository)
	service := NewService(mockRepo)

	// Definir el input
	input := &dto.CreateUser{
		Name:     "Test User",
		Email:    "test@example.com",
		Nickname: "@testuser",
		Bio:      "Test bio",
		Avatar:   "https://example.com/avatar.png",
	}

	// Definir el output esperado
	createdUser := &models.User{
		ID:        "12345",
		Name:      input.Name,
		Email:     input.Email,
		Nickname:  input.Nickname,
		Bio:       input.Bio,
		Avatar:    input.Avatar,
		Followers: 0,
		Following: 0,
	}

	// Configurar el mock para esperar la llamada a Create y devolver el usuario creado
	mockRepo.On("Create", mock.Anything, input).Return(createdUser, nil)

	// Ejecutar el método a probar
	result, err := service.Create(context.Background(), input)

	// Aserciones
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, result.ID)

	// Verificar que el mock fue llamado como se esperaba
	mockRepo.AssertExpectations(t)
}

func TestUserService_Create_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	service := NewService(mockRepo)

	input := &dto.CreateUser{
		Name:     "Test User",
		Email:    "test@example.com",
		Nickname: "@testuser",
		Bio:      "Test bio",
		Avatar:   "https://example.com/avatar.png",
	}

	mockRepo.On("Create", mock.Anything, input).Return(nil, errors.New("database error"))

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
