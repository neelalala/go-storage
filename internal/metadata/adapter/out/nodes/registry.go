package nodes

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

const (
	SweepInterval                  = 1 * time.Second
	NoHeartbeatCountToMarkNodeDead = 3
)

type nodeState struct {
	node     domain.StorageNode
	lastSeen time.Time
}

type DynamicNodeRegistry struct {
	heartbeatInterval time.Duration

	mu      sync.RWMutex
	nodeMap map[uuid.UUID]nodeState

	SweepInterval      time.Duration
	TTLCountToMarkDead int

	log *slog.Logger
}

func NewNodeRegistry(heartbeatInterval time.Duration, log *slog.Logger) *DynamicNodeRegistry {
	nodeMap := make(map[uuid.UUID]nodeState)

	return &DynamicNodeRegistry{
		heartbeatInterval:  heartbeatInterval,
		nodeMap:            nodeMap,
		SweepInterval:      SweepInterval,
		TTLCountToMarkDead: NoHeartbeatCountToMarkNodeDead,
		log:                log,
	}
}

func (r *DynamicNodeRegistry) ProcessHeartbeat(ctx context.Context, node domain.StorageNode) {
	r.mu.Lock()

	_, exists := r.nodeMap[node.ID]
	r.nodeMap[node.ID] = nodeState{
		node:     node,
		lastSeen: time.Now(),
	}

	r.mu.Unlock()

	if !exists {
		r.log.InfoContext(
			ctx, "new node",
			slog.Group(
				"node",
				slog.String("id", node.ID.String()),
				slog.String("address", node.Address),
			),
		)
	}
}

func (r *DynamicNodeRegistry) RunSweeper(ctx context.Context) {
	ticker := time.NewTicker(r.SweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.checkNodes(ctx)
		}
	}
}

func (r *DynamicNodeRegistry) GetAllNodes(_ context.Context) ([]domain.StorageNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]domain.StorageNode, 0, len(r.nodeMap))

	for _, state := range r.nodeMap {
		nodes = append(nodes, state.node)
	}

	return nodes, nil
}

func (r *DynamicNodeRegistry) GetNode(_ context.Context, id uuid.UUID) (domain.StorageNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.nodeMap[id]
	if !ok {
		return domain.StorageNode{}, fmt.Errorf("%w: id %s", domain.ErrStorageNodeNotFound, id)
	}

	return state.node, nil
}

func (r *DynamicNodeRegistry) checkNodes(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, state := range r.nodeMap {
		diff := time.Since(state.lastSeen)

		if diff > time.Duration(r.TTLCountToMarkDead)*r.heartbeatInterval {
			delete(r.nodeMap, id)
			r.log.InfoContext(
				ctx, "node died",
				slog.Group(
					"node",
					slog.String("id", id.String()),
					slog.String("address", state.node.Address),
				),
			)
		}
	}
}
