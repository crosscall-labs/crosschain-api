package entrypoint

import "github.com/xssnick/tonutils-go/tvm/cell"

func EntrypointMessageToCell(message EntrypointMessage, queryId uint64) *cell.Cell {
	msgBody := cell.BeginCell().
		MustStoreAddr(message.Destination).
		MustStoreRef(message.Body).
		EndCell()

	return cell.BeginCell().
		MustStoreUInt(1, 32).
		MustStoreUInt(queryId, 64).
		MustStoreRef(msgBody).
		EndCell()
}
