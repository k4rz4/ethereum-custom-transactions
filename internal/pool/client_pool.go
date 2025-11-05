package pool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ClientPool struct {
	clients []*ethclient.Client
	mu      sync.RWMutex
	current int
}

func New(rpcURL string, size int) (*ClientPool, error) {
	if size < 1 {
		size = 1
	}

	clients := make([]*ethclient.Client, size)
	for i := 0; i < size; i++ {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create client %d: %w", i, err)
		}
		clients[i] = client
	}

	return &ClientPool{clients: clients}, nil
}

func (p *ClientPool) Get() *ethclient.Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	client := p.clients[p.current]
	p.current = (p.current + 1) % len(p.clients)
	return client
}

func (p *ClientPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

func (p *ClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.Close()
	}
}
