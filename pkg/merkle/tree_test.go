package merkle_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/merkle"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func TestMerkleTree(t *testing.T) {
	txs := createTestTxs(8)
	tree := merkle.NewTree(txs)

	if tree.LeafCount() != 8 {
		t.Errorf("LeafCount = %d, want 8", tree.LeafCount())
	}

	// Generate proof for tx #3
	proof := tree.GenerateProof(3)
	if len(proof) == 0 {
		t.Fatal("Proof is empty")
	}

	// Verify proof
	leaf := txs[3].Hash()
	if !tree.VerifyProof(leaf, 3, proof) {
		t.Error("Proof verification failed")
	}

	t.Logf("✅ Proof verified: %d hashes for 8 txs", len(proof))
}

func BenchmarkProofGeneration(b *testing.B) {
	txs := createTestTxs(1000)
	tree := merkle.NewTree(txs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.GenerateProof(500)
	}
}

func createTestTxs(count int) types.Transactions {
	txs := make(types.Transactions, count)
	key, _ := crypto.GenerateKey()

	for i := 0; i < count; i++ {
		tx := transaction.NewCustomTransaction(
			big.NewInt(1), uint64(i), addrPtr("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
			big.NewInt(0), 21000, big.NewInt(1e9), big.NewInt(2e9),
			[]byte{}, []byte(fmt.Sprintf("data%d", i)),
		)
		signed, _ := types.SignTx(tx, types.NewLondonSigner(big.NewInt(1)), key)
		txs[i] = signed
	}
	return txs
}

func addrPtr(hex string) *common.Address {
	addr := common.HexToAddress(hex)
	return &addr
}
