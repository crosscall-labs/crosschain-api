package proxyWallet

import "github.com/xssnick/tonutils-go/tvm/cell"

func ExecutionDataToCell(message ExecutionData) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(uint8(message.Regime)), 8).
		MustStoreAddr(message.Destination).
		MustStoreUInt(message.Value, 64).
		MustStoreRef(message.Body).
		EndCell()
}

func SignatureToCell(signature Signature) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(signature.V, 8).
		MustStoreUInt(signature.R, 256).
		MustStoreUInt(signature.S, 256).
		EndCell()
}

func ProxyWalletMessageToCell(message ProxyWalletMessage) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(11, 32).
		MustStoreUInt(message.QueryId, 64).
		MustStoreRef(SignatureToCell(message.Signature)).
		MustStoreRef(ExecutionDataToCell(message.Data)).
		EndCell()
}
