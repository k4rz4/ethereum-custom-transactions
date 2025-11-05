package transaction

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MagicBytes is a unique identifier for custom transactions
var MagicBytes = []byte{0xCA, 0xFE, 0xDA, 0x7A}

type CustomTransaction struct {
	*types.Transaction
	customData []byte
}

func NewCustomTransaction(
	chainID *big.Int,
	nonce uint64,
	to *common.Address,
	value *big.Int,
	gasLimit uint64,
	gasTipCap *big.Int,
	gasFeeCap *big.Int,
	data []byte,
	customData []byte,
) *types.Transaction {
	encodedData := EncodeCustomData(data, customData)

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Gas:        gasLimit,
		To:         to,
		Value:      value,
		Data:       encodedData,
		AccessList: types.AccessList{},
	})
}

func EncodeCustomData(standardData, customData []byte) []byte {
	totalSize := len(MagicBytes) + 4 + len(customData) + len(standardData)
	result := make([]byte, 0, totalSize)

	result = append(result, MagicBytes...)

	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(customData)))
	result = append(result, lengthBytes...)

	result = append(result, customData...)

	result = append(result, standardData...)

	return result
}

func DecodeCustomData(encodedData []byte) (customData, standardData []byte, err error) {
	minLength := len(MagicBytes) + 4
	if len(encodedData) < minLength {
		return nil, encodedData, nil
	}

	if !bytes.Equal(encodedData[:len(MagicBytes)], MagicBytes) {
		return nil, encodedData, nil
	}

	length := binary.BigEndian.Uint32(encodedData[len(MagicBytes) : len(MagicBytes)+4])
	offset := len(MagicBytes) + 4

	if uint32(len(encodedData)) < uint32(offset)+length {
		return nil, nil, fmt.Errorf(
			"invalid custom data encoding: declared length %d exceeds available data",
			length,
		)
	}

	customData = encodedData[offset : offset+int(length)]

	standardData = encodedData[offset+int(length):]

	return customData, standardData, nil
}

func GetCustomData(tx *types.Transaction) ([]byte, error) {
	customData, _, err := DecodeCustomData(tx.Data())
	return customData, err
}

func IsCustomTransaction(tx *types.Transaction) bool {
	data := tx.Data()
	if len(data) < len(MagicBytes) {
		return false
	}
	return bytes.Equal(data[:len(MagicBytes)], MagicBytes)
}
