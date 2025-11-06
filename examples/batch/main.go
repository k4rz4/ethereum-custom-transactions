package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/k4rz4/ethereum-custom-transactions/pkg/batch"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func main() {
	fmt.Println("=== Batch Processing Example (50+ TX/sec) ===\n")

	rpcURL := getEnv("ETH_RPC_URL", "http://localhost:8545")
	privateKey := getEnv(
		"PRIVATE_KEY",
		"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
	)
	batchSize := getEnvInt("BATCH_SIZE", 100)
	workers := getEnvInt("WORKERS", 20)

	fmt.Printf("RPC URL: %s\n", rpcURL)
	fmt.Printf("Batch Size: %d\n", batchSize)
	fmt.Printf("Workers: %d\n\n", workers)

	// Create manager
	mgr, err := transaction.NewManager(rpcURL, privateKey, 10)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	fmt.Printf("Manager Address: %s\n", mgr.Address().Hex())

	// Create batch processor
	processor := batch.NewProcessor(mgr, workers, 1000)
	defer processor.Close()

	fmt.Printf("\n=== Submitting %d transactions ===\n", batchSize)
	startTime := time.Now()

	// Submit transactions
	for i := 0; i < batchSize; i++ {
		req := &batch.Request{
			ID: fmt.Sprintf("TX-%d", i),
			To: common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
			CustomData: []byte(
				fmt.Sprintf("Batch transaction #%d - timestamp: %d", i, time.Now().Unix()),
			),
		}

		if err := processor.Submit(req); err != nil {
			fmt.Printf("Warning: Failed to submit TX-%d: %v\n", i, err)
		}

		// Print progress every 10 transactions
		if (i+1)%10 == 0 {
			fmt.Printf("Submitted: %d/%d\n", i+1, batchSize)
		}
	}

	submitDuration := time.Since(startTime)
	fmt.Printf("\n✅ All transactions submitted in %v\n", submitDuration)

	// Collect results
	fmt.Printf("\n=== Collecting results (timeout: 3 minutes) ===\n")
	results := processor.GetResults(batchSize, 3*time.Minute)

	totalDuration := time.Since(startTime)

	// Analyze results
	successCount := 0
	failedCount := 0
	var totalTxDuration time.Duration

	for _, result := range results {
		if result.Error == nil {
			successCount++
			totalTxDuration += result.Duration
		} else {
			failedCount++
			fmt.Printf("Failed TX-%s: %v\n", result.Request.ID, result.Error)
		}
	}

	// Calculate metrics
	avgDuration := time.Duration(0)
	if successCount > 0 {
		avgDuration = totalTxDuration / time.Duration(successCount)
	}

	throughput := float64(successCount) / totalDuration.Seconds()

	// Display results
	fmt.Println("\n=== Batch Processing Results ===")
	fmt.Printf("Total Submitted: %d\n", batchSize)
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failedCount)
	fmt.Printf("Success Rate: %.2f%%\n", float64(successCount)/float64(batchSize)*100)
	fmt.Printf("Average TX Duration: %v\n", avgDuration)
	fmt.Printf("Total Duration: %v\n", totalDuration)
	fmt.Printf("Throughput: %.2f TX/sec\n", throughput)

	// Display processor metrics
	fmt.Println("\n=== Processor Metrics ===")
	procMetrics := processor.GetMetrics()
	fmt.Printf("Queued: %v\n", procMetrics["queued"])
	fmt.Printf("Processed: %v\n", procMetrics["processed"])
	fmt.Printf("Failed: %v\n", procMetrics["failed"])
	fmt.Printf("Avg Duration: %v ms\n", procMetrics["avg_duration"])
	fmt.Printf("Success Rate: %.2f%%\n", procMetrics["success_rate"])
	fmt.Printf("Workers: %v\n", procMetrics["workers"])

	// Display manager metrics
	fmt.Println("\n=== Manager Metrics ===")
	mgrMetrics := mgr.Metrics()
	fmt.Printf("TX Sent: %d\n", mgrMetrics["tx_sent"])
	fmt.Printf("TX Failed: %d\n", mgrMetrics["tx_failed"])
	fmt.Printf("Cache Hits: %d\n", mgrMetrics["cache_hits"])

	// Sample proof verification (on first successful transaction)
	if successCount > 0 {
		fmt.Println("\n=== Sample Proof Verification ===")
		fmt.Println("Waiting for first transaction to be mined (15 seconds)...")
		time.Sleep(15 * time.Second)

		// Find first successful result
		var sampleTx *batch.Result
		for _, result := range results {
			if result.Error == nil && result.Transaction != nil {
				sampleTx = result
				break
			}
		}

		if sampleTx != nil {
			fmt.Printf("Generating proof for TX: %s\n", sampleTx.Transaction.Hash().Hex())
			proof, err := mgr.GenerateProof(sampleTx.Transaction.Hash())
			if err != nil {
				fmt.Printf("Warning: Failed to generate proof: %v\n", err)
			} else {
				fmt.Printf("✅ Proof generated (Block #%s, %d hashes)\n",
					proof.BlockNumber.String(), len(proof.ProofPath))

				isValid, err := mgr.VerifyProof(proof)
				if err != nil {
					fmt.Printf("Warning: Failed to verify proof: %v\n", err)
				} else if isValid {
					fmt.Println("✅ Proof verification: VALID")
				} else {
					fmt.Println("❌ Proof verification: INVALID")
				}
			}
		}
	}

	fmt.Println("\n✨ Batch processing example completed!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
