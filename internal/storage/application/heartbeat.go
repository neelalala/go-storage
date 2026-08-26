package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type DiscoveryService interface {
	Heartbeat(context.Context, uuid.UUID) error
}

func RunHeartbeat(ctx context.Context, service DiscoveryService, interval time.Duration, id uuid.UUID, log *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("heartbeat stopped")
			return
		case <-ticker.C:
			err := service.Heartbeat(ctx, id)
			if err != nil {
				log.Error("error heartbeat", slog.String("error", err.Error()))
			}
		}
	}
}
