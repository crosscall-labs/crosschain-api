package handler

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	//"github.com/laminafinance/crosschain-api/api/main/utils"
	handler "github.com/laminafinance/crosschain-api/api/main"
	"github.com/laminafinance/crosschain-api/internal/utils"
	"golang.org/x/crypto/sha3"
)

// type VersionResponse struct {
// 	Version string `json:"version"`
// }

// type UnsignedBytecodeParams struct {
// 	MessageType  string `query:"msg-type" optional:"true"`
// 	Signer       string `query:"signer"`
// 	TargetId     string `query:"destination-id"`
// 	OriginId     string `query:"origin-id"`
// 	AssetAmount  string `query:"asset-amount"`
// 	AssetAddress string `query:"asset-address"`
// 	Calldata     string `query:"calldata" optional:"true"`
// }

type SignedBytecodeParams struct {
	Signer                   string `query:"signer"`
	DestinationId            string `query:"destination-id"`
	OriginId                 string `query:"origin-id"`
	AssetAddress             string `query:"asset-address"`
	AssetAmount              string `query:"asset-amount"`
	UseropSender             string `query:"useropSender"`
	UseropNonce              string `query:"useropNonce"`
	UseropInitCode           string `query:"useropInitCode" optional:"true"`
	UseropCallData           string `query:"useropCallData"`
	UseropAccountGasLimit    string `query:"useropAccountGasLimit"`
	UseropPreVerificationGas string `query:"useropPreVerificationGas"`
	UseropGasFees            string `query:"useropGasFees"`
	UseropPaymasterAndData   string `query:"useropPaymasterAndData"`
	UseropSignature          string `query:"useropSignature"`
}

type SignedEscrowPayoutParams struct {
	Bytecode string `query:"data"`
	TraceId  string `query:"traceid"`
}

// Error data structure
// type Error struct {
// 	Code    uint64 `json:"code"`
// 	Message string `json:"message"`
// 	Details string `json:"details"`
// }

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

var privateKey *ecdsa.PrivateKey
var relayAddress common.Address

type TestReceipt struct {
	Success string `json:"success"`
	TxHash  string `json:"tx-hash"`
}

// need query for creating an escrow lock
// this means that the query needs to call the escrow contract with the correct initializer data and salt
// need salt variable
// need initializer {eoaowner and delegate}, delegate is stored at:
//	storage slot 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc of the escrow contract

// need query for creating userop + scw initcode + paymasteranddata

