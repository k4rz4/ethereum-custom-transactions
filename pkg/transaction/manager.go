package transaction

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/k4rz4/ethereum-custom-transactions/internal/nonce"
	"github.com/k4rz4/ethereum-custom-transactions/internal/pool"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/cache"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/merkle"
)

// Proof contains transaction proof data
type Proof struct {
	Transaction      *types.Transaction
	BlockNumber      *big.Int
	BlockHash        common.Hash
	TransactionIndex uint
	Receipt          *types.Receipt
	CustomData       []byte
	ProofPath        []common.Hash
}

// Manager handles optimized transaction operations
type Manager struct {
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int

	clientPool   *pool.ClientPool
	nonceManager *nonce.Manager

	proofCache   *cache.ProofCache
	blockCache   *cache.BlockCache
	receiptCache *cache.ReceiptCache
	treeCache    *sync.Map

	metrics *Metrics
}

type Metrics struct {
	TxSent          uint64
	TxFailed        uint64
	ProofsGenerated uint64
	CacheHits       uint64
	CacheMisses     uint64
	mu              sync.RWMutex
}

// NewManager creates optimized transaction manager
func NewManager(rpcURL string, privateKeyHex string, poolSize int) (*Manager, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	clientPool, err := pool.New(rpcURL, poolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	chainID, err := clientPool.Get().ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chainID: %w", err)
	}

	blockCache, _ := cache.NewBlockCache(100)
	receiptCache, _ := cache.NewReceiptCache(1000)

	return &Manager{
		privateKey:   privateKey,
		chainID:      chainID,
		clientPool:   clientPool,
		nonceManager: nonce.New(clientPool.Get()),
		proofCache:   cache.NewProofCache(30 * time.Minute),
		blockCache:   blockCache,
		receiptCache: receiptCache,
		treeCache:    &sync.Map{},
		metrics:      &Metrics{},
	}, nil
}

// Send creates and sends custom transaction
func (m *Manager) Send(to common.Address, value *big.Int, customData, data []byte) (*types.Transaction, error) {
	publicKey := m.privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := m.nonceManager.GetNext(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	client := m.clientPool.Get()
	gasLimit := uint64(100000)
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		m.nonceManager.Reset(fromAddress)
		return nil, err
	}

	head, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		m.nonceManager.Reset(fromAddress)
		return nil, err
	}

	gasFeeCap := new(big.Int).Add(gasTipCap, new(big.Int).Mul(head.BaseFee, big.NewInt(2)))

	tx := NewCustomTransaction(m.chainID, nonce, &to, value, gasLimit, gasTipCap, gasFeeCap, data, customData)

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(m.chainID), m.privateKey)
	if err != nil {
		m.nonceManager.Reset(fromAddress)
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		m.nonceManager.Reset(fromAddress)
		m.metrics.IncrementTxFailed()
		return nil, err
	}

	m.metrics.IncrementTxSent()
	return signedTx, nil
}

// GenerateProof generates optimized proof
func (m *Manager) GenerateProof(txHash common.Hash) (*Proof, error) {
	if cached, exists := m.proofCache.Get(txHash); exists {
		m.metrics.IncrementCacheHits()
		return cached.(*Proof), nil
	}

	m.metrics.IncrementCacheMisses()

	var receipt *types.Receipt
	if cached, ok := m.receiptCache.Get(txHash); ok {
		receipt = cached
	} else {
		var err error
		receipt, err = m.clientPool.Get().TransactionReceipt(context.Background(), txHash)
		if err != nil {
			return nil, err
		}
		m.receiptCache.Set(txHash, receipt)
	}

	tx, isPending, err := m.clientPool.Get().TransactionByHash(context.Background(), txHash)
	if err != nil || isPending {
		return nil, fmt.Errorf("transaction not found or pending")
	}

	tree, err := m.getMerkleTree(receipt.BlockHash)
	if err != nil {
		return nil, err
	}

	proofPath := tree.GenerateProof(receipt.TransactionIndex)
	customData, _ := GetCustomData(tx)

	proof := &Proof{
		Transaction:      tx,
		BlockNumber:      receipt.BlockNumber,
		BlockHash:        receipt.BlockHash,
		TransactionIndex: receipt.TransactionIndex,
		Receipt:          receipt,
		CustomData:       customData,
		ProofPath:        proofPath,
	}

	m.proofCache.Set(txHash, proof)
	m.metrics.IncrementProofsGenerated()

	return proof, nil
}

// VerifyProof verifies transaction proof
func (m *Manager) VerifyProof(proof *Proof) (bool, error) {
	var block *types.Block
	if cached, ok := m.blockCache.Get(proof.BlockHash); ok {
		block = cached
	} else {
		var err error
		block, err = m.clientPool.Get().BlockByHash(context.Background(), proof.BlockHash)
		if err != nil {
			return false, err
		}
		m.blockCache.Set(proof.BlockHash, block)
	}

	if proof.TransactionIndex >= uint(len(block.Transactions())) {
		return false, fmt.Errorf("index out of range")
	}

	tx := block.Transactions()[proof.TransactionIndex]
	if tx.Hash() != proof.Transaction.Hash() {
		return false, fmt.Errorf("hash mismatch")
	}

	tree, err := m.getMerkleTree(proof.BlockHash)
	if err != nil {
		return false, err
	}

	return tree.VerifyProof(proof.Transaction.Hash(), proof.TransactionIndex, proof.ProofPath), nil
}

func (m *Manager) getMerkleTree(blockHash common.Hash) (*merkle.Tree, error) {
	if cached, ok := m.treeCache.Load(blockHash.Hex()); ok {
		return cached.(*merkle.Tree), nil
	}

	var block *types.Block
	if cached, ok := m.blockCache.Get(blockHash); ok {
		block = cached
	} else {
		var err error
		block, err = m.clientPool.Get().BlockByHash(context.Background(), blockHash)
		if err != nil {
			return nil, err
		}
		m.blockCache.Set(blockHash, block)
	}

	tree := merkle.NewTree(block.Transactions())
	m.treeCache.Store(blockHash.Hex(), tree)

	return tree, nil
}

// GetMetrics returns performance metrics
func (m *Manager) GetMetrics() map[string]uint64 {
	return m.metrics.GetStats()
}

func (m *Manager) Close() {
	m.clientPool.Close()
}

func (m *Metrics) IncrementTxSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TxSent++
}

func (m *Metrics) IncrementTxFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TxFailed++
}

func (m *Metrics) IncrementProofsGenerated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ProofsGenerated++
}

func (m *Metrics) IncrementCacheHits() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheHits++
}

func (m *Metrics) IncrementCacheMisses() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheMisses++
}

func (m *Metrics) GetStats() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]uint64{
		"tx_sent":           m.TxSent,
		"tx_failed":         m.TxFailed,
		"proofs_generated":  m.ProofsGenerated,
		"cache_hits":        m.CacheHits,
		"cache_misses":      m.CacheMisses,
	}
}
