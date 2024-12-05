package main

import (
	"context"
	"log"
	"timeline-service/config"
	"timeline-service/internal/application"
	"timeline-service/internal/infrastructure/clients"
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
	bager := cfg.Badger()

	// Inicializar repositorio
	repo := repository.NewRepository(bager)

	client := clients.NewClient()
	// Inicializar servicios
	service := application.NewService(repo, client)

	// Inicializar el cron con el servicio
	cronJob := cron.NewSyncData("@every 3s", service)

	// Iniciar el cron
	ctx := context.Background()
	if err := cronJob.Start(ctx); err != nil {
		log.Fatalf("Error al iniciar el cron: %v", err)
	}

	httpServer := http.NewHTTPServer(engine, service, validate)
	// Ejecutar el servidor HTTP
	go httpServer.Run(cfg.Port)

	// Esperar a que el contexto se cancele
	<-ctx.Done()

	// Detener el cron
	cronJob.Stop()
}
