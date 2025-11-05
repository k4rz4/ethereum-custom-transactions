package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/k4rz4/ethereum-custom-transactions/pkg/transaction"
)

func TestNewCustomTransaction(t *testing.T) {
	customData := []byte("Hello Blockchain!")

	tx := transaction.NewCustomTransaction(
		big.NewInt(1), 0, addrPtr("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		big.NewInt(0), 21000, big.NewInt(1000000000), big.NewInt(2000000000),
		[]byte{}, customData,
	)

	extracted, err := transaction.GetCustomData(tx)
	if err != nil {
		t.Fatalf("GetCustomData failed: %v", err)
	}

	if !bytes.Equal(extracted, customData) {
		t.Errorf("Mismatch: got %s, want %s", string(extracted), string(customData))
	}

	t.Logf("✅ Custom data verified: %s", string(extracted))
}

func TestEncodeDecodeCustomData(t *testing.T) {
	tests := []struct {
		name         string
		standardData []byte
		customData   []byte
	}{
		{"empty", []byte{}, []byte("test")},
		{"with standard", []byte{0x01, 0x02}, []byte("custom")},
		{"large", []byte{}, make([]byte, 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := transaction.EncodeCustomData(tt.standardData, tt.customData)
			custom, standard, err := transaction.DecodeCustomData(encoded)

			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(custom, tt.customData) {
				t.Error("Custom data mismatch")
			}

			if !bytes.Equal(standard, tt.standardData) {
				t.Error("Standard data mismatch")
			}
		})
	}
}

func addrPtr(hex string) *common.Address {
	addr := common.HexToAddress(hex)
	return &addr
}
