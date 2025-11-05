package nonce

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	DefaultTimeout = 10 * time.Second
)

type Manager struct {
	mu            sync.Mutex
	pendingNonces map[common.Address]uint64
	client        *ethclient.Client
}

func New(client *ethclient.Client) *Manager {
	return &Manager{
		pendingNonces: make(map[common.Address]uint64),
		client:        client,
	}
}

func (m *Manager) GetNext(address common.Address) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if nonce, exists := m.pendingNonces[address]; exists {
		m.pendingNonces[address]++
		return nonce, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	nonce, err := m.client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	m.pendingNonces[address] = nonce + 1
	return nonce, nil
}

func (m *Manager) Reset(address common.Address) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pendingNonces, address)
}

func (m *Manager) GetCached(address common.Address) (uint64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	nonce, exists := m.pendingNonces[address]
	return nonce, exists
}

func (m *Manager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pendingNonces = make(map[common.Address]uint64)
}
