package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"user_service/internal/application/dto"
	"user_service/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHTTPServer_CreateUser(t *testing.T) {
	// Configurar Gin en modo test
	gin.SetMode(gin.TestMode)

	// Crear un mock del UserService
	mockService := new(mocks.UserService)
	validate := validator.New()

	// Crear el servidor HTTP con el mock
	server := NewHTTPServer(gin.New(), mockService, validate)

	// Definir el input y el output esperado
	input := dto.CreateUser{
		Name:     "Test User",
		Email:    "test@example.com",
		Nickname: "testuser",
		Bio:      "Test bio",
		Avatar:   "https://example.com/avatar.png",
	}

	createdUser := dto.User{
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
	mockService.On("Create", mock.Anything, &input).Return(&createdUser, nil)

	// Convertir el input a JSON
	body, err := json.Marshal(input)
	assert.NoError(t, err)

	// Crear una solicitud HTTP POST
	req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Crear un recorder para capturar la respuesta
	recorder := httptest.NewRecorder()

	// Ejecutar la solicitud
	server.engine.ServeHTTP(recorder, req)

	fmt.Println(recorder.Body.String())

	// Verificar el código de estado
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Verificar el cuerpo de la respuesta
	var response dto.User
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, createdUser, response)

	// Verificar que el mock fue llamado
	mockService.AssertExpectations(t)
}

func TestHTTPServer_CreateUser_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(mocks.UserService)
	validate := validator.New()
	server := NewHTTPServer(gin.New(), mockService, validate)

	// Input inválido (falta el nombre)
	input := map[string]interface{}{
		"email":    "test@example.com",
		"nickname": "testuser",
		"bio":      "Test bio",
		"avatar":   "https://example.com/avatar.png",
	}

	body, err := json.Marshal(input)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.engine.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}
