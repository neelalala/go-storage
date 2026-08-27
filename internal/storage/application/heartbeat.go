package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type DiscoveryService interface {
	Heartbeat(context.Context, uuid.UUID, string) error
}

func RunHeartbeat(ctx context.Context, service DiscoveryService, interval time.Duration, id uuid.UUID, addr string, log *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("heartbeat stopped")
			return
		case <-ticker.C:
			err := service.Heartbeat(ctx, id, addr)
			if err != nil {
				log.Error("error heartbeat", slog.String("error", err.Error()))
			}
		}
	}
}
