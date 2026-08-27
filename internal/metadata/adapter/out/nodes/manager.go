package nodes

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type NodeRegistry interface {
	GetAllNodes(context.Context) ([]domain.StorageNode, error)
}

type RoundRobinNodeManager struct {
	registry NodeRegistry

	counter atomic.Uint64
}

func NewRoundRobinNodeManager(registry NodeRegistry) *RoundRobinNodeManager {
	return &RoundRobinNodeManager{
		registry: registry,
	}
}

func (m *RoundRobinNodeManager) NextNode(ctx context.Context) (domain.StorageNode, error) {
	nodes, err := m.registry.GetAllNodes(ctx)
	if err != nil {
		return domain.StorageNode{}, err
	}

	if len(nodes) == 0 {
		return domain.StorageNode{}, errors.New("no alive storage nodes")
	}

	slices.SortFunc(nodes, func(a, b domain.StorageNode) int {
		return strings.Compare(a.ID.String(), b.ID.String())
	})

	idx := m.counter.Add(1)

	node := nodes[idx%uint64(len(nodes))]

	return node, nil
}