func Handler(w http.ResponseWriter, r *http.Request) {

	privateKeyString := os.Getenv("RELAY_PRIVATE_KEY")
	var err error
	privateKey, err = crypto.HexToECDSA(privateKeyString)
	if err != nil {
		errInternal(w, fmt.Sprintf("Error converting private key: %v", err))
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		errInternal(w, "Error casting public key to ECDSA")
		return
	}

	relayAddress = crypto.PubkeyToAddress(*publicKeyECDSA)
	saltHash := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000037")

	// Cast common.Hash to [32]byte
	var SALT [32]byte
	copy(SALT[:], saltHash[:]) // 55

	contractAbiEntrypoint := `[{"type":"receive","stateMutability":"payable"},{"type":"function","name":"addStake","inputs":[{"name":"unstakeDelaySec","type":"uint32","internalType":"uint32"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"delegateAndRevert","inputs":[{"name":"target","type":"address","internalType":"address"},{"name":"data","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"depositTo","inputs":[{"name":"account","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"deposits","inputs":[{"name":"","type":"address","internalType":"address"}],"outputs":[{"name":"deposit","type":"uint256","internalType":"uint256"},{"name":"staked","type":"bool","internalType":"bool"},{"name":"stake","type":"uint112","internalType":"uint112"},{"name":"unstakeDelaySec","type":"uint32","internalType":"uint32"},{"name":"withdrawTime","type":"uint48","internalType":"uint48"}],"stateMutability":"view"},{"type":"function","name":"getDepositInfo","inputs":[{"name":"account","type":"address","internalType":"address"}],"outputs":[{"name":"info","type":"tuple","internalType":"struct IStakeManager.DepositInfo","components":[{"name":"deposit","type":"uint256","internalType":"uint256"},{"name":"staked","type":"bool","internalType":"bool"},{"name":"stake","type":"uint112","internalType":"uint112"},{"name":"unstakeDelaySec","type":"uint32","internalType":"uint32"},{"name":"withdrawTime","type":"uint48","internalType":"uint48"}]}],"stateMutability":"view"},{"type":"function","name":"getNonce","inputs":[{"name":"sender","type":"address","internalType":"address"},{"name":"key","type":"uint192","internalType":"uint192"}],"outputs":[{"name":"nonce","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"getSenderAddress","inputs":[{"name":"initCode","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"getUserOpHash","inputs":[{"name":"userOp","type":"tuple","internalType":"struct PackedUserOperation","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"handleAggregatedOps","inputs":[{"name":"opsPerAggregator","type":"tuple[]","internalType":"struct IEntryPoint.UserOpsPerAggregator[]","components":[{"name":"userOps","type":"tuple[]","internalType":"struct PackedUserOperation[]","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]},{"name":"aggregator","type":"address","internalType":"contract IAggregator"},{"name":"signature","type":"bytes","internalType":"bytes"}]},{"name":"beneficiary","type":"address","internalType":"address payable"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"handleOps","inputs":[{"name":"ops","type":"tuple[]","internalType":"struct PackedUserOperation[]","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]},{"name":"beneficiary","type":"address","internalType":"address payable"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"incrementNonce","inputs":[{"name":"key","type":"uint192","internalType":"uint192"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"innerHandleOp","inputs":[{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"opInfo","type":"tuple","internalType":"struct EntryPoint.UserOpInfo","components":[{"name":"mUserOp","type":"tuple","internalType":"struct EntryPoint.MemoryUserOp","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"verificationGasLimit","type":"uint256","internalType":"uint256"},{"name":"callGasLimit","type":"uint256","internalType":"uint256"},{"name":"paymasterVerificationGasLimit","type":"uint256","internalType":"uint256"},{"name":"paymasterPostOpGasLimit","type":"uint256","internalType":"uint256"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"paymaster","type":"address","internalType":"address"},{"name":"maxFeePerGas","type":"uint256","internalType":"uint256"},{"name":"maxPriorityFeePerGas","type":"uint256","internalType":"uint256"}]},{"name":"userOpHash","type":"bytes32","internalType":"bytes32"},{"name":"prefund","type":"uint256","internalType":"uint256"},{"name":"contextOffset","type":"uint256","internalType":"uint256"},{"name":"preOpGas","type":"uint256","internalType":"uint256"}]},{"name":"context","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"actualGasCost","type":"uint256","internalType":"uint256"}],"stateMutability":"nonpayable"},{"type":"function","name":"nonceSequenceNumber","inputs":[{"name":"","type":"address","internalType":"address"},{"name":"","type":"uint192","internalType":"uint192"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"supportsInterface","inputs":[{"name":"interfaceId","type":"bytes4","internalType":"bytes4"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"unlockStake","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawStake","inputs":[{"name":"withdrawAddress","type":"address","internalType":"address payable"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawTo","inputs":[{"name":"withdrawAddress","type":"address","internalType":"address payable"},{"name":"withdrawAmount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"event","name":"AccountDeployed","inputs":[{"name":"userOpHash","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"sender","type":"address","indexed":true,"internalType":"address"},{"name":"factory","type":"address","indexed":false,"internalType":"address"},{"name":"paymaster","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"event","name":"BeforeExecution","inputs":[],"anonymous":false},{"type":"event","name":"Deposited","inputs":[{"name":"account","type":"address","indexed":true,"internalType":"address"},{"name":"totalDeposit","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"PostOpRevertReason","inputs":[{"name":"userOpHash","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"sender","type":"address","indexed":true,"internalType":"address"},{"name":"nonce","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"revertReason","type":"bytes","indexed":false,"internalType":"bytes"}],"anonymous":false},{"type":"event","name":"SignatureAggregatorChanged","inputs":[{"name":"aggregator","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"StakeLocked","inputs":[{"name":"account","type":"address","indexed":true,"internalType":"address"},{"name":"totalStaked","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"unstakeDelaySec","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"StakeUnlocked","inputs":[{"name":"account","type":"address","indexed":true,"internalType":"address"},{"name":"withdrawTime","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"StakeWithdrawn","inputs":[{"name":"account","type":"address","indexed":true,"internalType":"address"},{"name":"withdrawAddress","type":"address","indexed":false,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"UserOperationEvent","inputs":[{"name":"userOpHash","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"sender","type":"address","indexed":true,"internalType":"address"},{"name":"paymaster","type":"address","indexed":true,"internalType":"address"},{"name":"nonce","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"success","type":"bool","indexed":false,"internalType":"bool"},{"name":"actualGasCost","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"actualGasUsed","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"UserOperationPrefundTooLow","inputs":[{"name":"userOpHash","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"sender","type":"address","indexed":true,"internalType":"address"},{"name":"nonce","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"UserOperationRevertReason","inputs":[{"name":"userOpHash","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"sender","type":"address","indexed":true,"internalType":"address"},{"name":"nonce","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"revertReason","type":"bytes","indexed":false,"internalType":"bytes"}],"anonymous":false},{"type":"event","name":"Withdrawn","inputs":[{"name":"account","type":"address","indexed":true,"internalType":"address"},{"name":"withdrawAddress","type":"address","indexed":false,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"error","name":"DelegateAndRevert","inputs":[{"name":"success","type":"bool","internalType":"bool"},{"name":"ret","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"FailedOp","inputs":[{"name":"opIndex","type":"uint256","internalType":"uint256"},{"name":"reason","type":"string","internalType":"string"}]},{"type":"error","name":"FailedOpWithRevert","inputs":[{"name":"opIndex","type":"uint256","internalType":"uint256"},{"name":"reason","type":"string","internalType":"string"},{"name":"inner","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"PostOpReverted","inputs":[{"name":"returnData","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"ReentrancyGuardReentrantCall","inputs":[]},{"type":"error","name":"SenderAddressResult","inputs":[{"name":"sender","type":"address","internalType":"address"}]},{"type":"error","name":"SignatureValidationFailed","inputs":[{"name":"aggregator","type":"address","internalType":"address"}]}]`
	//contractAbiEntrypointSimulations := ``
	contractAbiEscrow := `[{"type":"constructor","inputs":[{"name":"hyperlaneMailbox_","type":"address","internalType":"address"},{"name":"hyperlaneOrigin_","type":"address","internalType":"address"},{"name":"domain_","type":"uint32","internalType":"uint32"},{"name":"entrypoint_","type":"address","internalType":"address"},{"name":"interchainSecurityModule_","type":"address","internalType":"address"},{"name":"eoaRelay_","type":"address","internalType":"address"}],"stateMutability":"payable"},{"type":"receive","stateMutability":"payable"},{"type":"function","name":"_interchainSecurityModule","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"addHyperlane","inputs":[{"name":"hyperlaneOrigin_","type":"address","internalType":"address"},{"name":"state_","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"claim","inputs":[{"name":"asset_","type":"address","internalType":"address"},{"name":"amount_","type":"uint256","internalType":"uint256"},{"name":"to_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"delegateAddress","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"deposit","inputs":[{"name":"asset_","type":"address","internalType":"address"},{"name":"amount_","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"depositAndLock","inputs":[{"name":"asset_","type":"address","internalType":"address"},{"name":"amount_","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"entrypoint","inputs":[{"name":"","type":"uint32","internalType":"uint32"}],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"eoaRelay","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"extendLock","inputs":[{"name":"sec_","type":"uint256","internalType":"uint256"},{"name":"signature_","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"extendNonce","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"getEntrypoint","inputs":[{"name":"domain_","type":"uint32","internalType":"uint32"}],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"getEoaRelay","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"getHyperlaneMailbox","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"getHyperlaneOrigin","inputs":[{"name":"hyperlaneOrigin_","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"getUserOpHash","inputs":[{"name":"userOp","type":"tuple","internalType":"struct PackedUserOperation","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]},{"name":"entrypoint_","type":"address","internalType":"address"},{"name":"chainId_","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"handle","inputs":[{"name":"_origin","type":"uint32","internalType":"uint32"},{"name":"_sender","type":"bytes32","internalType":"bytes32"},{"name":"message","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"hashSeconds","inputs":[{"name":"account_","type":"address","internalType":"address"},{"name":"seconds_","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"hyperlaneMailbox","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"hyperlaneOrigin","inputs":[{"name":"","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"initialize","inputs":[{"name":"owner_","type":"address","internalType":"address"},{"name":"delegateAddress_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"interchainSecurityModule","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"lock","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"owner","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"releaseLock","inputs":[{"name":"asset_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"renounceOwnership","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"setEntrypoint","inputs":[{"name":"domain_","type":"uint32","internalType":"uint32"},{"name":"entrypoint_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"setEoaRelay","inputs":[{"name":"eoaRelay_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"setHyperlaneMailbox","inputs":[{"name":"hyperlaneMailbox_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"setHyperlaneOrigin","inputs":[{"name":"hyperlaneOrigin_","type":"address","internalType":"address"},{"name":"state_","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"setInterchainSecurityModule","inputs":[{"name":"interchainSecurityModule_","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"transferOwnership","inputs":[{"name":"newOwner","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdraw","inputs":[{"name":"asset_","type":"address","internalType":"address"},{"name":"amount_","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"event","name":"Initialized","inputs":[{"name":"version","type":"uint64","indexed":false,"internalType":"uint64"}],"anonymous":false},{"type":"event","name":"OwnershipTransferred","inputs":[{"name":"previousOwner","type":"address","indexed":true,"internalType":"address"},{"name":"newOwner","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"PrintUserOp","inputs":[{"name":"userOp","type":"tuple","indexed":false,"internalType":"struct PackedUserOperation","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]}],"anonymous":false},{"type":"event","name":"newBalance","inputs":[{"name":"asset","type":"address","indexed":false,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"error","name":"BadSignature","inputs":[]},{"type":"error","name":"BalanceError","inputs":[{"name":"requested","type":"uint256","internalType":"uint256"},{"name":"actual","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ECDSAInvalidSignature","inputs":[]},{"type":"error","name":"ECDSAInvalidSignatureLength","inputs":[{"name":"length","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ECDSAInvalidSignatureS","inputs":[{"name":"s","type":"bytes32","internalType":"bytes32"}]},{"type":"error","name":"InsufficentFunds","inputs":[{"name":"account","type":"address","internalType":"address"},{"name":"asset","type":"address","internalType":"address"},{"name":"amount","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"InvalidCCIPAddress","inputs":[{"name":"badSender","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidChain","inputs":[{"name":"badDestination","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"InvalidDeadline","inputs":[{"name":"","type":"string","internalType":"string"}]},{"type":"error","name":"InvalidDeltaValue","inputs":[]},{"type":"error","name":"InvalidHyperlaneAddress","inputs":[{"name":"badSender","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidInitialization","inputs":[]},{"type":"error","name":"InvalidLayerZeroAddress","inputs":[{"name":"badSender","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidOwner","inputs":[{"name":"owner","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidPaymaster","inputs":[{"name":"paymaster","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidSignature","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"notOwner","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidTimeInput","inputs":[]},{"type":"error","name":"NotInitializing","inputs":[]},{"type":"error","name":"OwnableInvalidOwner","inputs":[{"name":"owner","type":"address","internalType":"address"}]},{"type":"error","name":"OwnableUnauthorizedAccount","inputs":[{"name":"account","type":"address","internalType":"address"}]},{"type":"error","name":"PaymasterPaymentFailed","inputs":[{"name":"receiver","type":"address","internalType":"address"},{"name":"asset","type":"address","internalType":"address"},{"name":"amount","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"TransferFailed","inputs":[]},{"type":"error","name":"WithdrawRejected","inputs":[{"name":"","type":"string","internalType":"string"}]},{"type":"error","name":"testerror","inputs":[{"name":"","type":"bytes","internalType":"bytes"}]}]`
	contractAbiEscrowFactory := `[{"type":"constructor","inputs":[{"name":"_escrowImpl","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"function","name":"VERSION","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"createEscrow","inputs":[{"name":"_initializer","type":"bytes","internalType":"bytes"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"proxy","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"function","name":"escrowImpl","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"getEscrowAddress","inputs":[{"name":"_initializer","type":"bytes","internalType":"bytes"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"proxy","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"proxyCode","inputs":[],"outputs":[{"name":"","type":"bytes","internalType":"bytes"}],"stateMutability":"pure"}]`
	contractAbiSimpleAccountFactory := `[{"type":"constructor","inputs":[{"name":"_entryPoint","type":"address","internalType":"contract IEntryPoint"}],"stateMutability":"nonpayable"},{"type":"function","name":"accountImplementation","inputs":[],"outputs":[{"name":"","type":"address","internalType":"contract SimpleAccount"}],"stateMutability":"view"},{"type":"function","name":"createAccount","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"salt","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"ret","type":"address","internalType":"contract SimpleAccount"}],"stateMutability":"nonpayable"},{"type":"function","name":"getAddress","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"salt","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"}]`
	contractAbiSimpleAccount := `[{"type":"constructor","inputs":[{"name":"anEntryPoint","type":"address","internalType":"contract IEntryPoint"}],"stateMutability":"nonpayable"},{"type":"receive","stateMutability":"payable"},{"type":"function","name":"UPGRADE_INTERFACE_VERSION","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"addDeposit","inputs":[],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"entryPoint","inputs":[],"outputs":[{"name":"","type":"address","internalType":"contract IEntryPoint"}],"stateMutability":"view"},{"type":"function","name":"execute","inputs":[{"name":"dest","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"},{"name":"func","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"executeBatch","inputs":[{"name":"dest","type":"address[]","internalType":"address[]"},{"name":"value","type":"uint256[]","internalType":"uint256[]"},{"name":"func","type":"bytes[]","internalType":"bytes[]"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"getDeposit","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"getNonce","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"initialize","inputs":[{"name":"anOwner","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"onERC1155BatchReceived","inputs":[{"name":"","type":"address","internalType":"address"},{"name":"","type":"address","internalType":"address"},{"name":"","type":"uint256[]","internalType":"uint256[]"},{"name":"","type":"uint256[]","internalType":"uint256[]"},{"name":"","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bytes4","internalType":"bytes4"}],"stateMutability":"pure"},{"type":"function","name":"onERC1155Received","inputs":[{"name":"","type":"address","internalType":"address"},{"name":"","type":"address","internalType":"address"},{"name":"","type":"uint256","internalType":"uint256"},{"name":"","type":"uint256","internalType":"uint256"},{"name":"","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bytes4","internalType":"bytes4"}],"stateMutability":"pure"},{"type":"function","name":"onERC721Received","inputs":[{"name":"","type":"address","internalType":"address"},{"name":"","type":"address","internalType":"address"},{"name":"","type":"uint256","internalType":"uint256"},{"name":"","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bytes4","internalType":"bytes4"}],"stateMutability":"pure"},{"type":"function","name":"owner","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"proxiableUUID","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"supportsInterface","inputs":[{"name":"interfaceId","type":"bytes4","internalType":"bytes4"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"upgradeToAndCall","inputs":[{"name":"newImplementation","type":"address","internalType":"address"},{"name":"data","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"validateUserOp","inputs":[{"name":"userOp","type":"tuple","internalType":"struct PackedUserOperation","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]},{"name":"userOpHash","type":"bytes32","internalType":"bytes32"},{"name":"missingAccountFunds","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"validationData","type":"uint256","internalType":"uint256"}],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawDepositTo","inputs":[{"name":"withdrawAddress","type":"address","internalType":"address payable"},{"name":"amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"event","name":"Initialized","inputs":[{"name":"version","type":"uint64","indexed":false,"internalType":"uint64"}],"anonymous":false},{"type":"event","name":"SimpleAccountInitialized","inputs":[{"name":"entryPoint","type":"address","indexed":true,"internalType":"contract IEntryPoint"},{"name":"owner","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"Upgraded","inputs":[{"name":"implementation","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"error","name":"AddressEmptyCode","inputs":[{"name":"target","type":"address","internalType":"address"}]},{"type":"error","name":"ECDSAInvalidSignature","inputs":[]},{"type":"error","name":"ECDSAInvalidSignatureLength","inputs":[{"name":"length","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ECDSAInvalidSignatureS","inputs":[{"name":"s","type":"bytes32","internalType":"bytes32"}]},{"type":"error","name":"ERC1967InvalidImplementation","inputs":[{"name":"implementation","type":"address","internalType":"address"}]},{"type":"error","name":"ERC1967NonPayable","inputs":[]},{"type":"error","name":"FailedInnerCall","inputs":[]},{"type":"error","name":"InvalidInitialization","inputs":[]},{"type":"error","name":"NotInitializing","inputs":[]},{"type":"error","name":"UUPSUnauthorizedCallContext","inputs":[]},{"type":"error","name":"UUPSUnsupportedProxiableUUID","inputs":[{"name":"slot","type":"bytes32","internalType":"bytes32"}]}]`
	contractAbiHyperlaneMailbox := `[{"type":"constructor","inputs":[{"name":"domain_","type":"uint32","internalType":"uint32"}],"stateMutability":"payable"},{"type":"receive","stateMutability":"payable"},{"type":"function","name":"dispatch","inputs":[{"name":"_destinationDomain","type":"uint32","internalType":"uint32"},{"name":"_recipientAddress","type":"bytes32","internalType":"bytes32"},{"name":"_messageBody","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"nonpayable"},{"type":"function","name":"handleDispatch","inputs":[{"name":"destinationDomain","type":"uint256","internalType":"uint256"},{"name":"recipientAddress","type":"address","internalType":"address"},{"name":"messageBody","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"payMessage","inputs":[{"name":"messageId","type":"bytes32","internalType":"bytes32"},{"name":"refundAddress","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"quoteGas","inputs":[{"name":"destinationDomain","type":"uint32","internalType":"uint32"},{"name":"gasAmount","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"}]`
	contractAbiHyperlaneIgp := `[{"type":"constructor","inputs":[{"name":"hyperlaneMailbox_","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"receive","stateMutability":"payable"},{"type":"function","name":"payForGas","inputs":[{"name":"_messageId","type":"bytes32","internalType":"bytes32"},{"name":"_destinationDomain","type":"uint32","internalType":"uint32"},{"name":"_gasAmount","type":"uint256","internalType":"uint256"},{"name":"_refundAddress","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"quoteGasPayment","inputs":[{"name":"_destinationDomain","type":"uint32","internalType":"uint32"},{"name":"_gasAmount","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"}]`
	contractAbiMulticall := `[{"type":"receive","stateMutability":"payable"},{"type":"function","name":"at","inputs":[{"name":"_addr","type":"address","internalType":"address"}],"outputs":[{"name":"o_code","type":"bytes","internalType":"bytes"}],"stateMutability":"view"},{"type":"function","name":"getExtcodesize","inputs":[{"name":"address_","type":"address","internalType":"address"}],"outputs":[{"name":"size_","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"multicallExecute","inputs":[{"name":"calls","type":"tuple[]","internalType":"struct Multicall.Call2[]","components":[{"name":"target","type":"address","internalType":"address"},{"name":"success","type":"bool","internalType":"bool"},{"name":"isStatic","type":"bool","internalType":"bool"},{"name":"value","type":"uint256","internalType":"uint256"},{"name":"callData","type":"bytes","internalType":"bytes"}]}],"outputs":[{"name":"","type":"tuple[]","internalType":"struct Multicall.Result[]","components":[{"name":"success","type":"bool","internalType":"bool"},{"name":"returnData","type":"bytes","internalType":"bytes"}]}],"stateMutability":"payable"},{"type":"function","name":"multicallExecuteAll","inputs":[{"name":"calls","type":"tuple[]","internalType":"struct Multicall.Call3[]","components":[{"name":"target","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"},{"name":"callData","type":"bytes","internalType":"bytes"}]}],"outputs":[{"name":"","type":"tuple[]","internalType":"struct Multicall.Result[]","components":[{"name":"success","type":"bool","internalType":"bool"},{"name":"returnData","type":"bytes","internalType":"bytes"}]}],"stateMutability":"payable"},{"type":"function","name":"multicallView","inputs":[{"name":"calls","type":"tuple[]","internalType":"struct Multicall.Call[]","components":[{"name":"target","type":"address","internalType":"address"},{"name":"callData","type":"bytes","internalType":"bytes"}]}],"outputs":[{"name":"","type":"tuple[]","internalType":"struct Multicall.Result[]","components":[{"name":"success","type":"bool","internalType":"bool"},{"name":"returnData","type":"bytes","internalType":"bytes"}]}],"stateMutability":"view"}]`
	contractAbiPaymaster := `[{"type":"constructor","inputs":[{"name":"entryPoint_","type":"address","internalType":"contract IEntryPoint"},{"name":"hyperlaneMailbox_","type":"address","internalType":"address"},{"name":"hyperlaneIgp_","type":"address","internalType":"address"},{"name":"defaultReceiver_","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"fallback","stateMutability":"payable"},{"type":"receive","stateMutability":"payable"},{"type":"function","name":"acceptedAsset","inputs":[{"name":"","type":"uint256","internalType":"uint256"},{"name":"","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"acceptedChain","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"addAcceptedAsset","inputs":[{"name":"chainId_","type":"uint256","internalType":"uint256"},{"name":"asset_","type":"address","internalType":"address"},{"name":"state_","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"addAcceptedChain","inputs":[{"name":"chainId_","type":"uint256","internalType":"uint256"},{"name":"state_","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"addStake","inputs":[{"name":"unstakeDelaySec","type":"uint32","internalType":"uint32"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"deposit","inputs":[],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"entryPoint","inputs":[],"outputs":[{"name":"","type":"address","internalType":"contract IEntryPoint"}],"stateMutability":"view"},{"type":"function","name":"escrowAddress","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"getDeposit","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"owner","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"postOp","inputs":[{"name":"mode","type":"uint8","internalType":"enum IPaymaster.PostOpMode"},{"name":"context","type":"bytes","internalType":"bytes"},{"name":"actualGasCost","type":"uint256","internalType":"uint256"},{"name":"actualUserOpFeePerGas","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"renounceOwnership","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"transferOwnership","inputs":[{"name":"newOwner","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"unlockStake","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"validatePaymasterUserOp","inputs":[{"name":"userOp","type":"tuple","internalType":"struct PackedUserOperation","components":[{"name":"sender","type":"address","internalType":"address"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"initCode","type":"bytes","internalType":"bytes"},{"name":"callData","type":"bytes","internalType":"bytes"},{"name":"accountGasLimits","type":"bytes32","internalType":"bytes32"},{"name":"preVerificationGas","type":"uint256","internalType":"uint256"},{"name":"gasFees","type":"bytes32","internalType":"bytes32"},{"name":"paymasterAndData","type":"bytes","internalType":"bytes"},{"name":"signature","type":"bytes","internalType":"bytes"}]},{"name":"userOpHash","type":"bytes32","internalType":"bytes32"},{"name":"maxCost","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"context","type":"bytes","internalType":"bytes"},{"name":"validationData","type":"uint256","internalType":"uint256"}],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawStake","inputs":[{"name":"withdrawAddress","type":"address","internalType":"address payable"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawTo","inputs":[{"name":"withdrawAddress","type":"address","internalType":"address payable"},{"name":"amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"event","name":"OwnershipTransferred","inputs":[{"name":"previousOwner","type":"address","indexed":true,"internalType":"address"},{"name":"newOwner","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"error","name":"InvalidAsset","inputs":[{"name":"chainId","type":"uint32","internalType":"uint32"},{"name":"asset","type":"address","internalType":"address"}]},{"type":"error","name":"InvalidChainId","inputs":[{"name":"chainId","type":"uint32","internalType":"uint32"}]},{"type":"error","name":"InvalidDataLength","inputs":[{"name":"dataLength","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"InvalidOrigin","inputs":[{"name":"bundler","type":"address","internalType":"address"}]},{"type":"error","name":"OwnableInvalidOwner","inputs":[{"name":"owner","type":"address","internalType":"address"}]},{"type":"error","name":"OwnableUnauthorizedAccount","inputs":[{"name":"account","type":"address","internalType":"address"}]}]`

	contractABIs := map[string]string{
		"Entrypoint":           contractAbiEntrypoint,
		"SimpleAccount":        contractAbiSimpleAccount,
		"SimpleAccountFactory": contractAbiSimpleAccountFactory,
		"Multicall":            contractAbiMulticall,
		"HyperlaneMailbox":     contractAbiHyperlaneMailbox,
		"HyperlaneIgp":         contractAbiHyperlaneIgp,
		"Paymaster":            contractAbiPaymaster,
		"Escrow":               contractAbiEscrow,
		"EscrowFactory":        contractAbiEscrowFactory,
	}

	parsedABIs := make(map[string]abi.ABI)

	for name, abiStr := range contractABIs {
		parsedABI, err := abi.JSON(strings.NewReader(abiStr))
		if err != nil {
			fmt.Println(err)
			errInternal(w, err.Error())
			return
		}
		parsedABIs[name] = parsedABI
	}

	switch r.URL.Query().Get("query") {
	case "version":
		Version(w)
		return
	case "unsigned-bytecode":
		UnsignedBytecode(w, r)
		return
	case "signed-bytecode":
		SignedBytecode(w, r)
	case "signed-escrow-payout":
		// will add env restriction on origin later
		SignedEscrowPayout(w, r)
		TestRun(w, r)
		return
	case "test-run":
		TestRun(w, r)
		return
	default:
		version := "Hello, World!"
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func TestRun(w http.ResponseWriter, r *http.Request) {
	// Call the function from utils.go to write a JSON response
	//WriteJSONResponse(w, "TestRun executed successfully")
	utils.WriteJSONResponse(w, r, "Hello, World!")
}

func Version(w http.ResponseWriter) {
	versionResponse := utils.VersionResponse{Version: "Crosschain DEX API v0.0.4"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(versionResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func UnsignedBytecode(w http.ResponseWriter, r *http.Request) {
	params := &handler.UnsignedBytecodeParams{}

	// Parse and validate query parameters
	if !utils.ParseAndValidateParams(w, r, params) {
		return
	}

	// calldata: abi.encodeWithSignature("execute(address,uint256,bytes)", rando, 5 ether, hex"");
	// rando is signer

	// assetAmountInt, err := strconv.ParseInt(params.AssetAmount, 10, 64)
	// if err != nil {
	// 	errMalformedRequest(w, "Invalid integer for 'asset-amount'")
	// 	return
	// }

	// // connect to RPC
	// client, chainInfo, ok := checkClient(w, params.OriginId)
	// if !ok {
	// 	return
	// }

	// client2, chainInfo2, ok := checkClient(w, params.TargetId)
	// if !ok {
	// 	return
	// }

	// fmt.Printf("calldata: %s\n", useropCallData)
	// var unsignedDataResponse UnsignedDataResponse
	// var packedUserOperation PackedUserOperation

	// accountGasLimitsBytes := common.Hex2Bytes("0x00000000000000000000000001312d0000000000000000000000000000989680")
	// var accountGasLimits [32]byte
	// copy(accountGasLimits[:], accountGasLimitsBytes)

	// gasFeeBytes := common.Hex2Bytes("0x0000000000000000000000000000000200000000000000000000000000000000")
	// var gasFees [32]byte
	// copy(gasFees[:], gasFeeBytes)
	// packedUserOperation = PackedUserOperation{
	// 	Sender:             packedUserOperation.Sender,
	// 	Nonce:              packedUserOperation.Nonce,
	// 	InitCode:           packedUserOperation.InitCode,
	// 	CallData:           common.FromHex(calldata),
	// 	AccountGasLimits:   [32]byte(common.FromHex("0x0000000000000000000000000098968000000000000000000000000000989680")),
	// 	PreVerificationGas: big.NewInt(20000000),
	// 	GasFees:            [32]byte(common.FromHex("0x0000000000000000000000000000000200000000000000000000000000000000")),
	// 	PaymasterAndData:   packedUserOperation.PaymasterAndData,
	// 	Signature:          packedUserOperation.Signature,
	// }
	// // PrintUserOp(
	// // 	userOp: PackedUserOperation({
	// // 		sender: 0x907d3e885b8f286F27ED469aBB0e317BD62a7Fd3,
	// // 		nonce: 0,
	// // 		initCode: 0x2e234dae75c793f67a35089c9d99245e1c58470b5fbfb9cf000000000000000000000000f814aa444c49a5dbbbf8f59a654036a0ede26cce0000000000000000000000000000000000000000000000000000000000000055,
	// // 		callData: 0xb61d27f600000000000000000000000074bd103dbc4fa5187ca3d0914e560afdb81f5f340000000000000000000000000000000000000000000000004563918244f4000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000,
	// // 		accountGasLimits: 0x00000000000000000000000001312d0000000000000000000000000000989680,
	// // 		preVerificationGas: 20000000 [2e7],
	// // 		gasFees: 0x0000000000000000000000000000000200000000000000000000000000000000,
	// // 		paymasterAndData: 0xc7183455a4c133ae270771860664b6b7ec320bb10000000000000000000000000098968000000000000000000000000000989680f814aa444c49a5dbbbf8f59a654036a0ede26cce0000000000000000000000000000000000000000000000000000000000aa36a700000000000000000000000000000000000000000000000000000000000000000000000000000000000000004563918244f40000,
	// // 		signature: 0x5f4b4180c74fa301e8383304c8c43fa267a84674dba6365fd8d415f2ff775ce0446688d4b0145af3a51e98cee6f0fdc66522ed935437baa04b1e4c79214daa1c1c }))
	// unsignedDataResponse.Signer = signer
	// unsignedDataResponse.UserOp.CallData = calldata
	// unsignedDataResponse.UserOp.AccountGasLimits = "0x00000000000000000000000001312d0000000000000000000000000000989680"
	// unsignedDataResponse.UserOp.PreVerificationGas = "20000000"
	// unsignedDataResponse.UserOp.GasFees = "0x0000000000000000000000000000000200000000000000000000000000000000"

	// initializerBytes, err := GetViewCallBytes(*client, parsedABIs["Escrow"], "initialize", common.HexToAddress(signer), common.HexToAddress(chainInfo.AddressEscrow))
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// calls := []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	params          []interface{}
	// }{
	// 	{
	// 		contractName: "EscrowFactory",
	// 		method:       "getEscrowAddress",
	// 		params:       []interface{}{initializerBytes, SALT},
	// 	},
	// }

	// results, err := getMulticallViewResults(client, parsedABIs, chainInfo, calls)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// if !results[0].Success {
	// 	fmt.Printf("Escrow: getEscrowAddress failed for chain chain %s\n", chainInfo.ChainId)
	// 	errInternal(w)
	// 	return
	// }

	// parsedResults, err := parsedABIs["EscrowFactory"].Unpack("getEscrowAddress", results[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// escrowAddress := parsedResults[0].(common.Address)
	// packedUserOperation.Sender = escrowAddress
	// unsignedDataResponse.UserOp.Sender = escrowAddress.Hex()

	// calls2 := []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	params          []interface{}
	// }{
	// 	{
	// 		contractName: "SimpleAccountFactory",
	// 		method:       "getAddress",
	// 		params:       []interface{}{common.HexToAddress(signer), new(big.Int).SetBytes(SALT[:])},
	// 	},
	// }

	// results2, err := getMulticallViewResults(client2, parsedABIs, chainInfo2, calls2)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// if !results2[0].Success {
	// 	fmt.Printf("SCW: getAddress failed for chain chain %s\n", chainInfo2.ChainId)
	// 	errInternal(w)
	// 	return
	// }

	// parsedResults2, err := parsedABIs["SimpleAccountFactory"].Unpack("getAddress", results2[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// scwAddress := parsedResults2[0].(common.Address)

	// //escrowAddress
	// //scwAddress

	// calls = []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	params          []interface{}
	// }{
	// 	{
	// 		contractName: "Multicall",
	// 		method:       "getExtcodesize",
	// 		params:       []interface{}{escrowAddress},
	// 	},
	// }

	// results, err = getMulticallViewResults(client, parsedABIs, chainInfo, calls)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// extcodesize := parsedResults[0].(*big.Int)
	// calls2 = []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	params          []interface{}
	// }{
	// 	{
	// 		contractName: "Multicall",
	// 		method:       "getExtcodesize",
	// 		params:       []interface{}{scwAddress},
	// 	},
	// 	{
	// 		contractName: "Entrypoint",
	// 		method:       "getNonce",
	// 		params:       []interface{}{common.HexToAddress(signer), big.NewInt(55)},
	// 	},
	// }

	// results2, err = getMulticallViewResults(client2, parsedABIs, chainInfo2, calls2)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results2[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// extcodesize2 := parsedResults[0].(*big.Int)

	// parsedResults2, err = parsedABIs["Entrypoint"].Unpack("getNonce", results2[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// scwNonce := parsedResults2[0].(*big.Int)

	// var executionCalls []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	value           *big.Int
	// 	params          []interface{}
	// }
	// if extcodesize.Int64() > 0 { // escrow
	// 	executionCalls = []struct {
	// 		contractName    string
	// 		contractAddress string
	// 		method          string
	// 		value           *big.Int
	// 		params          []interface{}
	// 	}{
	// 		{
	// 			contractName:    "Escrow", //deposit(address asset_, uint256 amount_)
	// 			contractAddress: escrowAddress.Hex(),
	// 			method:          "depositAndLock",
	// 			value:           big.NewInt(assetAmountInt), // should be zero for token (not yet handled)
	// 			params:          []interface{}{common.HexToAddress(assetAddress), big.NewInt(assetAmountInt)},
	// 		},
	// 		// {
	// 		// 	contractName: escrowAddress.Hex(),
	// 		// 	method:       "extendLock",
	// 		// 	value:        *common.Big0,
	// 		// 	params:       []interface{}{},
	// 		// },
	// 	}

	// 	unsignedDataResponse.EscrowInit = false
	// } else {
	// 	executionCalls = []struct {
	// 		contractName    string
	// 		contractAddress string
	// 		method          string
	// 		value           *big.Int
	// 		params          []interface{}
	// 	}{
	// 		{
	// 			contractName: "EscrowFactory",
	// 			method:       "createEscrow",
	// 			value:        common.Big0,
	// 			params:       []interface{}{initializerBytes, SALT},
	// 		},
	// 		{
	// 			contractName:    "Escrow", //deposit(address asset_, uint256 amount_)
	// 			contractAddress: escrowAddress.Hex(),
	// 			method:          "depositAndLock",
	// 			value:           big.NewInt(assetAmountInt), // should be zero for token (not yet handled)
	// 			params:          []interface{}{common.HexToAddress(assetAddress), big.NewInt(assetAmountInt)},
	// 		},
	// 		// { // chnaged to require signature
	// 		// 	contractName:    "Escrow",
	// 		// 	contractAddress: scwAddress.Hex(),
	// 		// 	method:          "extendLock",
	// 		// 	value:           *common.Big0,
	// 		// 	params:          []interface{}{},
	// 		// },
	// 	}

	// 	unsignedDataResponse.EscrowInit = true
	// }
	// fmt.Println("got this far7")
	// escrowPayload, err := getMulticallExecuteAllBytecode(client, parsedABIs, chainInfo, executionCalls)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// unsignedDataResponse.EscrowPayload = "0x" + common.Bytes2Hex(escrowPayload)
	// unsignedDataResponse.EscrowTarget = chainInfo.AddressMulticall
	// unsignedDataResponse.EscrowValue = assetAmount // should gas and paymaster costs
	// fmt.Println("got this far8")
	// initcodeCall, err := GetViewCallBytes(*client2, parsedABIs["SimpleAccountFactory"], "createAccount", common.HexToAddress(signer), new(big.Int).SetBytes(SALT[:]))
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// fmt.Println("got this far9")
	// initcodeBytecode := append(common.Hex2Bytes(chainInfo2.AddressSimpleAccountFactory), initcodeCall...)

	// if extcodesize2.Int64() > 0 { // scw
	// 	packedUserOperation.Nonce = scwNonce
	// 	packedUserOperation.InitCode = []byte{}
	// 	unsignedDataResponse.ScwInit = false
	// 	unsignedDataResponse.UserOp.InitCode = common.Bytes2Hex([]byte{})
	// } else {
	// 	packedUserOperation.InitCode = initcodeBytecode
	// 	packedUserOperation.Nonce = common.Big0
	// 	unsignedDataResponse.ScwInit = true
	// 	unsignedDataResponse.UserOp.InitCode = "0x" + common.Bytes2Hex(initcodeBytecode)
	// 	unsignedDataResponse.UserOp.Nonce = "0"
	// }

	// // lets output everything into paymasteranddata field of UserOp, for testing
	// // paymaster prefix good, now suffix
	// paymasterSigner := common.FromHex(signer)
	// someint, err := strconv.Atoi(chainInfo.ChainId)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// paymasterTarget := padLeftHex(someint)
	// paymasterAsset := common.FromHex(assetAddress) // need to check for badd addresses
	// someint, err = strconv.Atoi(assetAmount)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// paymasterAmount := padLeftHex(someint)
	// paymasterPrefix := append(common.FromHex(chainInfo2.AddressPaymaster), common.FromHex("0x0000000000000000000000000098968000000000000000000000000000989680")...)
	// packedUserOperation.PaymasterAndData = bytes.Join([][]byte{
	// 	paymasterPrefix,
	// 	paymasterSigner,
	// 	paymasterTarget,
	// 	paymasterAsset,
	// 	paymasterAmount,
	// }, nil)
	// unsignedDataResponse.UserOp.PaymasterAndData = "0x" + common.Bytes2Hex(packedUserOperation.PaymasterAndData)

	// returnData, err := ViewFunction(*client2, common.HexToAddress(chainInfo2.AddressEntrypoint), parsedABIs["Entrypoint"], "getUserOpHash", packedUserOperation)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// unsignedDataResponse.UserOpHash = "0x" + common.Bytes2Hex(returnData)

	// //w.WriteHeader(http.StatusOK)
	// w.Header().Set("Content-Type", "application/json")
	// if err := json.NewEncoder(w).Encode(unsignedDataResponse); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
}

func SignedBytecode(w http.ResponseWriter, r *http.Request) {
	params := &SignedBytecodeParams{}

	if !utils.ParseAndValidateParams(w, r, params) {
		return
	}

	// client, chainInfo, ok := checkClient(w, params.OriginId)
	// if !ok {
	// 	return
	// }

	// client2, chainInfo2, ok := checkClient(w, params.TargetId)
	// if !ok {
	// 	return
	// }

	//var packedUserOperation PackedUserOperation

	// // need to fetch sender using signer
	// packedUserOperation = PackedUserOperation{
	// 	Sender:             common.HexToAddress(useropSender),
	// 	Nonce:              packedUserOperation.Nonce,
	// 	InitCode:           common.FromHex(useropInitCode),
	// 	CallData:           common.FromHex(useropCallData),
	// 	AccountGasLimits:   [32]byte(common.FromHex(useropAccountGasLimit)),
	// 	PreVerificationGas: packedUserOperation.PreVerificationGas,
	// 	GasFees:            [32]byte(common.FromHex(useropGasFees)),
	// 	PaymasterAndData:   common.FromHex(useropPaymasterAndData),
	// 	Signature:          common.FromHex(useropSignature),
	// }

	// var someint int64
	// // parse nonce to proper format
	// someint, err = strconv.ParseInt(useropNonce, 10, 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// packedUserOperation.Nonce = big.NewInt(someint)
	// // parse perverificationgas to proper format
	// someint, err = strconv.ParseInt(useropPreVerificationGas, 10, 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// packedUserOperation.PreVerificationGas = big.NewInt(someint)

	// // evaluate is paymaster matches expected cost
	// var paymasterAndData []byte
	// paymasterPrefix := append(common.FromHex(chainInfo2.AddressPaymaster), common.FromHex("0x0000000000000000000000000098968000000000000000000000000000989680")...)
	// paymasterSigner := common.FromHex(signer)
	// someint, err = strconv.ParseInt(originId, 10, 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// paymasterOrigin := padLeftHex(int(someint))
	// paymasterAsset := common.FromHex(assetAddress)
	// someint, err = strconv.ParseInt(assetAmount, 10, 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// paymasterAmount := padLeftHex(int(someint))
	// paymasterAndData = bytes.Join([][]byte{
	// 	paymasterPrefix,
	// 	paymasterSigner,
	// 	paymasterOrigin,
	// 	paymasterAsset,
	// 	paymasterAmount,
	// }, nil)
	// if !bytes.Equal(packedUserOperation.PaymasterAndData, paymasterAndData) {
	// 	fmt.Printf("packedUserOperation.PaymasterAndData: %s", common.Bytes2Hex(packedUserOperation.PaymasterAndData))
	// 	fmt.Printf("paymasterAndData: %s", common.Bytes2Hex(paymasterAndData))
	// 	errPaymasterAndDataMismatch(w)
	// 	return
	// }

	// initializerBytes, err := GetViewCallBytes(*client, parsedABIs["Escrow"], "initialize", common.HexToAddress(signer), common.HexToAddress(chainInfo.AddressEscrow))
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// calls := []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	params          []interface{}
	// }{
	// 	{
	// 		contractName: "EscrowFactory",
	// 		method:       "getEscrowAddress",
	// 		params:       []interface{}{initializerBytes, SALT},
	// 	},
	// }

	// results, err := getMulticallViewResults(client, parsedABIs, chainInfo, calls)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// if !results[0].Success {
	// 	fmt.Printf("Escrow: getEscrowAddress failed for chain chain %s\n", chainInfo.ChainId)
	// 	errInternal(w)
	// 	return
	// }

	// parsedResults, err := parsedABIs["EscrowFactory"].Unpack("getEscrowAddress", results[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// escrowAddress := parsedResults[0].(common.Address)

	// calls = []struct {
	// 	contractName    string
	// 	contractAddress string
	// 	method          string
	// 	params          []interface{}
	// }{
	// 	{
	// 		contractName: "Multicall",
	// 		method:       "getExtcodesize",
	// 		params:       []interface{}{escrowAddress},
	// 	},
	// 	// {
	// 	// 	contractName: "Escrow",
	// 	// 	contractAddress: escrowAddress.Hex(),
	// 	// 	method: // no public function for mapping(address => uint256) assetLocked;
	// 	// }
	// }

	// results, err = getMulticallViewResults(client, parsedABIs, chainInfo, calls)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results[0].ReturnData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// extcodesize := parsedResults[0].(*big.Int)

	// if extcodesize.Int64() == 0 {
	// 	errEscrowNotFound(w)
	// 	return
	// }

	// // because no public function, just call balance because we are only using address(0)
	// escrowBalance, err := client.BalanceAt(context.Background(), escrowAddress, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// someint, err = strconv.ParseInt(assetAmount, 10, 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// compareResult := escrowBalance.Cmp(big.NewInt(someint))
	// if compareResult == -1 {
	// 	errInsufficientEscrowBalance(w)
	// 	return
	// }

	// // need to validate userop, not going to happen need to use entrypointsimulations
	// // calling base simpleaccount
	// // validateUserOp
	// // function validateUserOp(
	// // 			PackedUserOperation calldata userOp,
	// // 			bytes32 userOpHash,
	// // 			uint256 missingAccountFunds
	// // 	) external virtual override returns (uint256 validationData) {
	// // 			_requireFromEntryPoint();
	// // 			validationData = _validateSignature(userOp, userOpHash);
	// // 			_validateNonce(userOp.nonce);
	// // 			_payPrefund(missingAccountFunds);
	// // 	}

	// // var executablePackedUserop []PackedUserOperation
	// // executablePackedUserop = append(executablePackedUserop, packedUserOperation)

	// someint, err = strconv.ParseInt(assetAmount, 10, 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }
	// // datainput, err := parsedABIs["Entrypoint"].Pack("handleOps", executablePackedUserop, common.HexToAddress("0xaeD6b252635DcEF5Ba85dE52173FF040a18CEC6a"))
	// // if err != nil {
	// // 	fmt.Print(err)
	// // 	return //fmt.Errorf("failed to pack multicallExecuteAll input: %v", err)
	// // }

	// // var noinput []byte
	// //recipet, data, err := PackedExecuteFunction(*client2, common.HexToAddress(chainInfo2.AddressEntrypoint), common.Big0, datainput)
	// // recipet, data, err := PackedExecuteFunction(*client2, common.HexToAddress(signer), big.NewInt(someint), noinput)
	// // if err != nil {
	// // 	fmt.Print(err)
	// // 	return //fmt.Errorf("failed to pack multicallExecuteAll input: %v", err)
	// // }

	// // fmt.Printf("recipet: %s", recipet)
	// // fmt.Printf("data: %s", data)

	// gasPrice, _ := client2.SuggestGasPrice(context.Background())
	// fmt.Printf("gasPrice: %s\n", gasPrice)
	// fmt.Printf("gasPrice: %s\n", gasPrice)
	// // //lchainid, _ := client2.ChainID(context.Background())
	// // // auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, lchainid)
	// // // auth.Value = big.NewInt(1000000000000000000)

	// // addy := common.HexToAddress(signer)
	// // callMsg := ethereum.CallMsg{
	// // 	From:     relayAddress,
	// // 	To:       &addy,
	// // 	Gas:      0,
	// // 	GasPrice: gasPrice,
	// // 	Value:    big.NewInt(someint),
	// // 	Data:     noinput,
	// // }

	// // _, _ = client.CallContract(context.Background(), callMsg, nil)

	// receipt, err := TransferEth(*client, "8e80f019af2ae825c10e261594aa7ce5f8898fcc30eec7a25110a906914968d7", signer, someint)
	// if err != nil {
	// 	fmt.Println(err)
	// 	errInternal(w)
	// 	return
	// }

	// fmt.Printf("receipt: %s\n", receipt)

	// fmt.Println(chainInfo2)
	// // TestReceipt
	// // for not only handle escrowPayload, which is a payload to execute test contract increment
	// w.WriteHeader(http.StatusOK)
	// w.Header().Set("Content-Type", "application/json")
	// if err := json.NewEncoder(w).Encode(packedUserOperation); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
}

func UnsignedEscrowPayout(w http.ResponseWriter, r *http.Request) {
	// params := &UnsignedEscrowPayoutParams{}

	// if !parseAndValidateParams(w, r, params) {
	// 	return
	// }
	/*
	   _CrosschainPaymasterAddress 0x3647fbDD26946850f7A18599394A4685aaD550BC
	   paymasterAndDataType1_ data:
	   0x0000000000000000000000003647fbdd26946850f7a18599394a4685aad550bc00000000000000000000000000989680000000000000000000000000009896800000000000000000000000000000000000000000000000000000000000000001000000000000000000000000f814aa444c49a5dbbbf8f59a654036a0ede26cce00000000000000000000000006e7cb26c760a7a2b72cd73515de65ee431b01240000000000000000000000000000000000000000000000000000000000aa36a700000000000000000000000000000000000000000000000000000000499602d20000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
	   The above data is expected to be returned with
	   signer
	   data
	   v
	   s
	   r
	*/

	// fmt.Printf("\ninput data: %s", params.Bytecode)
	// fmt.Printf("\ntrace id: %s", params.TraceId)
}

func SignedEscrowPayout(w http.ResponseWriter, r *http.Request) {
	params := &SignedEscrowPayoutParams{}

	if !utils.ParseAndValidateParams(w, r, params) {
		return
	}
	/*
	   _CrosschainPaymasterAddress 0x3647fbDD26946850f7A18599394A4685aaD550BC
	   paymasterAndDataType1_ data:
	   0x0000000000000000000000003647fbdd26946850f7a18599394a4685aad550bc00000000000000000000000000989680000000000000000000000000009896800000000000000000000000000000000000000000000000000000000000000001000000000000000000000000f814aa444c49a5dbbbf8f59a654036a0ede26cce00000000000000000000000006e7cb26c760a7a2b72cd73515de65ee431b01240000000000000000000000000000000000000000000000000000000000aa36a700000000000000000000000000000000000000000000000000000000499602d20000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
	   The above data is expected to be returned signed as
	   signer
	   data
	   v will be skipped for now due to a known bug
	   s
	   r
	*/

	fmt.Printf("\ninput data: %s", params.Bytecode)
	fmt.Printf("\ntrace id: %s", params.TraceId)
}

func ViewFunction(client ethclient.Client, contractAddress common.Address, parsedABI abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{To: &contractAddress, Data: data}
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func TransferEth(client ethclient.Client, privKey string, to string, amount int64) (string, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// Assuming you've already connected a client, the next step is to load your private key.
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return "", err
	}

	// Function requires the public address of the account we're sending from -- which we can derive from the private key.
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Now we can read the nonce that we should use for the account's transaction.
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	value := big.NewInt(amount) // in wei (1 eth)
	gasLimit := uint64(21000)   // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	// We figure out who we're sending the ETH to.
	toAddress := common.HexToAddress(to)
	var data []byte

	// We create the transaction payload
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	// We sign the transaction using the sender's private key
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	// Now we are finally ready to broadcast the transaction to the entire network
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	// We return the transaction hash
	return signedTx.Hash().String(), nil
}

func PackedViewFunction(client ethclient.Client, contractAddress common.Address, packedData []byte) ([]byte, error) {
	block_, err := GetLatestBlock(client)
	if err != nil {
		return nil, err
	}

	blockNumber := big.NewInt(int64(block_.BlockNumber))
	callMsg := ethereum.CallMsg{To: &contractAddress, Data: packedData}
	//var result []Result
	result, err := client.CallContract(context.Background(), callMsg, blockNumber)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ExecuteFunction(client ethclient.Client, contractAddress common.Address, parsedABI abi.ABI, methodName string, value *big.Int, args ...interface{}) (receiptJSON []byte, err error) {
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, err
	}
	auth.Value = big.NewInt(1000000000000000000)

	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{
		From:     relayAddress,
		To:       &contractAddress,
		Gas:      0,
		GasPrice: gasPrice,
		Value:    value,
		Data:     data,
	}

	_, err = client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}

	nonce, err := client.PendingNonceAt(context.Background(), relayAddress)
	if err != nil {
		return nil, err
	}

	estimatedGas, err := client.EstimateGas(context.Background(), callMsg)
	if err != nil {
		return nil, err
	}

	gasLimit := 120 * estimatedGas / 100

	tx := types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainId), privateKey)
	if err != nil {
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	receipt, err := bind.WaitMined(context.Background(), &client, signedTx)
	if err != nil {
		return nil, err
	}

	var returnedData []byte
	for _, log := range receipt.Logs {
		if len(log.Data) > 0 {
			// Assuming the returned data is in the first log entry
			returnedData = log.Data
			break
		}
	}

	// receiptJSON, err = json.Marshal(receipt)
	// if err != nil {
	// 	log.Fatalf("Failed to JSON marshal receipt: %v", err)
	// 	return nil, err
	// }

	return returnedData, nil
}

func createMulticallExecuteAllData(chainInfo Chain, tuples [][3]interface{}) []byte {
	paddedTuples := make([][]byte, len(tuples))
	paddedTuplesLen := len(tuples)

	// create tuple raw bytes
	for i, tuple := range tuples {
		addy := tuple[0].(common.Address)
		addrBytes := padLeft(addy.Bytes())
		valueBytes := padLeftHex(tuple[1].(int))
		dataBytes := tuple[2].([]byte)
		paddedLen := ((len(dataBytes) + 31) / 32) * 32 // future error?
		paddedBytes := make([]byte, paddedLen)
		copy(paddedBytes, dataBytes)

		// Concatenate the padded address and padded bytes
		tupleBytes := append(addrBytes, valueBytes...)
		tupleBytes = append(tupleBytes, common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000060")...)
		tupleBytes = append(tupleBytes, padLeftHex(len(dataBytes))...)
		tupleBytes = append(tupleBytes, paddedBytes...)

		paddedTuples[i] = tupleBytes
	}

	var buffer bytes.Buffer

	parse, _ := common.ParseHexOrString("multicallExecuteAll((address,uint256,bytes)[])")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(parse)
	selector := hash.Sum(nil)[:4]

	buffer.Write(selector)
	buffer.Write(common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000020"))
	buffer.Write(padLeftHex(paddedTuplesLen))
	buffer.Write(padLeftHex(paddedTuplesLen * 32))

	var sum int
	for i := 1; i < len(paddedTuples); i++ {
		sum += len(paddedTuples[i-1]) // Adjust index to access the correct tuple
		buffer.Write(padLeftHex(sum + paddedTuplesLen*32))
	}

	for _, paddedTuple := range paddedTuples {
		buffer.Write(paddedTuple)
	}

	return buffer.Bytes()
}

func PackedExecuteFunction(client ethclient.Client, contractAddress common.Address, value *big.Int, packedData []byte) (receiptJSON []byte, returnedData []byte, err error) {
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("I got into here %s", gasPrice)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, nil, err
	}
	auth.Value = big.NewInt(1000000000000000000)

	callMsg := ethereum.CallMsg{
		From:     relayAddress,
		To:       &contractAddress,
		Gas:      3,
		GasPrice: gasPrice,
		Value:    value,
		Data:     packedData,
	}
	fmt.Printf("I got into here0")
	returnedData, err = client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, nil, err
	}

	nonce, err := client.PendingNonceAt(context.Background(), relayAddress)
	if err != nil {
		return nil, nil, err
	}

	estimatedGas, err := client.EstimateGas(context.Background(), callMsg)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("I got into here")
	gasLimit := 120 * estimatedGas / 100

	tx := types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, packedData)

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainId), privateKey)
	if err != nil {
		return nil, nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, nil, err
	}

	receipt, err := bind.WaitMined(context.Background(), &client, signedTx)
	if err != nil {
		return nil, nil, err
	}

	receiptJSON, err = json.Marshal(receipt)
	if err != nil {
		log.Fatalf("Failed to JSON marshal receipt: %v", err)
		return nil, nil, err
	}

	return receiptJSON, returnedData, nil
}

