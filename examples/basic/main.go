package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func main() {
	fmt.Println("=== Ethereum Custom Transaction (Basic Example) ===\n")

	// Generate a test key
	privateKey, _ := crypto.GenerateKey()
	publicKey := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("Address: %s\n\n", publicKey.Hex())

	// Custom data to embed in transaction
	customData := []byte("Hello from custom transaction!")

	// Create custom transaction (EIP-1559 / Type 2)
	tx := transaction.NewCustomTransaction(
		big.NewInt(1), // Chain ID
		0,             // Nonce
		addrPtr("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"), // To address
		big.NewInt(0),          // Value (0 ETH)
		21000,                  // Gas limit
		big.NewInt(1000000000), // Gas tip cap (1 gwei)
		big.NewInt(2000000000), // Gas fee cap (2 gwei)
		[]byte{},               // Standard data
		customData,             // Custom data
	)

	fmt.Println("✅ Transaction Created:")
	fmt.Printf("  Type: 0x%x\n", tx.Type())
	fmt.Printf("  Nonce: %d\n", tx.Nonce())
	fmt.Printf("  Gas: %d\n", tx.Gas())
	fmt.Printf("  To: %s\n", tx.To().Hex())

	// Extract and verify custom data
	extracted, _ := transaction.GetCustomData(tx)
	fmt.Printf("\n✅ Custom Data: %s\n", string(extracted))
	fmt.Printf("  Is Custom TX: %v\n", transaction.IsCustomTransaction(tx))
	fmt.Printf("  TX Hash: %s\n", tx.Hash().Hex())
}

func addrPtr(hex string) *common.Address {
	addr := common.HexToAddress(hex)
	return &addr
}
