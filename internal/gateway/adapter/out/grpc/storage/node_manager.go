package storage

import (
	"sync"

	"github.com/neelalala/go-storage/internal/gateway/domain"
	"google.golang.org/grpc/connectivity"
)

var _ domain.StorageNodeManager = (*NodeManager)(nil)

type NodeManager struct {
	conns map[string]*Client
	mu    sync.RWMutex
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		conns: make(map[string]*Client),
	}
}

func (m *NodeManager) GetStorage(address string) (domain.Storage, error) {
	m.mu.RLock()
	client, ok := m.conns[address]
	m.mu.RUnlock()

	if ok {
		return client, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok = m.conns[address]
	if ok {
		return client, nil
	}

	client, err := NewClient(address)
	if err != nil {
		return nil, err
	}

	m.conns[address] = client
	return client, nil
}

func (m *NodeManager) Invalidate(address string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.conns[address]
	if !ok {
		return
	}

	state := client.conn.GetState()
	if state == connectivity.TransientFailure ||
		state == connectivity.Shutdown {
		client.conn.Close()
		delete(m.conns, address)
	}
}