func GetViewCallBytes(client ethclient.Client, parsedABI abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		fmt.Printf("some error data \n")
		return nil, err
	}
	return data, nil
}

func getMulticallViewResults(client *ethclient.Client, parsedABIs map[string]abi.ABI, chainInfo *Chain, calls []struct {
	contractName    string
	contractAddress string
	method          string
	params          []interface{}
}) ([]Result, error) {
	results, err := multicallView(client, parsedABIs, chainInfo, calls)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func getMulticallExecuteAllBytecode(client *ethclient.Client, parsedABIs map[string]abi.ABI, chainInfo *Chain, calls []struct {
	contractName    string
	contractAddress string
	method          string
	value           *big.Int
	params          []interface{}
}) ([]byte, error) {
	var multicallViewInput []Call3

	for _, call := range calls {
		if strings.HasPrefix(call.contractAddress, "0x") {
			target := common.HexToAddress(call.contractAddress)
			//value := call.value
			packedData, err := parsedABIs[call.contractName].Pack(call.method, call.params...)
			if err != nil {
				return nil, fmt.Errorf("failed to pack data: %v", err)
			}

			c := Call3{
				Target:   target,
				Value:    call.value,
				CallData: packedData,
			}

			multicallViewInput = append(multicallViewInput, c)
		} else {
			c, err := createCall3(client, parsedABIs, chainInfo, call.contractName, call.method, common.Big0, call.params...)
			if err != nil {
				return nil, fmt.Errorf("failed to create call: %v", err)
			}

			multicallViewInput = append(multicallViewInput, c)
		}
	}

	data, err := parsedABIs["Multicall"].Pack("multicallExecuteAll", multicallViewInput)
	if err != nil {
		return nil, fmt.Errorf("failed to pack multicallExecuteAll input: %v", err)
	}

	return data, nil
}

func multicallView(client *ethclient.Client, parsedABIs map[string]abi.ABI, chainInfo *Chain, calls []struct {
	contractName    string
	contractAddress string
	method          string
	params          []interface{}
}) ([]Result, error) {
	var multicallViewInput []Call
	for _, call := range calls {
		if strings.HasPrefix(call.contractName, "0x") {
			target := common.HexToAddress(call.contractAddress)
			packedData, err := parsedABIs[call.contractName].Pack(call.method, call.params...)
			if err != nil {
				return nil, fmt.Errorf("failed to pack data: %v", err)
			}

			c := Call{
				Target:   target,
				CallData: packedData,
			}

			multicallViewInput = append(multicallViewInput, c)
		} else {
			c, err := createCall(client, parsedABIs, chainInfo, call.contractName, call.method, call.params...)
			if err != nil {
				return nil, fmt.Errorf("failed to create call: %v", err)
			}

			multicallViewInput = append(multicallViewInput, c)
		}
	}

	returnData, err := ViewFunction(*client, common.HexToAddress(chainInfo.AddressMulticall), parsedABIs["Multicall"], "multicallView", multicallViewInput)
	if err != nil {
		return nil, fmt.Errorf("failed to execute multicallView: %v", err)
	}

	data, err := parsedABIs["Multicall"].Unpack("multicallView", returnData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack multicallView result: %v", err)
	}

	var results []Result
	for _, v := range data {
		for _, vv := range v.([]struct {
			Success    bool   "json:\"success\""
			ReturnData []byte "json:\"returnData\""
		}) {
			results = append(results, Result{
				Success:    vv.Success,
				ReturnData: vv.ReturnData,
			})
		}
	}

	return results, nil
}

func createCall(client *ethclient.Client, parsedABIs map[string]abi.ABI, chainInfo *Chain, contractName, method string, params ...interface{}) (Call, error) {
	callData, err := GetViewCallBytes(*client, parsedABIs[contractName], method, params...)
	if err != nil {
		return Call{}, err
	}

	var target common.Address
	switch contractName {
	case "Entrypoint":
		target = common.HexToAddress(chainInfo.AddressEntrypoint)
	case "SimpleAccountFactory":
		target = common.HexToAddress(chainInfo.AddressSimpleAccountFactory)
	case "Multicall":
		target = common.HexToAddress(chainInfo.AddressMulticall)
	case "HyperlaneMailbox":
		target = common.HexToAddress(chainInfo.AddressHyperlaneMailbox)
	case "HyperlaneIgp":
		target = common.HexToAddress(chainInfo.AddressHyperlaneIgp)
	case "Paymaster":
		target = common.HexToAddress(chainInfo.AddressPaymaster)
	case "Escrow":
		target = common.HexToAddress(chainInfo.AddressEscrow)
	case "EscrowFactory":
		target = common.HexToAddress(chainInfo.AddressEscrowFactory)
	default:
		return Call{}, fmt.Errorf("unsupported contract name: %s", contractName)
	}

	return Call{
		Target:   target,
		CallData: callData,
	}, nil
}

// if address is provided need to auto create one on the fly
func createCall3(client *ethclient.Client, parsedABIs map[string]abi.ABI, chainInfo *Chain, contractName, method string, value *big.Int, params ...interface{}) (Call3, error) {
	callData, err := GetViewCallBytes(*client, parsedABIs[contractName], method, params...)
	if err != nil {
		return Call3{}, err
	}

	var target common.Address
	switch contractName {
	case "Entrypoint":
		target = common.HexToAddress(chainInfo.AddressEntrypoint)
	case "SimpleAccountFactory":
		target = common.HexToAddress(chainInfo.AddressSimpleAccountFactory)
	case "Multicall":
		target = common.HexToAddress(chainInfo.AddressMulticall)
	case "HyperlaneMailbox":
		target = common.HexToAddress(chainInfo.AddressHyperlaneMailbox)
	case "HyperlaneIgp":
		target = common.HexToAddress(chainInfo.AddressHyperlaneIgp)
	case "Paymaster":
		target = common.HexToAddress(chainInfo.AddressPaymaster)
	case "Escrow":
		target = common.HexToAddress(chainInfo.AddressEscrow)
	case "EscrowFactory":
		target = common.HexToAddress(chainInfo.AddressEscrowFactory)
	default:
		return Call3{}, fmt.Errorf("unsupported contract name: %s", contractName)
	}

	return Call3{
		Target:   target,
		Value:    value,
		CallData: callData,
	}, nil
}

func padLeft(b []byte) []byte {
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

func padLeftHex(value int) []byte {
	hexStr := fmt.Sprintf("%064x", value)
	padded, _ := hex.DecodeString(hexStr)
	return padded
}

func GetLatestBlock(client ethclient.Client) (*Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve latest block header: %w", err)
	}
	if header == nil {
		return nil, fmt.Errorf("latest block header is nil")
	}

	blockNumber := big.NewInt(header.Number.Int64())
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blcok by number: %w", err)
	}

	_block := &Block{
		BlockNumber:       block.Number().Int64(),
		Timestamp:         block.Time(),
		Difficulty:        block.Difficulty().Uint64(),
		Hash:              block.Hash().String(),
		TransactionsCount: len(block.Transactions()),
		Transactions:      []Transaction{},
	}

	// We add a recover function from panics to prevent our API from crashing due to an unexpected error
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// Query the latest block
	header, _ = client.HeaderByNumber(context.Background(), nil)
	blockNumber = big.NewInt(header.Number.Int64())
	block, err = client.BlockByNumber(context.Background(), blockNumber)

	if err != nil {
		log.Fatal(err)
	}

	// Build the response to our model
	_block = &Block{
		BlockNumber:       block.Number().Int64(),
		Timestamp:         block.Time(),
		Difficulty:        block.Difficulty().Uint64(),
		Hash:              block.Hash().String(),
		TransactionsCount: len(block.Transactions()),
		Transactions:      []Transaction{},
	}

	for _, tx := range block.Transactions() {
		_block.Transactions = append(_block.Transactions, Transaction{
			Hash:     tx.Hash().String(),
			Value:    tx.Value().String(),
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice().Uint64(),
			Nonce:    tx.Nonce(),
			To:       tx.To().String(),
		})
	}

	return _block, nil
}

