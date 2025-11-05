package pool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ClientPool struct {
	clients []*ethclient.Client
	current int
	mu      sync.RWMutex
	closed  bool
}

// New creates a new client pool
// rpcURL: Ethereum node RPC endpoint (e.g., "http://localhost:8545")
// size: Number of connections to create (minimum 1)
func New(rpcURL string, size int) (*ClientPool, error) {
	if size < 1 {
		size = 1
	}

	clients := make([]*ethclient.Client, size)

	// Create all clients
	for i := 0; i < size; i++ {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			// Clean up any clients we already created
			for j := 0; j < i; j++ {
				clients[j].Close()
			}
			return nil, fmt.Errorf("failed to create client %d: %w", i, err)
		}
		clients[i] = client
	}

	return &ClientPool{
		clients: clients,
		current: 0,
		closed:  false,
	}, nil
}

// Get returns the next client in the pool using round-robin
func (p *ClientPool) Get() *ethclient.Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	client := p.clients[p.current]
	p.current = (p.current + 1) % len(p.clients)
	return client
}

// Size returns the number of clients in the pool
func (p *ClientPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

// IsClosed returns whether the pool has been closed
func (p *ClientPool) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// Close closes all clients in the pool
func (p *ClientPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	var firstErr error
	for i, client := range p.clients {
		if client != nil {
			client.Close()
			_ = i // Keep for potential error reporting
		}
	}

	return firstErr
}
