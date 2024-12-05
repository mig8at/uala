package main

import (
	"timeline-service/config"
	"timeline-service/internal/application"
	"timeline-service/internal/infrastructure/cron"
	"timeline-service/internal/infrastructure/http"
	"timeline-service/internal/infrastructure/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()
	engine := gin.Default()
	validate := validator.New()
	redis := cfg.Redis()

	precess := cron.NewCron(redis)

	go precess.ProcessTweets()

	// Inicializar repositorio
	repo := repository.NewRepository(redis)

	// Inicializar servicios
	service := application.NewService(repo)

	httpServer := http.NewHTTPServer(engine, service, validate)

	httpServer.Run(cfg.Port)

}
