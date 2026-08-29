package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type NodeRegistry interface {
	GetAllNodes(context.Context) ([]domain.StorageNode, error)
}

type NodeManager struct {
	registry NodeRegistry

	mu    sync.RWMutex
	nodes map[uuid.UUID]*Client

	log *slog.Logger
}

func NewNodeManager(registry NodeRegistry, log *slog.Logger) *NodeManager {
	return &NodeManager{
		registry: registry,
		nodes:    make(map[uuid.UUID]*Client),
		log:      log,
	}
}

func (m *NodeManager) DeleteObjectOn(ctx context.Context, nodeID uuid.UUID, path string) error {
	m.mu.RLock()
	node, ok := m.nodes[nodeID]
	m.mu.RUnlock()
	if ok {
		return node.DeleteObject(ctx, path)
	}

	allNodes, err := m.registry.GetAllNodes(ctx)
	if err != nil {
		return fmt.Errorf("error updating node map: error getting all nodes: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var client *Client
	m.nodes = make(map[uuid.UUID]*Client, len(allNodes))
	for _, node := range allNodes {
		c, err := New(node.Address)
		if err != nil {
			if node.ID == nodeID {
				return err
			} else {
				m.log.ErrorContext(
					ctx, "error creating node client for alive node",
					slog.String("error", err.Error()),
					slog.String("node id", node.ID.String()),
					slog.String("node address", node.Address),
				)
			}
		}

		m.nodes[node.ID] = c
		if nodeID == node.ID {
			client = c
		}
	}

	if client == nil {
		fmt.Errorf("error getting node with ID %s", nodeID)
	}

	return client.DeleteObject(ctx, path)
}

func (m *NodeManager) CloseAll() error {
	var errs error

	for _, client := range m.nodes {
		errs = errors.Join(errs, client.Close())
	}

	return errs
}
