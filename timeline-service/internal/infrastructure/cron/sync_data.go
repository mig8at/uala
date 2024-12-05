package cron

import (
	"context"
	"log"
	"time"
	"timeline-service/internal/interfaces"

	"github.com/robfig/cron/v3"
)

type SyncData interface {
	Start(ctx context.Context) error
	Stop()
}

type syncData struct {
	cron     *cron.Cron
	interval string
	service  interfaces.Service
}

func NewSyncData(interval string, service interfaces.Service) SyncData {
	return &syncData{
		cron:     cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.Default()))),
		interval: interval,
		service:  service,
	}
}

func (s *syncData) Start(ctx context.Context) error {

	_, err := s.cron.AddFunc(s.interval, func() {
		if err := s.service.SyncData(context.Background()); err != nil {
			log.Printf("Error al sincronizar datos: %v", err)
		} else {
			log.Println("Sincronización de datos exitosa")
		}
	})
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s *syncData) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.cron.Stop()
	<-ctx.Done()
}
