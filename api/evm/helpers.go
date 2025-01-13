package evmHandler

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/crosscall-labs/crosschain-api/pkg/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"
)

// calc address
// calc init
// check if exists

// estimate cost for init
// estimate cost of execute payout
// for now, we don't care but this should be tracked

// recommended lock amount + time
// lock payload
// full init + lock payload

func GetEscrowAssetInfo(
	client *ethclient.Client,
	escrowAddress common.Address,
	assetAddress common.Address) (*big.Int, *big.Int, *big.Int, error) {
	//getAssetInfo(assetAddress)

	parsedJSON, _ := abi.JSON(strings.NewReader(`[{
		"type":"function",
		"name":"getAssetInfo",
		"inputs":[
			{"name":"asset_","type":"address","internalType":"address"}
		],
		"outputs":[
			{"name":"","type":"uint256","internalType":"uint256"},
			{"name":"","type":"uint256","internalType":"uint256"},
			{"name":"","type":"uint256","internalType":"uint256"}],
		"stateMutability":"view"
	},{
		"type":"function",
		"name":"depositAndLock",
		"inputs":[
			{"name":"asset_","type":"address","internalType":"address"},
			{"name":"amount_","type":"uint256","internalType":"uint256"}
		],
		"outputs":[],
		"stateMutability":"payable"
	},{
		"type":"function",
		"name":"extendLockHash",
		"inputs":[
			{"name":"sec_","type":"uint256","internalType":"uint256"},
			{"name":"asset_","type":"address","internalType":"address"}
		],
		"outputs":[
			{"name":"","type":"bytes32","internalType":"bytes32"}
		],
		"stateMutability":"view"
	},{
		"type":"function",
		"name":"extendNonce",
		"inputs":[],
		"outputs":[
			{"name":"","type":"uint256","internalType":"uint256"}
		],
		"stateMutability":"view"
	}
	]`))

	response, err := ViewFunction(client, escrowAddress, parsedJSON, "getAssetInfo", assetAddress)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed getAssetInfo call: %v\n", err)
	}

	fmt.Printf("response: %v\n", response)

	parsedResults, err := parsedJSON.Unpack("getAssetInfo", response)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed getAssetInfo parse: %v\n", err)
	}
	fmt.Printf("parsedResults results: \n%v\n", parsedResults)

	return parsedResults[0].(*big.Int), parsedResults[1].(*big.Int), parsedResults[2].(*big.Int), nil
}