// case "0x310C5", "200901": // bitlayer mainnet
// case "0xE35", "3637": // botanix mainnet
// case "0xC4", "196": // x layer mainnet
// case "0xA4EC", "42220": // celo mainnet
// case "0x82750", "534352": // scroll mainnet
// case "0xA", "10": // op mainnet
// case "0xA4B1", "42161": // arbitrum one
// case "0x2105", "8453": // base mainnet
// case "0x13A", "314": // filecoin mainnet
// case "0x63630000", "1667432448": // tvm workchain_id == 0
// case "0x53564D0001", "357930172419": // solana mainnet
// case "0xBF04", "48900": // zircuit mainnet
func checkChainType(chainId string) (string, []int, []int, error) { // out: vm, entrypointType, escrowTyoe, error
	disabled := fmt.Errorf("unsupported chain ID: %s", chainId)
	switch chainId {
	case "0x3106A", "200810": // bitlayer testnet
		return "evm", []int{0, 1}, []int{0, 1, 2}, nil
	case "0x4268", "17000": // holesky
		return "evm", []int{0, 1}, []int{0, 1, 2}, nil
	case "0xAA36A7", "11155111": // sepolia
		return "evm", []int{0, 1}, []int{0, 1, 2}, nil
	case "0xE34", "3636": // botanix testnet
		return "evm", []int{0, 1}, []int{0, 1, 2}, disabled
	case "0xF35A", "62298": // citrea testnet
		return "evm", []int{0, 1}, []int{0, 1, 2}, nil
	case "0x13881", "80001": // matic mumbai
		return "evm", nil, nil, disabled
	case "0x13882", "80002": // matic amoy
		return "evm", nil, nil, disabled
	case "0xC3", "195": // x layer testnet
		return "evm", nil, nil, disabled
	case "0xAEF3", "44787": // celo alfajores
		return "evm", nil, nil, disabled
	case "0x5E9", "1513": // story testnet
		return "evm", nil, nil, disabled
	case "0x8274F", "534351": // scroll testnet
		return "evm", nil, nil, disabled
	case "0xAA37DC", "11155420": // op sepolia
		return "evm", nil, nil, disabled
	case "0x66EEE", "421614": // arbitrum sepolia
		return "evm", nil, nil, disabled
	case "0x14A34", "84532": // base sepolia
		return "evm", nil, nil, disabled
	case "0x4CB2F", "314159": // filecoin calibration
		return "evm", nil, nil, disabled
	case "0xBF03", "48899": // zircuit testnet
		return "evm", nil, nil, disabled
	case "0x63639999", "1667471769": // tvm workchain_id == -1 ton testnet
		return "evm", []int{2}, []int{0, 1, 2}, nil
	case "0x53564D0002", "357930172418": // solana devnet
		return "svm", nil, nil, disabled
	case "0x53564D0003", "357930172419": // solana testnet
		return "svm", nil, nil, disabled
	case "0x53564D0004", "357930172420": // eclipse (solana) testnet
		return "svm", nil, nil, disabled
	default:
		return "NaN", nil, nil, disabled
	}
}

