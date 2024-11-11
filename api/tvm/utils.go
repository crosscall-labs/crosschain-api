package tvmHandler

import (
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func CellToAddress(cellData *cell.Cell) (*address.Address, error) {
	c := cellData.BeginParse()

	flags, err := c.LoadUInt(8)
	if err != nil {
		return nil, fmt.Errorf("failed to load flags: %v", err)
	}

	workchain, err := c.LoadUInt(8)
	if err != nil {
		return nil, fmt.Errorf("failed to load workchain ID: %v", err)
	}

	addrData, err := c.LoadSlice(256)
	if err != nil {
		return nil, fmt.Errorf("failed to load address data: %v", err)
	}

	return address.NewAddress(byte(flags), byte(workchain), addrData), nil
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
