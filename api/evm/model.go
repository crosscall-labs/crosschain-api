package evmHandler

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

type UnsignedEntryPointRequestResponse struct {
	Header  utils.MessageHeader `query:"header"`
	Payload string              `query:"payload" optional:"true"`
}

type MessageOpEvm struct {
	UserOp           PackedUserOperationResponse `json:"op-packed-data"` // parsed data, recommended to validate data
	PaymasterAndData PaymasterAndDataResponse    `json:"op-paymaster"`
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
	Paymaster                     common.Address
	PaymasterVerificationGasLimit [32]byte
	PaymasterPostOpGasLimit       [32]byte
	Signer                        common.Address
	DestinationDomain             [4]byte
	MessageType                   byte
	AssetAddress                  common.Address
	AssetAmount                   *big.Int
}

// needs to return hash for recommended, but also needs to make it easy for solver to hash
// to hash abi encode packed the data (in order)
type PaymasterAndDataResponse struct {
	Paymaster                     string `json:"pad-paymaster"`
	PaymasterVerificationGasLimit string `json:"pad-verification-gas-limit"`
	PaymasterPostOpGasLimit       string `json:"pad-post-op-gas-limit"`
	Signer                        string `json:"pad-signer"`
	DestinationDomain             string `json:"pad-destination-domain"`
	MessageType                   string `json:"pad-message-type"`
	AssetAddress                  string `json:"pad-asset-address"`
	AssetAmount                   string `json:"pad-asset-amount"`
}

type MessageEscrowEvm struct {
	Init           EscrowInitRaw           `json:"init"`
	DepositAndLock EscrowDepositAndLockRaw `json:"deposit"`
	TimeLockHash   EscrowTimeLockHashRaw   `json:"timelock"`
}

type EscrowDepositAndLockRaw struct {
	AssetAddress  string `json:"asset-address"`
	AssetValue    string `json:"asset-value"`
	AssetAmount   string `json:"asset-amount"`
	AssetLocked   string `json:"asset-locked"`
	AssetDeadline string `json:"asset-dealine"`
	EscrowAddress string `json:"escrow-address"`
	Payload       string `json:"payload"`
}

type EscrowDepositAndLock struct {
	AssetAddress  common.Address
	AssetAmount   *big.Int
	EscrowAddress common.Address
}

type EscrowInitRaw struct {
	SingletonAddress string `json:"singleton"`
	FactoryAddress   string `json:"factory"`
	Salt             string `json:"salt"`
	IsInitialized    bool   `json:"is-initialized"`
	EscrowAddress    string `json:"address"`
	Initalizer       string `json:"initializer"`
	Payload          string `json:"payload"`
}

type EscrowInit struct {
	SingletonAddress common.Address
	FactoryAddress   common.Address
	Salt             []byte
	IsInitialized    bool
	EscrowAddress    common.Address
	Initalizer       []byte
	Payload          []byte
}

type EscrowTimeLockHashRaw struct {
	ExtendTime   string `json:"extend-time"`
	AssetAddress string `json:"asset-address"`
	ExtendNonce  string `json:"extend-nonce"`
	ChainId      string `json:"chain-id"`
	Hash         string `json:"hash"`
}

type EscrowTimeLockHash struct {
	ExtendTime   *big.Int
	AssetAddress common.Address
	ExtendNonce  *big.Int
	ChainId      *big.Int
	Hash         []byte
}

func (m MessageEscrowEvm) GetType() string {
	return "EVM Escrow"
}

func (m MessageOpEvm) GetType() string {
	return "EVM UserOp"
}

type Call struct {
	Target   common.Address
	CallData []byte
}

type Calls struct {
	contractAddress common.Address
	abi             abi.ABI
	method          string
	params          interface{}
}

type MulticallResult struct {
	Success    bool
	ReturnData []byte
}
