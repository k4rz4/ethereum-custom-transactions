package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func main() {
	fmt.Println("=== Optimized Ethereum Custom Transactions ===\n")

	// Generate test key
	privateKey, _ := crypto.GenerateKey()
	keyHex := fmt.Sprintf("%x", crypto.FromECDSA(privateKey))

	fmt.Println("NOTE: This example requires a local Ethereum node")
	fmt.Println("Run: ganache-cli or geth --dev --http\n")

	// Example 1: Basic usage
	fmt.Println("Example 1: Create Transaction")
	tx := transaction.NewCustomTransaction(
		big.NewInt(1337), 0, addrPtr("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		big.NewInt(0), 21000, big.NewInt(1e9), big.NewInt(2e9),
		[]byte{}, []byte("Hello Blockchain!"),
	)

	customData, _ := transaction.GetCustomData(tx)
	fmt.Printf("  Custom Data: %s\n", string(customData))
	fmt.Printf("  TX Hash: %s\n", tx.Hash().Hex())
	fmt.Printf("  Is Custom: %v\n\n", transaction.IsCustomTransaction(tx))

	// Example 2: Manager (requires running node)
	fmt.Println("Example 2: Transaction Manager")
	fmt.Println("  To test with real node, uncomment the code below\n")

	// Uncomment to test with real node:
	/*
		mgr, err := transaction.NewManager(
			"http://localhost:8545",
			keyHex,
			10, // pool size
		)
		if err != nil {
			log.Fatal(err)
		}
		defer mgr.Close()

		// Send transaction
		tx, err := mgr.Send(
			common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
			nil,
			[]byte("Custom data!"),
			[]byte{},
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("  Sent TX: %s\n", tx.Hash().Hex())

		// Wait for mining
		time.Sleep(15 * time.Second)

		// Generate proof
		proof, err := mgr.GenerateProof(tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("  Block: %s\n", proof.BlockNumber.String())
		fmt.Printf("  Proof length: %d\n", len(proof.ProofPath))

		// Verify proof
		isValid, _ := mgr.VerifyProof(proof)
		fmt.Printf("  Proof valid: %v\n\n", isValid)

		// Example 3: Batch processing
		fmt.Println("Example 3: Batch Processing (50+ TX/sec)")

		processor := batch.NewProcessor(mgr, 20, 1000)
		defer processor.Close()

		// Submit 100 transactions
		for i := 0; i < 100; i++ {
			req := &batch.Request{
				ID:         fmt.Sprintf("TX-%d", i),
				To:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
				CustomData: []byte(fmt.Sprintf("Batch TX %d", i)),
			}
			processor.Submit(req)
		}

		// Get results
		results := processor.GetResults(100, 2*time.Minute)

		successCount := 0
		for _, result := range results {
			if result.Error == nil {
				successCount++
			}
		}

		fmt.Printf("  Submitted: 100\n")
		fmt.Printf("  Success: %d\n", successCount)

		metrics := processor.GetMetrics()
		fmt.Printf("  Success Rate: %.2f%%\n", metrics["success_rate"])
		fmt.Printf("  Avg Duration: %vms\n\n", metrics["avg_duration"])

		// Manager metrics
		mgrMetrics := mgr.GetMetrics()
		fmt.Println("Manager Metrics:")
		fmt.Printf("  TX Sent: %d\n", mgrMetrics["tx_sent"])
		fmt.Printf("  Cache Hits: %d\n", mgrMetrics["cache_hits"])
		fmt.Printf("  Proofs Generated: %d\n", mgrMetrics["proofs_generated"])
	*/

	fmt.Println("✨ All examples shown!")
	fmt.Println("\nTo test with real Ethereum node:")
	fmt.Println("1. Start ganache-cli or geth --dev --http")
	fmt.Println("2. Uncomment the code in examples/complete/main.go")
	fmt.Println("3. Run: go run examples/complete/main.go")
}

func addrPtr(hex string) *common.Address {
	addr := common.HexToAddress(hex)
	return &addr
}
