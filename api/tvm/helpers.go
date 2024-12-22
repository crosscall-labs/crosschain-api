package tvmHandler

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func MintMessage(to address.Address, query_id uint64, jetton_amount uint64, forward_ton_amount uint64, from address.Address, total_ton_amount uint64) *cell.Cell {
	mintMsg := cell.BeginCell().
		MustStoreUInt(0x178d4519, 32).
		MustStoreUInt(query_id, 64).
		MustStoreCoins(jetton_amount).
		MustStoreAddr(&address.Address{}).
		MustStoreAddr(&from).
		MustStoreCoins(forward_ton_amount).
		MustStoreMaybeRef(cell.BeginCell().EndCell()).
		EndCell()

	c := cell.BeginCell().
		MustStoreUInt(0x15, 32).
		MustStoreUInt(query_id, 64).
		MustStoreAddr(&to).
		MustStoreCoins(total_ton_amount).
		MustStoreCoins(jetton_amount).
		MustStoreRef(mintMsg).
		EndCell()

	return c
}
