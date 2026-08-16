package registry

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

var _ domain.NodeRegistry = (*RoundRobinStaticNodeRegistry)(nil)

type RoundRobinStaticNodeRegistry struct {
	nodes   []domain.StorageNode
	nodeMap map[uuid.UUID]domain.StorageNode

	counter atomic.Uint64
}

func NewRoundRobinStaticNodeRegistry(nodes []domain.StorageNode) *RoundRobinStaticNodeRegistry {
	nodeMap := make(map[uuid.UUID]domain.StorageNode, len(nodes))

	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	return &RoundRobinStaticNodeRegistry{
		nodes:   nodes,
		nodeMap: nodeMap,
	}
}

func (m *RoundRobinStaticNodeRegistry) NextNode(ctx context.Context) (domain.StorageNode, error) {
	if len(m.nodes) == 0 {
		return domain.StorageNode{}, errors.New("no storage nodes available")
	}

	idx := m.counter.Add(1)

	node := m.nodes[idx%uint64(len(m.nodes))]

	return node, nil
}

func (m *RoundRobinStaticNodeRegistry) GetNode(ctx context.Context, id uuid.UUID) (domain.StorageNode, error) {
	node, ok := m.nodeMap[id]
	if !ok {
		return domain.StorageNode{}, fmt.Errorf("no node with id %s", id)
	}

	return node, nil
}