func hasInt(inputArray []int, input int) error {
	for value := range inputArray {
		if value == input {
			return nil
		}
	}
	return fmt.Errorf("int not found in input array")
}

//unsignedDataResponse.Escrow, unsignedDataResponse.EscrowInit := createEscrowBytecode(signer, originId, assetAddress, assetAmount)

// func createEscrowBytecode(messageTypeInt int, signer string, originId string, assetAddress string, assetAmount string) (string, string, error) {
// 	chainType, entrypointTypes, _, err := checkChainType(originId)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	err = hasInt(entrypointTypes, messageTypeInt)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	switch chainType {
// 	case "evm":
// 		//return createEscrowBytecodeEVM(messageTypeInt, signer, originId, assetAddress, assetAmount)
// 	case "tvm":
// 		return createEscrowBytecodeTVM(messageTypeInt, signer, originId, assetAddress, assetAmount)
// 	case "svm":
// 		return createEscrowBytecodeSVM(messageTypeInt, signer, originId, assetAddress, assetAmount)
// 	default:
// 		return "", "", &Error{
// 			Code:    500,
// 			Message: "Internal error: chain type could not be determined",
// 		}
// 	}
// }

// still require the messageTypeInt because we allow for multiple signature schema
// func createEscrowBytecodeEVM(messageTypeInt int, signer string, originId string, assetAddress string, assetAmount string) (string, string, error) {
// 	initializerBytes, err := GetViewCallBytes(*client, parsedABIs["Escrow"], "initialize", common.HexToAddress(signer), common.HexToAddress(chainInfo.AddressEscrow))
// 	if err != nil {
// 		fmt.Println(err)
// 		errInternal(w)
// 		return
// 	}

