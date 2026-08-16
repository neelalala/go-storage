package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type NodeManager struct {
	nodes map[uuid.UUID]*Client
}

func NewNodeManager(nodes []domain.StorageNode) (*NodeManager, error) {
	nodesMap := make(map[uuid.UUID]*Client, len(nodes))

	for _, node := range nodes {
		client, err := New(node.Address)
		if err != nil {
			return nil, fmt.Errorf("error creating grcp client fot storage with address %s: %v", node.Address, err)
		}

		nodesMap[node.ID] = client
	}

	return &NodeManager{
		nodes: nodesMap,
	}, nil
}

func (m *NodeManager) DeleteObjectOn(ctx context.Context, nodeID uuid.UUID, path string) error {
	node, ok := m.nodes[nodeID]
	if !ok {
		return fmt.Errorf("error getting node with ID %s", nodeID)
	}

	return node.DeleteObject(ctx, path)
}

func (m *NodeManager) CloseAll() error {
	var errs error

	for _, client := range m.nodes {
		errs = errors.Join(errs, client.Close())
	}

	return errs
}
