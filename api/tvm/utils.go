package tvmHandler

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/laminafinance/crosschain-api/pkg/utils"
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
	//messageHash, _ := hex.DecodeString("a5b70ca6d63a158f2d5ef3965f1fac9eb23498040e0066896cd18c0c5f7e670c")
	ethHash := append(header, hash...)
	unsignedHash := crypto.Keccak256(ethHash)
	fmt.Print(hex.EncodeToString(unsignedHash))

	//h, _ := hex.DecodeString("a5b70ca6d63a158f2d5ef3965f1fac9eb23498040e0066896cd18c0c5f7e670c")
	// pkey, _ := crypto.HexToECDSA("725c99cba93fd5107e92608a1c3aa6d6e2c68caadf86a3cfc6670aae7e47ba07")
	// rs, _ := crypto.Sign(unsignedHash, pkey)
	// fmt.Printf("\n\nsigned via backend: \n%v\n%v", "61aad896921377300b1b735cd47a8939e2cddabf700fc6c7e9fe391660b35474", hex.EncodeToString(rs))

	// testhash, _ := hex.DecodeString("9b5fa1bdd9cd0904847f13a677e0c361595fddf12cf12d89c54f20410c163d7a")
	// testsig, _ := hex.DecodeString("28128b14ea8a48fd625e1622747dc58c80a03da52ea45bf0b1f9044605c3aa7830b3ddf8bf9ed03e5f456d9b9f4a4d170b8099caa38364312b4b08dcf7d12a7e01")
	// fmt.Print("\ngot here")
	recoveredPubKey, err := crypto.SigToPub(unsignedHash, signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}
	// fmt.Print("\ngot here")
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)

	//4a39b40ce89cd1c214c342357b67d41a6085c27a6d1a4e8e077cbf80696af2410b5bc50532c8919d5a02b84a44e7f0a871599c3aebf3c9351ed3fce5b157882000
	//7b63174a7e802e9470e47317ef873d19126187bfafecdd500dfd54abe72323d64bf9659d8aa05b694fadc2a31c4c12daab812b3c7234cc6a91a6ea7cffabd06700
	// backend:
	// 730617d12b5b9a54267e0cf6f4cf385c2f4a247726bef3c29504b0b8e86c21a537ec787db8dc684c0b3ad677c9945a2fb558a36ff594a25de49de63c4e991c3600
	// 730617d12b5b9a54267e0cf6f4cf385c2f4a247726bef3c29504b0b8e86c21a537ec787db8dc684c0b3ad677c9945a2fb558a36ff594a25de49de63c4e991c361b

	// fmt.Print("\nI got this far0")
	// recoveredAddress := crypto.PubkeyToAddress(*pubKey)
	utils.LogInfo("Signature validation results", utils.FormatKeyValueLogs([][2]string{
		{"recovered address", recoveredAddress.String()},
		{" expected address", address.Hex()},
	}))
	return bytes.Equal(recoveredAddress.Bytes(), address.Bytes()), nil
}