// 	calls := []struct {
// 		contractName    string
// 		contractAddress string
// 		method          string
// 		params          []interface{}
// 	}{
// 		{
// 			contractName: "EscrowFactory",
// 			method:       "getEscrowAddress",
// 			params:       []interface{}{initializerBytes, SALT},
// 		},
// 	}

// 	results, err := getMulticallViewResults(client, parsedABIs, chainInfo, calls)
// 	if err != nil {
// 		fmt.Println(err)
// 		errInternal(w)
// 		return
// 	}

// 	if !results[0].Success {
// 		fmt.Printf("Escrow: getEscrowAddress failed for chain chain %s\n", chainInfo.ChainId)
// 		errInternal(w)
// 		return
// 	}

// 	parsedResults, err := parsedABIs["EscrowFactory"].Unpack("getEscrowAddress", results[0].ReturnData)
// 	if err != nil {
// 		fmt.Println(err)
// 		errInternal(w)
// 		return
// 	}

// 	escrowAddress := parsedResults[0].(common.Address)

// 	calls = []struct {
// 		contractName    string
// 		contractAddress string
// 		method          string
// 		params          []interface{}
// 	}{
// 		{
// 			contractName: "Multicall",
// 			method:       "getExtcodesize",
// 			params:       []interface{}{escrowAddress},
// 		},
// 	}

