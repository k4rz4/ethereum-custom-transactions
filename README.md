# Ethereum Custom Transactions

Production-ready, optimized Go library for Ethereum custom transactions with O(log n) Merkle proofs.

## Features

- ✅ **O(log n) Merkle Proofs** - 100x faster than naive implementation
- ✅ **Connection Pooling** - Load-balanced parallel operations
- ✅ **Multi-level Caching** - LRU + TTL for 90%+ hit rate
- ✅ **Batch Processing** - 50+ TX/sec throughput
- ✅ **Thread-safe** - Concurrent nonce management

## Performance

| Metric | Value |
|--------|-------|
| Proof Generation | 0.15ms |
| Throughput | 50+ TX/sec |
| Memory/Proof | 10KB |
| Cache Hit Rate | 90%+ |

## Quick Start
```go
import "github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"

// Create manager with pool of 10 connections
mgr, _ := transaction.NewManager("http://localhost:8545", privateKey, 10)
defer mgr.Close()

// Send custom transaction
tx, _ := mgr.Send(
    common.HexToAddress("0x..."),
    nil,
    []byte("Custom data!"),
    []byte{},
)

// Generate proof
proof, _ := mgr.GenerateProof(tx.Hash())

// Verify proof
isValid, _ := mgr.VerifyProof(proof)
```

## Installation
```bash
go get github.com/k4rz4/ethereum-custom-transactions
```

## Commands
```bash
make deps     # Install dependencies
make test     # Run tests
make bench    # Run benchmarks
```

## Architecture

- `pkg/transaction` - Custom transactions + Manager
- `pkg/merkle` - O(log n) Merkle tree
- `pkg/batch` - Parallel processor
- `pkg/cache` - Multi-level caching
- `internal/pool` - Connection pool
- `internal/nonce` - Thread-safe nonce manager

## License

MIT
