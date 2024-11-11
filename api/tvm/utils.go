package tvmHandler

import (
	"fmt"

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
