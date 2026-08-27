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
	SweepInterval          = 1 * time.Second
	TTLCountToMarkNodeDead = 3
)

type nodeState struct {
	node     domain.StorageNode
	lastSeen time.Time
}

type StaticNodeRegistry struct {
	ttl time.Duration

	mu      sync.RWMutex
	nodeMap map[uuid.UUID]nodeState

	SweepInterval      time.Duration
	TTLCountToMarkDead int

	log *slog.Logger
}

func NewStaticNodeRegistry(nodes []domain.StorageNode, ttl time.Duration, log *slog.Logger) *StaticNodeRegistry {
	nodeMap := make(map[uuid.UUID]nodeState, len(nodes))

	for _, node := range nodes {
		nodeMap[node.ID] = nodeState{
			node:     node,
			lastSeen: time.Now(),
		}
	}

	return &StaticNodeRegistry{
		ttl:                ttl,
		nodeMap:            nodeMap,
		SweepInterval:      SweepInterval,
		TTLCountToMarkDead: TTLCountToMarkNodeDead,
		log:                log,
	}
}

func (r *StaticNodeRegistry) ProcessHeartbeat(ctx context.Context, node domain.StorageNode) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.nodeMap[node.ID]
	r.nodeMap[node.ID] = nodeState{
		node:     node,
		lastSeen: time.Now(),
	}

	if exists {
		//	r.log.DebugContext(ctx, "node heartbeat", slog.String("node id", node.ID.String()))
	} else {
		r.log.InfoContext(ctx, "new node", slog.String("node id", node.ID.String()))
	}
}

func (r *StaticNodeRegistry) RunSweeper(ctx context.Context) {
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

func (r *StaticNodeRegistry) GetAllNodes(_ context.Context) ([]domain.StorageNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]domain.StorageNode, 0, len(r.nodeMap))

	for _, state := range r.nodeMap {
		nodes = append(nodes, state.node)
	}

	return nodes, nil
}

func (r *StaticNodeRegistry) GetNode(_ context.Context, id uuid.UUID) (domain.StorageNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.nodeMap[id]
	if !ok {
		return domain.StorageNode{}, fmt.Errorf("%w: id %s", domain.ErrStorageNodeNotFound, id)
	}

	return state.node, nil
}

func (r *StaticNodeRegistry) checkNodes(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, state := range r.nodeMap {
		diff := time.Since(state.lastSeen)

		if diff > time.Duration(r.TTLCountToMarkDead)*r.ttl {
			delete(r.nodeMap, id)
			r.log.InfoContext(ctx, "node died", slog.String("node id", id.String()))
		}
	}
}
