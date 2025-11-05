package batch

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

// Processor handles high-throughput parallel processing
type Processor struct {
	manager *transaction.Manager
	workers int
	queue   chan *Request
	results chan *Result
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	metrics *Metrics
}

type Request struct {
	ID         string
	To         common.Address
	Value      *big.Int
	CustomData []byte
	Data       []byte
	Timestamp  time.Time
}

type Result struct {
	Request     *Request
	Transaction *types.Transaction
	Error       error
	Duration    time.Duration
}

type Metrics struct {
	TotalQueued    uint64
	TotalProcessed uint64
	TotalFailed    uint64
	AvgDuration    time.Duration
	mu             sync.RWMutex
}

func NewProcessor(manager *transaction.Manager, workers int, queueSize int) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Processor{
		manager: manager,
		workers: workers,
		queue:   make(chan *Request, queueSize),
		results: make(chan *Result, queueSize),
		ctx:     ctx,
		cancel:  cancel,
		metrics: &Metrics{},
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	return p
}

func (p *Processor) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case req := <-p.queue:
			if req == nil {
				return
			}
			p.processRequest(req)
		}
	}
}

func (p *Processor) processRequest(req *Request) {
	startTime := time.Now()

	tx, err := p.manager.Send(req.To, req.Value, req.CustomData, req.Data)

	duration := time.Since(startTime)

	result := &Result{
		Request:     req,
		Transaction: tx,
		Error:       err,
		Duration:    duration,
	}

	p.metrics.Update(result)

	select {
	case p.results <- result:
	case <-p.ctx.Done():
	}
}

func (p *Processor) Submit(req *Request) error {
	req.Timestamp = time.Now()
	p.metrics.IncrementQueued()

	select {
	case p.queue <- req:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("processor shut down")
	default:
		return fmt.Errorf("queue full")
	}
}

func (p *Processor) GetResult() *Result {
	select {
	case result := <-p.results:
		return result
	case <-p.ctx.Done():
		return nil
	}
}

func (p *Processor) GetResults(count int, timeout time.Duration) []*Result {
	results := make([]*Result, 0, count)
	deadline := time.After(timeout)

	for i := 0; i < count; i++ {
		select {
		case result := <-p.results:
			results = append(results, result)
		case <-deadline:
			return results
		case <-p.ctx.Done():
			return results
		}
	}

	return results
}

// Close shuts down processor
func (p *Processor) Close() {
	p.cancel()
	close(p.queue)
	p.wg.Wait()
	close(p.results)
}

func (p *Processor) GetMetrics() map[string]interface{} {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return map[string]interface{}{
		"queued":       p.metrics.TotalQueued,
		"processed":    p.metrics.TotalProcessed,
		"failed":       p.metrics.TotalFailed,
		"avg_duration": p.metrics.AvgDuration.Milliseconds(),
		"success_rate": p.calculateSuccessRate(),
	}
}

func (p *Processor) calculateSuccessRate() float64 {
	if p.metrics.TotalProcessed == 0 {
		return 0
	}
	success := p.metrics.TotalProcessed - p.metrics.TotalFailed
	return float64(success) / float64(p.metrics.TotalProcessed) * 100
}

func (m *Metrics) IncrementQueued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalQueued++
}

func (m *Metrics) Update(result *Result) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalProcessed++
	if result.Error != nil {
		m.TotalFailed++
	}

	if m.TotalProcessed == 1 {
		m.AvgDuration = result.Duration
	} else {
		total := m.AvgDuration*time.Duration(m.TotalProcessed-1) + result.Duration
		m.AvgDuration = total / time.Duration(m.TotalProcessed)
	}
}
