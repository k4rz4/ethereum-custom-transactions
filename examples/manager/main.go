package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func main() {
	fmt.Println("=== Manager Example with Connection Pool ===\n")

	rpcURL := getEnv("ETH_RPC_URL", "http://localhost:8545")
	privateKey := getEnv(
		"PRIVATE_KEY",
		"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
	)

	fmt.Printf("Connecting to: %s\n", rpcURL)

	// Create manager with pool of 10 connections
	mgr, err := transaction.NewManager(rpcURL, privateKey, 10)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	fmt.Printf("Manager Address: %s\n", mgr.Address().Hex())
	fmt.Printf("Chain ID: %s\n\n", mgr.ChainID().String())

	// Send custom transaction
	fmt.Println("Sending custom transaction...")
	tx, err := mgr.Send(
		common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		nil,
		[]byte("Hello from Manager!"),
		[]byte{},
	)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent: %s\n", tx.Hash().Hex())

	// Wait for mining
	fmt.Println("\nWaiting for transaction to be mined (15 seconds)...")
	time.Sleep(15 * time.Second)

	// Generate proof
	fmt.Println("Generating Merkle proof...")
	proof, err := mgr.GenerateProof(tx.Hash())
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	fmt.Printf("✅ Proof generated!\n")
	fmt.Printf("  Block Number: %s\n", proof.BlockNumber.String())
	fmt.Printf("  Block Hash: %s\n", proof.BlockHash.Hex())
	fmt.Printf("  TX Index: %d\n", proof.TransactionIndex)
	fmt.Printf("  Proof Path Length: %d hashes\n", len(proof.ProofPath))
	fmt.Printf("  Custom Data: %s\n\n", string(proof.CustomData))

	// Verify proof
	fmt.Println("Verifying proof...")
	isValid, err := mgr.VerifyProof(proof)
	if err != nil {
		log.Fatalf("Failed to verify proof: %v", err)
	}

	if isValid {
		fmt.Println("✅ Proof verification: VALID")
	} else {
		fmt.Println("❌ Proof verification: INVALID")
	}

	// Display metrics
	fmt.Println("\n=== Manager Metrics ===")
	metrics := mgr.Metrics()
	fmt.Printf("Transactions Sent: %d\n", metrics["tx_sent"])
	fmt.Printf("Transactions Failed: %d\n", metrics["tx_failed"])
	fmt.Printf("Proofs Generated: %d\n", metrics["proofs_generated"])
	fmt.Printf("Cache Hits: %d\n", metrics["cache_hits"])
	fmt.Printf("Cache Misses: %d\n", metrics["cache_misses"])

	fmt.Println("\n✨ Manager example completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
