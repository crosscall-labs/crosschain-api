package proxyWallet

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type GetWalletInfoResponse struct {
	Nonce             string `json:"nonce"`
	EntrypointAddress string `json:"entrypoint-address"`
	EvmAddress        string `json:"evm-address"`
	TvmAddress        string `json:"tvm-address"`
}

type ExecutionData struct {
	Regime      byte
	Destination *address.Address
	Value       uint64
	Body        *cell.Cell
}

type Signature struct {
	V uint64
	R uint64
	S uint64
}

type ProxyWalletMessage struct {
	QueryId   uint64
	Signature Signature
	Data      ExecutionData
}
