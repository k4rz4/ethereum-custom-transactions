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

const (
	DefaultGasLimit   = uint64(100_000)
	BaseFeeMultiplier = 2
	DefaultTimeout    = 30 * time.Second
)

type Proof struct {
	Transaction      *types.Transaction
	BlockNumber      *big.Int
	BlockHash        common.Hash
	TransactionIndex uint
	Receipt          *types.Receipt
	CustomData       []byte
	ProofPath        []common.Hash
}

type Manager struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    *big.Int

	clientPool   *pool.ClientPool
	nonceManager *nonce.Manager

	proofCache   *cache.ProofCache
	blockCache   *cache.BlockCache
	receiptCache *cache.ReceiptCache
	treeCache    *sync.Map // stores common.Hash -> *merkle.Tree

	metrics *Metrics
	mu      sync.RWMutex
}

// Metrics tracks manager performance
type Metrics struct {
	TxSent          uint64
	TxFailed        uint64
	ProofsGenerated uint64
	CacheHits       uint64
	CacheMisses     uint64
	mu              sync.RWMutex
}

// NewManager creates an optimized transaction manager
// rpcURL: Ethereum node RPC endpoint (e.g., "http://localhost:8545")
// privateKeyHex: Private key in hex format (without 0x prefix)
// poolSize: Number of client connections to pool (recommended: 5-10)
func NewManager(rpcURL string, privateKeyHex string, poolSize int) (*Manager, error) {
	if poolSize < 1 {
		poolSize = 5
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	clientPool, err := pool.New(rpcURL, poolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create client pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := clientPool.Get().ChainID(ctx)
	if err != nil {
		clientPool.Close()
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	blockCache, err := cache.NewBlockCache(100)
	if err != nil {
		clientPool.Close()
		return nil, fmt.Errorf("failed to create block cache: %w", err)
	}

	receiptCache, err := cache.NewReceiptCache(1000)
	if err != nil {
		clientPool.Close()
		return nil, fmt.Errorf("failed to create receipt cache: %w", err)
	}

	return &Manager{
		privateKey:   privateKey,
		address:      address,
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

func (m *Manager) Send(
	to common.Address,
	value *big.Int,
	customData, data []byte,
) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	return m.SendWithContext(ctx, to, value, customData, data)
}

func (m *Manager) SendWithContext(
	ctx context.Context,
	to common.Address,
	value *big.Int,
	customData, data []byte,
) (*types.Transaction, error) {
	if value == nil {
		value = big.NewInt(0)
	}

	nonce, err := m.nonceManager.GetNext(m.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	client := m.clientPool.Get()

	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		m.nonceManager.Reset(m.address)
		return nil, fmt.Errorf("failed to get gas tip: %w", err)
	}

	head, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		m.nonceManager.Reset(m.address)
		return nil, fmt.Errorf("failed to get block header: %w", err)
	}

	if head.BaseFee == nil {
		m.nonceManager.Reset(m.address)
		return nil, fmt.Errorf("base fee is nil, chain may not support EIP-1559")
	}

	gasFeeCap := new(big.Int).Add(
		gasTipCap,
		new(big.Int).Mul(head.BaseFee, big.NewInt(BaseFeeMultiplier)),
	)

	// Create custom transaction
	tx := NewCustomTransaction(
		m.chainID,
		nonce,
		&to,
		value,
		DefaultGasLimit,
		gasTipCap,
		gasFeeCap,
		data,
		customData,
	)

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(m.chainID), m.privateKey)
	if err != nil {
		m.nonceManager.Reset(m.address)
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		m.nonceManager.Reset(m.address)
		m.metrics.IncrementTxFailed()
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	m.metrics.IncrementTxSent()
	return signedTx, nil
}

func (m *Manager) GenerateProof(txHash common.Hash) (*Proof, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	return m.GenerateProofWithContext(ctx, txHash)
}

func (m *Manager) GenerateProofWithContext(
	ctx context.Context,
	txHash common.Hash,
) (*Proof, error) {
	// Check cache first
	if cached, exists := m.proofCache.Get(txHash); exists {
		m.metrics.IncrementCacheHits()
		proof, ok := cached.(*Proof)
		if !ok {
			return nil, fmt.Errorf("invalid cached proof type")
		}
		return proof, nil
	}

	m.metrics.IncrementCacheMisses()

	// Get receipt
	receipt, err := m.getReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	// Get transaction
	tx, isPending, err := m.clientPool.Get().TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if isPending {
		return nil, fmt.Errorf("transaction is still pending")
	}

	// Get Merkle tree for this block
	tree, err := m.getMerkleTree(ctx, receipt.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get merkle tree: %w", err)
	}

	// Generate proof path
	proofPath := tree.GenerateProof(receipt.TransactionIndex)
	if proofPath == nil {
		return nil, fmt.Errorf("failed to generate proof path")
	}

	// Extract custom data
	customData, err := GetCustomData(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract custom data: %w", err)
	}

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

func (m *Manager) VerifyProof(proof *Proof) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	return m.VerifyProofWithContext(ctx, proof)
}

func (m *Manager) VerifyProofWithContext(ctx context.Context, proof *Proof) (bool, error) {
	if proof == nil {
		return false, fmt.Errorf("proof is nil")
	}

	block, err := m.getBlock(ctx, proof.BlockHash)
	if err != nil {
		return false, fmt.Errorf("failed to get block: %w", err)
	}

	if proof.TransactionIndex >= uint(len(block.Transactions())) {
		return false, fmt.Errorf("transaction index %d out of range (block has %d transactions)",
			proof.TransactionIndex, len(block.Transactions()))
	}

	tx := block.Transactions()[proof.TransactionIndex]
	if tx.Hash() != proof.Transaction.Hash() {
		return false, fmt.Errorf("transaction hash mismatch")
	}

	if proof.Receipt.TxHash != proof.Transaction.Hash() {
		return false, fmt.Errorf("receipt transaction hash mismatch")
	}

	tree, err := m.getMerkleTree(ctx, proof.BlockHash)
	if err != nil {
		return false, fmt.Errorf("failed to get merkle tree: %w", err)
	}

	isValid := tree.VerifyProof(proof.Transaction.Hash(), proof.TransactionIndex, proof.ProofPath)
	if !isValid {
		return false, fmt.Errorf("merkle proof verification failed")
	}

	extractedData, err := GetCustomData(proof.Transaction)
	if err != nil {
		return false, fmt.Errorf("failed to extract custom data: %w", err)
	}

	if len(extractedData) != len(proof.CustomData) {
		return false, fmt.Errorf("custom data length mismatch")
	}

	for i := range extractedData {
		if extractedData[i] != proof.CustomData[i] {
			return false, fmt.Errorf("custom data mismatch at byte %d", i)
		}
	}

	return true, nil
}

func (m *Manager) Address() common.Address {
	return m.address
}

func (m *Manager) ChainID() *big.Int {
	return new(big.Int).Set(m.chainID)
}

func (m *Manager) Metrics() map[string]uint64 {
	return m.metrics.GetStats()
}

func (m *Manager) Close() error {
	return m.clientPool.Close()
}

func (m *Manager) getReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	// Check cache
	if cached, ok := m.receiptCache.Get(txHash); ok {
		return cached, nil
	}

	receipt, err := m.clientPool.Get().TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	m.receiptCache.Set(txHash, receipt)
	return receipt, nil
}

func (m *Manager) getBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	if cached, ok := m.blockCache.Get(blockHash); ok {
		return cached, nil
	}

	block, err := m.clientPool.Get().BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	m.blockCache.Set(blockHash, block)
	return block, nil
}

func (m *Manager) getMerkleTree(ctx context.Context, blockHash common.Hash) (*merkle.Tree, error) {
	if cached, ok := m.treeCache.Load(blockHash); ok {
		tree, ok := cached.(*merkle.Tree)
		if !ok {
			return nil, fmt.Errorf("invalid cached tree type")
		}
		return tree, nil
	}

	block, err := m.getBlock(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	tree := merkle.NewTree(block.Transactions())
	m.treeCache.Store(blockHash, tree)

	return tree, nil
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
		"tx_sent":          m.TxSent,
		"tx_failed":        m.TxFailed,
		"proofs_generated": m.ProofsGenerated,
		"cache_hits":       m.CacheHits,
		"cache_misses":     m.CacheMisses,
	}
}
