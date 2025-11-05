package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func main() {
	fmt.Println("=== Ethereum Custom Transaction ===\n")

	privateKey, _ := crypto.GenerateKey()
	publicKey := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("Address: %s\n\n", publicKey.Hex())

	customData := []byte("Hello from custom transaction!")

	tx := transaction.NewCustomTransaction(
		big.NewInt(1), 0, addrPtr("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		big.NewInt(0), 21000, big.NewInt(1000000000), big.NewInt(2000000000),
		[]byte{}, customData,
	)

	fmt.Println("✅ Transaction Created:")
	fmt.Printf("  Type: 0x%x\n", tx.Type())
	fmt.Printf("  Nonce: %d\n", tx.Nonce())
	fmt.Printf("  Gas: %d\n", tx.Gas())

	extracted, _ := transaction.GetCustomData(tx)
	fmt.Printf("\n✅ Custom Data: %s\n", string(extracted))
	fmt.Printf("  Is Custom TX: %v\n", transaction.IsCustomTransaction(tx))
}

func addrPtr(hex string) *common.Address {
	addr := common.HexToAddress(hex)
	return &addr
}
