package main

import (
	"user_service/config"
	"user_service/internal/application"
	"user_service/internal/infrastructure/http"
	"user_service/internal/infrastructure/repository"
	"user_service/internal/infrastructure/seeder"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()
	engine := gin.Default()
	validate := validator.New()
	sqlite := cfg.Sqlite()
	redis := cfg.Redis()

	// Inicializar repositorio
	repo := repository.NewRepository(sqlite, redis)

	// Ejecutar el seeder solo en entornos de desarrollo o prueba
	if cfg.Env == "development" || cfg.Env == "test" {
		// Datos de prueba
		seed := seeder.NewSeeder(sqlite, redis)
		seed.Seed()
	}

	service := application.NewService(repo)

	httpServer := http.NewHTTPServer(engine, service, validate)
	httpServer.Run(cfg.Port)
}

//mockery --name=UserRepository --dir=./internal/interfaces --output=./internal/mocks --outpkg=mocks --filename=repository_mock.go