// 	results, err = getMulticallViewResults(client, parsedABIs, chainInfo, calls)
// 	if err != nil {
// 		fmt.Println(err)
// 		errInternal(w)
// 		return
// 	}

// 	parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results[0].ReturnData)
// 	if err != nil {
// 		fmt.Println(err)
// 		errInternal(w)
// 		return
// 	}
// 	return "", "", nil
// }

func createEscrowBytecodeTVM(messageTypeInt int, signer string, originId string, assetAddress string, assetAmount string) (string, string, error) {
	return "", "", nil
}

func createEscrowBytecodeSVM(messageTypeInt int, signer string, originId string, assetAddress string, assetAmount string) (string, string, error) {
	return "", "", nil
}

func checkChainStatus(chainId string) (*ethclient.Client, *Chain, error) {
	var client *ethclient.Client
	var chain *Chain
	var err error

	var rpcURL string
	var addresses Chain

	switch chainId {
	case "0x3106A", "200810":
		chainId = "200810"
		rpcURL = "https://testnet-rpc.bitlayer.org"
		addresses = Chain{
			AddressEntrypoint:            "0x317bBdFbAe7845648864348A0C304392d0F2925F",
			AddressEntrypointSimulations: "0x6960fA06d5119258533B5d715c8696EE66ca4042",
			AddressSimpleAccountFactory:  "0xCF730748FcDc78A5AB854B898aC24b6d6001AbF7",
			AddressSimpleAccount:         "0xfaAe830bA56C40d17b7e23bfe092f23503464114",
			AddressMulticall:             "0x66e4f2437c5F612Ae25e94C1C549cb9f151E0cB3",
			AddressHyperlaneMailbox:      "0x2EaAd60F982f7B99b42f30e98B3b3f8ff89C0A46",
			AddressHyperlaneIgp:          "0x16e81e1973939bD166FDc61651F731e1658060F3",
			AddressPaymaster:             "0xdAE5e7CEBe4872BF0776477EcCCD2A0eFdF54f0e",
			AddressEscrow:                "0x9925D4a40ea432A25B91ab424b16c8FC6e0Eec5A",
			AddressEscrowFactory:         "0xC531388B2C2511FDFD16cD48f1087A747DC34b33",
		}
	case "0x4268", "17000":
		chainId = "200810"
		rpcURL = "https://ethereum-holesky-rpc.publicnode.com"
		addresses = Chain{
			AddressEntrypoint:            "0xc5Ff094002cdaF36d6a766799eB63Ec82B8C79F1",
			AddressEntrypointSimulations: "0x67B9841e9864D394FDc02e787A0Ac37f32B49eC7",
			AddressSimpleAccountFactory:  "0x39351b719D044CF6E91DEC75E78e5d128c582bE7",
			AddressSimpleAccount:         "0x0983a4e9D9aB03134945BFc9Ec9EF31338AB7465",
			AddressMulticall:             "0x98876409cc48507f8Ee8A0CCdd642469DBfB3E21",
			AddressHyperlaneMailbox:      "0x913A6477496eeb054C9773843a64c8621Fc46e8C",
			AddressHyperlaneIgp:          "0x2Fb9F9bd9034B6A5CAF3eCDB30db818619EbE9f1",
			AddressPaymaster:             "0xA5bcda4aA740C02093Ba57A750a8f424BC8B4B13",
			AddressEscrow:                "0x686130A96724734F0B6f99C6D32213BC62C1830A",
			AddressEscrowFactory:         "0x45d5D46B097870223fDDBcA9a9eDe35A7D37e2A1",
		}
	case "0xaa36a7", "11155111":
		chainId = "11155111"
		rpcURL = "https://rpc2.sepolia.org"
		addresses = Chain{
			AddressEntrypoint:            "0xA6eBc93dA2C99654e7D6BC12ed24362061805C82",
			AddressEntrypointSimulations: "0x0d17dE0436b65279c8D7A75847F84626687A1647",
			AddressSimpleAccountFactory:  "0x54bed3E354cbF23C2CADaB1dF43399473e38a358",
			AddressSimpleAccount:         "0x54bed3E354cbF23C2CADaB1dF43399473e38a358",
			AddressMulticall:             "0x6958206f218D8f889ECBb76B89eE9bF1CAe37715",
			AddressHyperlaneMailbox:      "0xAc165ff97Dc42d87D858ba8BC4AA27429a8C48e8",
			AddressHyperlaneIgp:          "0x00eb6D45afac57E708eC3FA6214BFe900aFDb95D",
			AddressPaymaster:             "0x31aCA626faBd9df61d24A537ecb9D646994b4d4d",
			AddressEscrow:                "0xea8D264dF67c9476cA80A24067c2F3CF7726aC4d",
			AddressEscrowFactory:         "0xd9842E241B7015ea1E1B5A90Ae20b6453ADF2723",
		}
	case "0xe34", "3636":
		chainId = "3636"
		rpcURL = "https://node.botanixlabs.dev"
		addresses = Chain{
			AddressEntrypoint:            "0xF7B12fFBC58dd654aeA52f1c863bf3f4731f848F",
			AddressEntrypointSimulations: "0x1db7F1263FbfBe5d91548B3422563179f6bE8d99",
			AddressSimpleAccountFactory:  "0xFB23dB8098Faf2dB307110905dC3698Fe27E136d",
			AddressSimpleAccount:         "0x15aA997cC02e103a7570a1C26F09996f6FBc1829",
			AddressMulticall:             "0x6cB50ee0241C7AE6Ebc30A34a9F3C23A96098bBf",
			AddressHyperlaneMailbox:      "0xd2DB8440B7dC1d05aC2366b353f1cF205Cf875EA",
			AddressHyperlaneIgp:          "0x8439DBdca66C9F72725f1B2d50dFCdc7c6CBBbEb",
			AddressPaymaster:             "0xbbfb649f42Baf44729a150464CBf6B89349A634a",
			AddressEscrow:                "0xCD77545cA802c4B05ff359f7b10355EC220E7476",
			AddressEscrowFactory:         "0xA6eBc93dA2C99654e7D6BC12ed24362061805C82",
		}
	default:
		return nil, nil, fmt.Errorf("unsupported chain ID: %s", chainId)
	}

	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		return nil, nil, err
	}

	domain, err := strconv.ParseUint(chainId, 0, 32)
	if err != nil {
		return nil, nil, err
	}

	chain = &Chain{
		ChainId:                      chainId,
		Domain:                       uint32(domain),
		AddressEntrypoint:            addresses.AddressEntrypoint,
		AddressEntrypointSimulations: addresses.AddressEntrypointSimulations,
		AddressSimpleAccountFactory:  addresses.AddressSimpleAccountFactory,
		AddressMulticall:             addresses.AddressMulticall,
		AddressHyperlaneMailbox:      addresses.AddressHyperlaneMailbox,
		AddressHyperlaneIgp:          addresses.AddressHyperlaneIgp,
		AddressPaymaster:             addresses.AddressPaymaster,
		AddressEscrow:                addresses.AddressEscrow,
		AddressEscrowFactory:         addresses.AddressEscrowFactory,
	}

	return client, chain, nil
}

func checkClient(w http.ResponseWriter, chainId string) (*ethclient.Client, *Chain, bool) {
	client, chainInfo, err := checkChainStatus(chainId)
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return nil, nil, false
	}
	if client == nil {
		errUnsupportedChain(w)
		return nil, nil, false
	}
	return client, chainInfo, true
}

func errUnsupportedChain(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    0,
		Message: "Chain not currently supported",
	})
}

func errPaymasterAndDataMismatch(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    7,
		Message: "PaymasterAndData mismatch",
	})
}

func errInternal(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    500,
		Message: "Internal server error",
		Details: message,
	})
}

func errRpcFailed(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    501,
		Message: "Internal server error: RPC connection failed",
	})
}

func errEscrowNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    1000,
		Message: "Escrow address not exist",
	})
}

func errInsufficientEscrowBalance(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    1001,
		Message: "Insufficient escrow balance",
	})
}
