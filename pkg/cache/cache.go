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

	cached := val.(*CachedProof)
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

func (pc *ProofCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pc.cache.Range(func(key, value interface{}) bool {
			cached := value.(*CachedProof)
			if time.Since(cached.Timestamp) > pc.ttl {
				pc.cache.Delete(key)
			}
			return true
		})
	}
}

// BlockCache wraps LRU cache for blocks
type BlockCache struct {
	cache *lru.Cache
}

func NewBlockCache(size int) (*BlockCache, error) {
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
	return val.(*types.Block), true
}

func (bc *BlockCache) Set(blockHash common.Hash, block *types.Block) {
	bc.cache.Add(blockHash.Hex(), block)
}

// ReceiptCache wraps LRU cache for receipts
type ReceiptCache struct {
	cache *lru.Cache
}

func NewReceiptCache(size int) (*ReceiptCache, error) {
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
	return val.(*types.Receipt), true
}

func (rc *ReceiptCache) Set(txHash common.Hash, receipt *types.Receipt) {
	rc.cache.Add(txHash.Hex(), receipt)
}
