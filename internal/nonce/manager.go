package nonce

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Manager handles thread-safe nonce tracking
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

// GetNext returns next nonce (thread-safe)
func (m *Manager) GetNext(address common.Address) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if nonce, exists := m.pendingNonces[address]; exists {
		m.pendingNonces[address]++
		return nonce, nil
	}

	nonce, err := m.client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0, err
	}

	m.pendingNonces[address] = nonce + 1
	return nonce, nil
}

func (m *Manager) Reset(address common.Address) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pendingNonces, address)
}