func TestRequest(r *http.Request, parameters ...*UnsignedEscrowRequestParams) (interface{}, error) {
	salt := common.Hex2Bytes("0x0000000000000000000000000000000000000000000000000000000000000037")
	signer := common.HexToAddress("19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A") // should be from params
	client, err := ethclient.Dial("https://rpc2.sepolia.org")                 // should be from inputs but ignored
	if err != nil {
		return nil, err
	}

	a := common.HexToAddress("f39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	size, err := ExtCodeSize(client, a)
	if err != nil {
		fmt.Printf("\ngot an error: \n%v\n", err)
	}
	fmt.Printf("\ncontract size: \n%v\n", size)

	isInit := false
	if size != 0 {
		isInit = true
	}

	messageEscrowEvm := MessageEscrowEvm{}
	extendNonce := common.Big0

	assetAddress := common.HexToAddress("0000000000000000000000000000000000000000")
	value := common.Big0
	escrowSingletonAddress := common.HexToAddress("ea8D264dF67c9476cA80A24067c2F3CF7726aC4d")
	escrowFactoryAddress := common.HexToAddress("d9842E241B7015ea1E1B5A90Ae20b6453ADF2723")
	//multicallAddress := common.HexToAddress("6958206f218D8f889ECBb76B89eE9bF1CAe37715")

	escrowAddressBytes, initalizerBytes, err := GetEscrowAddress(client, signer, escrowFactoryAddress, escrowSingletonAddress, salt)
	if err != nil {
		fmt.Printf("\ngot an error: \n%v\n", err)
	}

	parsedJSON, _ := abi.JSON(strings.NewReader(`[{
		"type":"function",
		"name":"getAssetInfo",
		"inputs":[
			{"name":"asset_","type":"address","internalType":"address"}
		],
		"outputs":[
			{"name":"","type":"uint256","internalType":"uint256"},
			{"name":"","type":"uint256","internalType":"uint256"},
			{"name":"","type":"uint256","internalType":"uint256"}],
		"stateMutability":"view"
	},{
		"type":"function",
		"name":"depositAndLock",
		"inputs":[
			{"name":"asset_","type":"address","internalType":"address"},
			{"name":"amount_","type":"uint256","internalType":"uint256"}
		],
		"outputs":[],
		"stateMutability":"payable"
	},{
		"type":"function",
		"name":"extendLockHash",
		"inputs":[
			{"name":"sec_","type":"uint256","internalType":"uint256"},
			{"name":"asset_","type":"address","internalType":"address"}
		],
		"outputs":[
			{"name":"","type":"bytes32","internalType":"bytes32"}
		],
		"stateMutability":"view"
	},{
		"type":"function",
		"name":"extendNonce",
		"inputs":[],
		"outputs":[
			{"name":"","type":"uint256","internalType":"uint256"}
		],
		"stateMutability":"view"
	}
	]`))

	assetAmount := common.Big0
	assetAmountLocked := common.Big0
	deadline := common.Big0
	if isInit {
		assetAmount, assetAmountLocked, deadline, _ = GetEscrowAssetInfo(client, common.BytesToAddress(escrowAddressBytes), assetAddress)
		fmt.Printf("output:\n%v\n%v\n%v\n%v\n%v\n", isInit, initalizerBytes, assetAmount, assetAmountLocked, deadline)

		response, err := ViewFunction(client, common.BytesToAddress(escrowAddressBytes), parsedJSON, "extendNonce")
		if err != nil {
			return nil, fmt.Errorf("failed extendNonce call: %v\n", err)
		}

		parsedResults, err := parsedJSON.Unpack("extendNonce", response)
		if err != nil {
			return nil, fmt.Errorf("failed extendNonce parse: %v\n", err)
		}

		extendNonce = parsedResults[0].(*big.Int)
	} else {
		fmt.Printf("output:\n%v\n%v\n", isInit, hex.EncodeToString(initalizerBytes))
	}

	callData, err := GetCallBytes(parsedJSON, "depositAndLock", assetAddress, value)
	if err != nil {
		return nil, err
	}

	extendTime := big.NewInt(3600)
	chainId := big.NewInt(11155111)

	lockHash := EncodeAndHash(extendTime, assetAddress, extendNonce, chainId)

	// current nonce
	//

	// ned to make lock funds
	// depositAndLock(address asset_, uint256 amount_)
	// Escrow(payable(escrowAddress_)).depositAndLock{value: 5 ether}(address(0), 0);
	// query extendNonce()
	// (uint8 v, bytes32 r, bytes32 s) = vm.sign(signerPk, keccak256(abi.encode(3600, address(0), extendNonce, block.chainid)).toEthSignedMessageHash());
	// Escrow(payable(escrowAddress_)).extendLock(3600, address(0), abi.encodePacked(r, s, bytes1(v)));
	// console.log("resulting user asset info");
	// (uint256 a, uint256 b, uint256 c) = Escrow(payable(escrowAddress_)).getAssetInfo(address(0));

	messageEscrowEvm.Init = EscrowInitRaw{
		SingletonAddress: escrowSingletonAddress.Hex(),
		FactoryAddress:   escrowFactoryAddress.Hex(),
		Salt:             hex.EncodeToString(salt),
		IsInitialized:    isInit,
		EscrowAddress:    hex.EncodeToString(escrowAddressBytes),
		Initalizer:       hex.EncodeToString(initalizerBytes),
		Payload:          hex.EncodeToString(initalizerBytes),
	}

	messageEscrowEvm.DepositAndLock = EscrowDepositAndLockRaw{
		AssetAddress:  assetAddress.Hex(),
		AssetValue:    value.String(),
		AssetAmount:   assetAmount.String(),
		AssetLocked:   assetAmountLocked.String(),
		AssetDeadline: deadline.String(),
		EscrowAddress: hex.EncodeToString(escrowAddressBytes),
		Payload:       hex.EncodeToString(callData),
	}

	messageEscrowEvm.TimeLockHash = EscrowTimeLockHashRaw{
		ExtendTime:   extendTime.String(),
		AssetAddress: assetAddress.Hex(),
		ExtendNonce:  extendNonce.String(),
		ChainId:      chainId.String(),
		Hash:         hex.EncodeToString(lockHash),
	}

	return messageEscrowEvm, nil
}

/*
type UnsignedEscrowRequestParams struct {
	Header utils.PartialHeader `query:"header"`
	Amount string              `query:"amount"` // gwei
}

http://localhost:8080/api/main?query=
unsigned-message&
txtype=1&
fid=11155111&
fsigner=0x19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A&
tid=1667471769&
tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&
payload=00&
target=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf
value=20000000
*/

// func CalculateSelector(input string) [4]byte {
// 	// methodSignature := "getEscrowAddress(bytes,bytes32)"
// 	hasher := sha3.NewLegacyKeccak256()
// 	hasher.Write([]byte(input))
// 	return hasher.Sum(nil)[:4]
// }

func CalculateSelector(input string) [4]byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(input))
	sum := hasher.Sum(nil)
	return [4]byte{sum[0], sum[1], sum[2], sum[3]}
}

