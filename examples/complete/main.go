package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func main() {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  Ethereum Custom Transactions - Complete Demo")
	fmt.Println("  Educational/Demonstration Project")
	fmt.Println("═══════════════════════════════════════════════════════\n")

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate key:", err)
	}
	// keyHex := fmt.Sprintf("%x", crypto.FromECDSA(privateKey))
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	fmt.Printf("📍 Generated Address: %s\n\n", address.Hex())

	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Println("IMPORTANT: This example requires a local Ethereum node")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Println("To run this example:")
	fmt.Println("1. Start a local Ethereum node:")
	fmt.Println("   Option A: ganache-cli --port 8545")
	fmt.Println("   Option B: geth --dev --http --http.port 8545")
	fmt.Println("")
	fmt.Println("2. Uncomment the code below in the main() function")
	fmt.Println("3. Run: go run examples/complete/main.go")
	fmt.Println("────────────────────────────────────────────────────────\n")

	// ═══════════════════════════════════════════════════════
	// Example 1: Create a Custom Transaction (Works Offline)
	// ═══════════════════════════════════════════════════════
	fmt.Println("📝 Example 1: Creating a Custom Transaction")
	fmt.Println("─────────────────────────────────────────────")

	customData := []byte("Hello from Ethereum! This is my custom data.")

	tx := transaction.NewCustomTransaction(
		big.NewInt(1337), // Chain ID (1337 for local dev)
		0,                // Nonce
		addrPtr("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		big.NewInt(0),   // Value (0 ETH)
		21000,           // Gas limit
		big.NewInt(1e9), // Gas tip (1 gwei)
		big.NewInt(2e9), // Max fee (2 gwei)
		[]byte{},        // Standard data
		customData,      // Custom data
	)

	// Extract and verify custom data
	extracted, err := transaction.GetCustomData(tx)
	if err != nil {
		log.Fatal("Failed to extract custom data:", err)
	}

	fmt.Printf("✅ Transaction created successfully\n")
	fmt.Printf("   TX Hash:      %s\n", tx.Hash().Hex())
	fmt.Printf("   TX Type:      %d (EIP-1559)\n", tx.Type())
	fmt.Printf("   Custom Data:  %s\n", string(extracted))
	fmt.Printf("   Is Custom TX: %v\n\n", transaction.IsCustomTransaction(tx))

	// ═══════════════════════════════════════════════════════
	// Example 2: Full Workflow (Requires Running Node)
	// ═══════════════════════════════════════════════════════
	fmt.Println("📡 Example 2: Full Workflow with Local Node")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println("⚠️  Commented out - uncomment to test with real node\n")

	// UNCOMMENT THE CODE BELOW TO TEST WITH A RUNNING NODE
	/*
		// Create transaction manager
		fmt.Println("Creating transaction manager...")
		mgr, err := transaction.NewManager(
			"http://localhost:8545",  // RPC endpoint
			keyHex,                    // Private key
			5,                         // Connection pool size
		)
		if err != nil {
			log.Fatal("Failed to create manager:", err)
		}
		defer mgr.Close()

		fmt.Printf("✅ Connected to chain ID: %s\n", mgr.ChainID().String())
		fmt.Printf("   Using address: %s\n\n", mgr.Address().Hex())

		// Send a transaction
		fmt.Println("Sending transaction with custom data...")
		tx, err := mgr.Send(
			common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
			nil, // 0 ETH
			[]byte("Hello, Blockchain! 👋"),
			[]byte{},
		)
		if err != nil {
			log.Fatal("Failed to send transaction:", err)
		}

		fmt.Printf("✅ Transaction sent!\n")
		fmt.Printf("   TX Hash: %s\n", tx.Hash().Hex())
		fmt.Printf("   Waiting for mining...\n\n")

		// Wait for transaction to be mined
		time.Sleep(5 * time.Second)

		// Generate proof
		fmt.Println("Generating Merkle proof...")
		proof, err := mgr.GenerateProof(tx.Hash())
		if err != nil {
			log.Fatal("Failed to generate proof:", err)
		}

		fmt.Printf("✅ Proof generated!\n")
		fmt.Printf("   Block Number:  %s\n", proof.BlockNumber.String())
		fmt.Printf("   Block Hash:    %s\n", proof.BlockHash.Hex())
		fmt.Printf("   TX Index:      %d\n", proof.TransactionIndex)
		fmt.Printf("   Proof Length:  %d hashes (log₂ n)\n", len(proof.ProofPath))
		fmt.Printf("   Custom Data:   %s\n\n", string(proof.CustomData))

		// Verify proof
		fmt.Println("Verifying proof...")
		isValid, err := mgr.VerifyProof(proof)
		if err != nil {
			log.Fatal("Verification failed:", err)
		}

		if isValid {
			fmt.Printf("✅ Proof is VALID!\n")
			fmt.Println("   This proves:")
			fmt.Println("   1. Transaction exists in the block")
			fmt.Println("   2. Custom data was included on-chain")
			fmt.Println("   3. Merkle proof is mathematically sound\n")
		} else {
			fmt.Printf("❌ Proof is INVALID!\n\n")
		}

		// ═══════════════════════════════════════════════════════
		// Example 3: Batch Processing
		// ═══════════════════════════════════════════════════════
		fmt.Println("🚀 Example 3: Batch Processing (High Throughput)")
		fmt.Println("─────────────────────────────────────────────────────")

		processor := batch.NewProcessor(mgr, 10, 500)
		defer processor.Close()

		// Submit 50 transactions
		fmt.Println("Submitting 50 transactions...")
		startTime := time.Now()

		for i := 0; i < 50; i++ {
			req := &batch.Request{
				ID:         fmt.Sprintf("TX-%d", i),
				To:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
				CustomData: []byte(fmt.Sprintf("Batch transaction #%d", i)),
			}

			if err := processor.Submit(req); err != nil {
				log.Printf("Failed to submit TX-%d: %v", i, err)
			}
		}

		// Collect results
		results := processor.GetResults(50, 2*time.Minute)
		duration := time.Since(startTime)

		successCount := 0
		for _, result := range results {
			if result.Error == nil {
				successCount++
			}
		}

		fmt.Printf("\n✅ Batch processing complete!\n")
		fmt.Printf("   Total Time:     %v\n", duration)
		fmt.Printf("   Submitted:      50\n")
		fmt.Printf("   Completed:      %d\n", len(results))
		fmt.Printf("   Success:        %d\n", successCount)
		fmt.Printf("   Throughput:     %.2f TX/sec\n\n", float64(successCount)/duration.Seconds())

		// Show metrics
		metrics := processor.GetMetrics()
		fmt.Println("📊 Processor Metrics:")
		fmt.Printf("   Queued:         %v\n", metrics["queued"])
		fmt.Printf("   Processed:      %v\n", metrics["processed"])
		fmt.Printf("   Failed:         %v\n", metrics["failed"])
		fmt.Printf("   Success Rate:   %.2f%%\n", metrics["success_rate"])
		fmt.Printf("   Avg Duration:   %vms\n\n", metrics["avg_duration"])

		// Manager metrics
		mgrMetrics := mgr.Metrics()
		fmt.Println("📊 Manager Metrics:")
		fmt.Printf("   TX Sent:           %d\n", mgrMetrics["tx_sent"])
		fmt.Printf("   TX Failed:         %d\n", mgrMetrics["tx_failed"])
		fmt.Printf("   Proofs Generated:  %d\n", mgrMetrics["proofs_generated"])
		fmt.Printf("   Cache Hits:        %d\n", mgrMetrics["cache_hits"])
		fmt.Printf("   Cache Misses:      %d\n", mgrMetrics["cache_misses"])
	*/

	// ═══════════════════════════════════════════════════════
	// Summary
	// ═══════════════════════════════════════════════════════
	fmt.Println("\n═══════════════════════════════════════════════════════")
	fmt.Println("  Summary")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("✅ Example 1: Created custom transaction (offline)")
	fmt.Println("⚠️  Example 2: Full workflow (needs local node)")
	fmt.Println("⚠️  Example 3: Batch processing (needs local node)")
	fmt.Println("")
	fmt.Println("To test examples 2 and 3:")
	fmt.Println("1. Start: ganache-cli --port 8545")
	fmt.Println("2. Uncomment the code in main.go")
	fmt.Println("3. Run: go run examples/complete/main.go")
	fmt.Println("═══════════════════════════════════════════════════════\n")
}

func addrPtr(hex string) *common.Address {
	addr := common.HexToAddress(hex)
	return &addr
}
