package cache

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
)

// ProofCache stores proofs with TTL
type ProofCache struct {
	cache *sync.Map
	ttl   time.Duration
}

type CachedProof struct {
	Proof     interface{}
	Timestamp time.Time
}

func NewProofCache(ttl time.Duration) *ProofCache {
	pc := &ProofCache{
		cache: &sync.Map{},
		ttl:   ttl,
	}
	go pc.cleanup()
	return pc
}

func (pc *ProofCache) Get(txHash common.Hash) (interface{}, bool) {
	val, exists := pc.cache.Load(txHash.Hex())
	if !exists {
		return nil, false
	}

	cached, ok := val.(*CachedProof)
	if !ok {
		// Invalid type, remove it
		pc.cache.Delete(txHash.Hex())
		return nil, false
	}

	// Check if expired
	if time.Since(cached.Timestamp) > pc.ttl {
		pc.cache.Delete(txHash.Hex())
		return nil, false
	}

	return cached.Proof, true
}

func (pc *ProofCache) Set(txHash common.Hash, proof interface{}) {
	pc.cache.Store(txHash.Hex(), &CachedProof{
		Proof:     proof,
		Timestamp: time.Now(),
	})
}

func (pc *ProofCache) Delete(txHash common.Hash) {
	pc.cache.Delete(txHash.Hex())
}

func (pc *ProofCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pc.cache.Range(func(key, value interface{}) bool {
			cached, ok := value.(*CachedProof)
			if !ok {
				pc.cache.Delete(key)
				return true
			}

			if time.Since(cached.Timestamp) > pc.ttl {
				pc.cache.Delete(key)
			}
			return true
		})
	}
}

type BlockCache struct {
	cache *lru.Cache
}

func NewBlockCache(size int) (*BlockCache, error) {
	if size < 1 {
		size = 100
	}

	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &BlockCache{cache: cache}, nil
}

func (bc *BlockCache) Get(blockHash common.Hash) (*types.Block, bool) {
	val, ok := bc.cache.Get(blockHash.Hex())
	if !ok {
		return nil, false
	}

	block, ok := val.(*types.Block)
	if !ok {
		// Invalid type, remove it
		bc.cache.Remove(blockHash.Hex())
		return nil, false
	}

	return block, true
}

func (bc *BlockCache) Set(blockHash common.Hash, block *types.Block) {
	bc.cache.Add(blockHash.Hex(), block)
}

func (bc *BlockCache) Delete(blockHash common.Hash) {
	bc.cache.Remove(blockHash.Hex())
}

func (bc *BlockCache) Len() int {
	return bc.cache.Len()
}

type ReceiptCache struct {
	cache *lru.Cache
}

// NewReceiptCache creates a new receipt cache
// size: Maximum number of receipts to cache
func NewReceiptCache(size int) (*ReceiptCache, error) {
	if size < 1 {
		size = 1000
	}

	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &ReceiptCache{cache: cache}, nil
}

func (rc *ReceiptCache) Get(txHash common.Hash) (*types.Receipt, bool) {
	val, ok := rc.cache.Get(txHash.Hex())
	if !ok {
		return nil, false
	}

	receipt, ok := val.(*types.Receipt)
	if !ok {
		// Invalid type, remove it
		rc.cache.Remove(txHash.Hex())
		return nil, false
	}

	return receipt, true
}

func (rc *ReceiptCache) Set(txHash common.Hash, receipt *types.Receipt) {
	rc.cache.Add(txHash.Hex(), receipt)
}

func (rc *ReceiptCache) Delete(txHash common.Hash) {
	rc.cache.Remove(txHash.Hex())
}

func (rc *ReceiptCache) Len() int {
	return rc.cache.Len()
}
