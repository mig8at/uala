package main

import (
	"tweet-service/config"
	"tweet-service/internal/application"
	"tweet-service/internal/infrastructure/http"
	"tweet-service/internal/infrastructure/repository"
	"tweet-service/internal/infrastructure/seeder"

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
		seed := seeder.NewSeeder(sqlite, redis, repo)
		seed.Seed()
	}

	service := application.NewService(repo)

	httpServer := http.NewHTTPServer(engine, service, validate)
	httpServer.Run(cfg.Port)
}
