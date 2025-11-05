package merkle

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Tree struct {
	root   common.Hash
	leaves []common.Hash
	layers [][]common.Hash
	mu     sync.RWMutex
}

func NewTree(txs types.Transactions) *Tree {
	leaves := make([]common.Hash, len(txs))
	for i, tx := range txs {
		leaves[i] = tx.Hash()
	}

	tree := &Tree{
		leaves: leaves,
		layers: make([][]common.Hash, 0),
	}

	tree.build()
	return tree
}

// build constructs all layers once (O(n))
func (t *Tree) build() {
	t.mu.Lock()
	defer t.mu.Unlock()

	currentLevel := make([]common.Hash, len(t.leaves))
	copy(currentLevel, t.leaves)
	t.layers = append(t.layers, currentLevel)

	// Build tree bottom-up, cache each layer
	for len(currentLevel) > 1 {
		nextLevel := make([]common.Hash, 0, (len(currentLevel)+1)/2)

		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				combined := append(currentLevel[i].Bytes(), currentLevel[i+1].Bytes()...)
				parentHash := crypto.Keccak256Hash(combined)
				nextLevel = append(nextLevel, parentHash)
			} else {
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}

		t.layers = append(t.layers, nextLevel)
		currentLevel = nextLevel
	}

	t.root = currentLevel[0]
}

func (t *Tree) GenerateProof(index uint) []common.Hash {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if index >= uint(len(t.leaves)) {
		return nil
	}

	proof := make([]common.Hash, 0, len(t.layers)-1)
	currentIndex := index

	for level := 0; level < len(t.layers)-1; level++ {
		layer := t.layers[level]
		siblingIndex := currentIndex ^ 1

		if siblingIndex < uint(len(layer)) {
			proof = append(proof, layer[siblingIndex])
		}

		currentIndex >>= 1
	}

	return proof
}

func (t *Tree) VerifyProof(leaf common.Hash, index uint, proof []common.Hash) bool {
	t.mu.RLock()
	root := t.root
	t.mu.RUnlock()

	currentHash := leaf
	currentIndex := index

	for _, siblingHash := range proof {
		if currentIndex%2 == 0 {
			combined := append(currentHash.Bytes(), siblingHash.Bytes()...)
			currentHash = crypto.Keccak256Hash(combined)
		} else {
			combined := append(siblingHash.Bytes(), currentHash.Bytes()...)
			currentHash = crypto.Keccak256Hash(combined)
		}
		currentIndex >>= 1
	}

	return currentHash == root
}

func (t *Tree) Root() common.Hash {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

func (t *Tree) LeafCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.leaves)
}

func (t *Tree) Depth() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.layers)
}
