package tvmHandler

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// mint jetton specified to any wallet, backend burn is only possible if proxy account
func JettonMintMessage(to address.Address, query_id uint64, jetton_amount uint64, forward_ton_amount uint64, from address.Address, total_ton_amount uint64) *cell.Cell {
	mintMsg := cell.BeginCell().
		MustStoreUInt(0x178d4519, 32).
		MustStoreUInt(query_id, 64).
		MustStoreCoins(jetton_amount).
		MustStoreAddr(address.NewAddressNone()).
		MustStoreAddr(&from).
		MustStoreCoins(forward_ton_amount).
		MustStoreMaybeRef(cell.BeginCell().EndCell()).
		EndCell()

	return cell.BeginCell().
		MustStoreUInt(0x15, 32).
		MustStoreUInt(query_id, 64).
		MustStoreAddr(&to).
		MustStoreCoins(total_ton_amount).
		MustStoreCoins(jetton_amount).
		MustStoreRef(mintMsg).
		EndCell()
}

// mint jetton specified to any wallet, backend burn is only possible if proxy account
func JettonBurnMessage(query_id uint64, jetton_amount uint64, response_address *address.Address, customPayload *cell.Cell) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0x595f07bc, 32).
		MustStoreUInt(query_id, 64).
		MustStoreCoins(jetton_amount).
		MustStoreAddr(response_address).
		MustStoreMaybeRef(customPayload).
		EndCell()
}

func JettonSendWithdrawTons(query_id uint64) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0x6d8e5e3c, 32).
		MustStoreUInt(query_id, 64).
		EndCell()
}

func JettonSendWithdrawJettons(query_id uint64, from *address.Address, amount uint64) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0x768a50b2, 32).
		MustStoreUInt(query_id, 64).
		MustStoreAddr(from).
		MustStoreCoins(amount).
		MustStoreMaybeRef(cell.BeginCell().EndCell()).
		EndCell()
}
