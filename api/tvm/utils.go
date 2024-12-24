package tvmHandler

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func CellToAddress(bouncable bool, testnet bool, workchain uint8, cellData *cell.Cell) *address.Address {
	return address.NewAddress(FlagsToByte(bouncable, testnet), byte(int32(workchain)), cellData.Hash())
}

func FlagsToByte(bouncable bool, testnet bool) (flags byte) {
	// TODO check this magic...
	flags = 0b00010001
	if !bouncable {
		setBit(&flags, 6)
	}
	if testnet {
		setBit(&flags, 7)
	}
	return flags
}

func setBit(n *byte, pos uint) {
	*n |= 1 << pos
}

func AddressToCell(addr *address.Address) (*cell.Cell, error) {
	c := cell.BeginCell().
		MustStoreUInt(uint64(addr.FlagsToByte()), 8).
		MustStoreUInt(uint64(addr.Workchain()), 8)

	if err := c.StoreSlice(addr.Data(), 256); err != nil {
		return nil, err
	}

	return c.EndCell(), nil
}

func ByteArrayToCellDictionary(data []byte) (*cell.Dictionary, error) {
	// Begin parsing the BOC (Bag of Cells)
	c, err := cell.FromBOC(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert byte array to cell dictionary: %v", err)
	}
	return c.AsDict(256), nil
}

func ValidateEvmEcdsaSignature(hash []byte, signature []byte, address common.Address) (bool, error) {
	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length: %d", len(signature))
	}

	header, _ := hex.DecodeString("19457468657265756d205369676e6564204d6573736167653a0a3332")
	ethHash := append(header, hash...)
	unsignedHash := crypto.Keccak256(ethHash)

	recoveredPubKey, err := crypto.SigToPub(unsignedHash, signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)

	fmt.Printf("\nrecovered address: %v", recoveredAddress.Hex())
	fmt.Printf("\nexpected address: %v", address.Hex())
	return bytes.Equal(recoveredAddress.Bytes(), address.Bytes()), nil
}
