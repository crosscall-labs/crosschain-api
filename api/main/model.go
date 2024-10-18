package handler

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Chain struct {
	ChainId                      string
	Domain                       uint32
	AddressEntrypoint            string
	AddressEntrypointSimulations string
	AddressSimpleAccountFactory  string
	AddressSimpleAccount         string
	AddressMulticall             string
	AddressHyperlaneMailbox      string
	AddressHyperlaneIgp          string
	AddressPaymaster             string
	AddressEscrow                string
	AddressEscrowFactory         string
}

type Call struct {
	Target   common.Address
	CallData []byte
}

type Call3 struct {
	Target   common.Address
	Value    *big.Int
	CallData []byte
}

type Test struct {
	Success    bool
	ReturnData string
}

type TestBytecode struct {
	Payload string `json:"payload"`
}

type Result struct {
	Success    bool
	ReturnData []byte
}

type Compare struct {
	Correct string
	Test    string
}

// Block data structure
type Block struct {
	BlockNumber       int64         `json:"blockNumber"`
	Timestamp         uint64        `json:"timestamp"`
	Difficulty        uint64        `json:"difficulty"`
	Hash              string        `json:"hash"`
	TransactionsCount int           `json:"transactionsCount"`
	Transactions      []Transaction `json:"transactions"`
}

// Transaction data structure
type Transaction struct {
	Hash     string `json:"hash"`
	Value    string `json:"value"`
	Gas      uint64 `json:"gas"`
	GasPrice uint64 `json:"gasPrice"`
	Nonce    uint64 `json:"nonce"`
	To       string `json:"to"`
	Pending  bool   `json:"pending"`
}

// TransferEthRequest data structure
type TransferEthRequest struct {
	PrivKey string `json:"privKey"`
	To      string `json:"to"`
	Amount  int64  `json:"amount"`
}

type UserOperation struct {
	Sender                string `json:"sender"`
	Nonce                 string `json:"nonce"`
	InitCode              string `json:"init-code"`
	CallData              string `json:"call-data"`
	CallGasLimit          string `json:"call-gas-limit"`
	VerificationGasLimit  string `json:"verification-gas-limit"`
	PreVerificationGas    string `json:"pre-verification-gas"`
	MaxFeePerGas          string `json:"max-fee-per-gas"`
	MaxPritorityFeePerGas string `json:"max-priority-fee-per-gas"`
	PaymasterAndData      string `json:"paymaster-and-data"`
	Signature             string `json:"signature"`
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
	Sender             string `json:"sender"`
	Nonce              string `json:"nonce"`
	InitCode           string `json:"init-code"`
	CallData           string `json:"call-data"`
	AccountGasLimits   string `json:"account-gas-limits"`
	PreVerificationGas string `json:"pre-verification-gas"`
	GasFees            string `json:"gas-fees"`
	PaymasterAndData   string `json:"paymaster-and-data"`
	Signature          string `json:"signature"`
}

type PaymasterAndData struct {
	Paymaster    string `json:"paymaster"`
	Signer       string `json:"signer"`
	Escrow       string `json:"escrow"`
	TargetDomain string `json:"target-domain"`
	AssetAddress string `json:"asset-address"`
	AssetAmount  string `json:"asset-amount"`
	Calldata     string `json:"calldata"`
}

type UnsignedDataResponse struct {
	Signer        string                      `json:"signer"`
	ScwInit       bool                        `json:"swc-init"`
	EscrowInit    bool                        `json:"escrow-init"`
	EscrowPayload string                      `json:"escrow-payload"`
	EscrowTarget  string                      `json:"escrow-target"`
	EscrowValue   string                      `json:"escrow-value"`  // need to implement
	UserOp        PackedUserOperationResponse `json:"packed-userop"` // parsed data, recommended to validate data
	UserOpHash    string                      `json:"userop-hash"`
}

type UnsignedDataResponse2 struct {
	Signer           string                      `json:"signer"`
	ScwInit          bool                        `json:"swc-init"`
	Escrow           string                      `json:"escrow"`
	EscrowInit       string                      `json:"escrow-init"`
	EscrowPayload    string                      `json:"escrow-payload"`
	EscrowAsset      string                      `json:"escrow-asset"`
	EscrowValue      string                      `json:"escrow-value"`  // need to implement
	UserOp           PackedUserOperationResponse `json:"packed-userop"` // parsed data, recommended to validate data
	PaymasterAndData PaymasterAndData            `json:"paymaster-and-data"`
	UserOpHash       string                      `json:"userop-hash"`
}

type TestReceipt struct {
	Success string `json:"success"`
	TxHash  string `json:"tx-hash"`
}
