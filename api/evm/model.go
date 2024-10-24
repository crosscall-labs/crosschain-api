package evmHandler

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

type UnsignedEntryPointRequestResponse struct {
	Header  utils.MessageHeader `query:"header"`
	Payload string              `query:"payload" optional:"true"`
}

type MessageEscrowEvm struct {
	EscrowAddress   string `json:"eaddress"`
	EscrowInit      string `json:"einit"`
	EscrowPayload   string `json:"epayload"`
	EscrowAsset     string `json:"easset"`
	EscrowAmount    string `json:"eamount"`
	EscrowValueType string `json:"evaluetype"`
	EscrowValue     string `json:"evalue"`
}

type MessageOpEvm struct {
	UserOp           PackedUserOperationResponse `json:"op-packed-data"` // parsed data, recommended to validate data
	PaymasterAndData PaymasterAndData            `json:"op-paymaster"`
	UserOpHash       string                      `json:"op-hash"`
	PriceGwei        string                      `json:"op-price"`
}

type PackedUserOperation struct {
	Sender             common.Address
	Nonce              *big.Int
	InitCode           []byte
	CallData           []byte
	AccountGasLimits   [32]byte
	PreVerificationGas *big.Int
	GasFees            [32]byte
	PaymasterAndData   []byte
	Signature          []byte
}

type PackedUserOperationResponse struct {
	Sender             string `json:"op-sender"`
	Nonce              string `json:"op-nonce"`
	InitCode           string `json:"op-init-code"`
	CallData           string `json:"op-call-data"`
	AccountGasLimits   string `json:"op-gas-limits"`
	PreVerificationGas string `json:"op-pre-gas"`
	GasFees            string `json:"op-gas-fees"`
	PaymasterAndData   string `json:"op-paymaster-and-data"`
	Signature          string `json:"op-signature"`
}

type PaymasterAndData struct {
	Paymaster     string `json:"paymaster"`
	Signer        string `json:"signer"`
	EscrowAddress string `json:"escrow-address"`
	TargetDomain  string `json:"target-domain"`
	AssetAddress  string `json:"asset-address"`
	AssetAmount   string `json:"asset-amount"`
	Calldata      string `json:"calldata"`
}

func (m MessageEscrowEvm) GetType() string {
	return "EVM Escrow"
}

func (m MessageOpEvm) GetType() string {
	return "EVM UserOp"
}