func PackArgs(methodID []byte, methodInputs abi.Arguments, args ...interface{}) ([]byte, error) {
	packedArgs, err := methodInputs.Pack(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments: %v", err)
	}

	return append(methodID, packedArgs...), nil
}

func ViewFunction(client *ethclient.Client, contractAddress common.Address, parsedABI abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{To: &contractAddress, Data: data}
	result, err := client.CallContract(context.Background(), callMsg, big.NewInt(305965178))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetCallBytes(parsedABI abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	isArgsEmpty := func(args []interface{}) bool {
		if len(args) == 0 {
			return true
		}

		for _, arg := range args {
			if arg != nil {
				return false
			}
		}
		return true
	}

	var data []byte
	var err error
	if !isArgsEmpty(args) {
		data, err = parsedABI.Pack(methodName, args...)
	} else {
		data, err = parsedABI.Pack(methodName)
	}

	return data, err
}

func createCall(parsedABI abi.ABI, contractAddress common.Address, methodName string, params ...interface{}) (Call, error) {
	callData, err := GetCallBytes(parsedABI, methodName, params...)
	if err != nil {
		fmt.Print("uncaught error")
		return Call{}, fmt.Errorf("bytes for call %v failed: %v", methodName, err.Error())
	}
	fmt.Printf("\ninternal calldata: \n%v\n", callData)

	return Call{
		Target:   contractAddress,
		CallData: callData,
	}, nil
}

func MulticallView(client *ethclient.Client, multicallAddress common.Address, calls []Calls) ([]MulticallResult, error) {
	var multicallViewInput []Call
	fmt.Print("\ngot multicall far0\n")
	for _, call := range calls {
		c, err := createCall(call.abi, call.contractAddress, call.method, call.params)
		if err != nil {
			return nil, fmt.Errorf("failed to create call: %v", err)
		}

		multicallViewInput = append(multicallViewInput, c)
	}

	parsedJSON, _ := abi.JSON(strings.NewReader(contractAbiMulticall))
	returnData, err := ViewFunction(client, multicallAddress, parsedJSON, "multicallView", multicallViewInput)
	if err != nil {
		return nil, fmt.Errorf("failed to execute multicallView: %v", err)
	}

	data, err := parsedJSON.Unpack("multicallView", returnData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack multicallView result: %v", err)
	}

	var results []MulticallResult
	for _, v := range data {
		for _, vv := range v.([]struct {
			Success    bool   "json:\"success\""
			ReturnData []byte "json:\"returnData\""
		}) {
			results = append(results, MulticallResult{
				Success:    vv.Success,
				ReturnData: vv.ReturnData,
			})
		}
	}

	return results, nil
}

/*
Returns:

- Escrow address as []byte: The derived escrow contract address.

- Initializer code as []byte: The encoded initializer data for the escrow contract.

- err: An error if the computation fails.
*/
func GetEscrowAddress(
	client *ethclient.Client,
	signer common.Address,
	escrowFactoryAddress common.Address,
	escrowSingletonAddress common.Address,
	salt []byte) ([]byte, []byte, error) {
	// from Escrow ABI
	parsedJSON, _ := abi.JSON(strings.NewReader(`[{
		"inputs": [
			{"internalType": "bytes", "name": "_initializer", "type": "bytes"},
			{"internalType": "bytes32", "name": "_salt", "type": "bytes32"}
		],
		"name": "getEscrowAddress",
		"outputs": [{"internalType": "address", "name": "proxy", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	},{
		"inputs": [
			{"name":"owner_","type":"address","internalType":"address"},
			{"name":"delegateAddress_","type":"address","internalType":"address"}
		],
		"name":"initialize",
		"outputs":[],
		"stateMutability":"nonpayable",
		"type":"function"
	}]`))

	initializerBytes, err := GetCallBytes(parsedJSON, "initialize", signer, escrowSingletonAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("initializerBytes generation failed: %v", err.Error())
	}

	escrowAddress, err := ViewFunction(
		client,
		escrowFactoryAddress,
		parsedJSON,
		"getEscrowAddress", initializerBytes, utils.Bytes32PadLeft(salt))
	if err != nil {
		return nil, nil, fmt.Errorf("failed getEscrowAddress request: %v\n", err)
	}

	return escrowAddress, initializerBytes, nil
}

func ExtCodeSize(client *ethclient.Client, address common.Address) (int, error) {
	ctx := context.Background()
	code, err := client.CodeAt(ctx, address, nil) // nil block number for the latest state
	if err != nil {
		return 0, fmt.Errorf("geth client failed to get extcodesize: %v", err)
	}
	return len(code), nil
}

func ToEthSignedMessageHash(input []byte) []byte {
	return crypto.Keccak256(append(utils.EthDomainHeader, input...))
}

// EncodeAndHash encodes and hashes the input data
func EncodeAndHash(extendTime *big.Int, assetAddress common.Address, extendNonce *big.Int, chainID *big.Int) []byte {
	var bytes_ []byte
	var padded_ [32]byte
	padded_ = utils.Bytes32PadLeft(extendTime.Bytes())
	bytes_ = append(bytes_, padded_[:]...)
	padded_ = utils.Bytes32PadLeft(assetAddress.Bytes())
	bytes_ = append(bytes_, padded_[:]...)
	padded_ = utils.Bytes32PadLeft(extendNonce.Bytes())
	bytes_ = append(bytes_, padded_[:]...)
	padded_ = utils.Bytes32PadLeft(chainID.Bytes())
	bytes_ = append(bytes_, padded_[:]...)

	return ToEthSignedMessageHash(crypto.Keccak256(bytes_))
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

	// privateKey, relayAddress, err := utils.EnvKey2Ecdsa()
	// if err != nil {
	// 	return nil, err
	// }

	privateKey, relayAddress, err := utils.Key2Ecdsa(os.Getenv("HYPERLIQUID_PRIVATE_KEY"))
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

	reciept, err := bind.WaitMined(context.Background(), &client, signedTx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("tx receipt: \n%v", reciept)

	var returnedData []byte
	for _, log := range reciept.Logs {
		if len(log.Data) > 0 {
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
