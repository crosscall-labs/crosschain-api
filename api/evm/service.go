package evmHandler

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/crosscall-labs/crosschain-api/pkg/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func AssetMintRequest(r *http.Request, parameters ...*utils.AssetMintRequestParams) (interface{}, error) {
	fmt.Print("I got here")
	var params *utils.AssetMintRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &utils.AssetMintRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, err
		}
	}

	fmt.Print("our params data:\n")
	utils.PrintStructFields(params)

	if !common.IsHexAddress(params.UserAddress) {
		return nil, fmt.Errorf("escrow invalid Ethereum address")
	}

	if !common.IsHexAddress(params.AssetAddress) {
		return nil, fmt.Errorf("escrow invalid Ethereum address")
	}

	bigInt := new(big.Int)
	_, success := bigInt.SetString(params.AssetAmount, 10) // Base 10 for decimal numbers

	if !success {
		return nil, fmt.Errorf("ailed to parse the string into *big.Int")
	}

	jsonrpc, _ := getChainRpc(params.ChainId)
	client, err := ethclient.Dial(jsonrpc)
	if err != nil {
		fmt.Printf("\nclient connection failed: %v\n", err)
		return nil, fmt.Errorf("client connection failed: %v", err)
	}

	userAddress := common.HexToAddress(params.UserAddress)
	assetAddress := common.HexToAddress(params.AssetAddress)

	parsedFaucetABI, _ := abi.JSON(strings.NewReader(`[{
		"type":"function",
		"name":"mint",
		"inputs":[
			{"name":"to","type":"address","internalType":"address"},
			{"name":"amount","type":"uint256","internalType":"uint256"}
		],
		"outputs":[],
		"stateMutability":"nonpayable"
	}]`))

	fmt.Print("\ngot here")

	return ExecuteFunction(*client, assetAddress, parsedFaucetABI, "mint", common.Big0, userAddress, bigInt)
}

func AssetInfoRequest(r *http.Request, parameters ...*utils.AssetInfoRequestParams) (interface{}, error) {
	var params *utils.AssetInfoRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &utils.AssetInfoRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
	}

	var err error
	response := &utils.AssetInfoRequestResponse{}
	var calls []Calls
	var results []MulticallResult
	response.ChainId = params.ChainId
	response.VM = params.VM

	fmt.Printf("\nparams fields: \n")
	utils.PrintStructFields(params)

	if !common.IsHexAddress(params.UserAddress) {
		return nil, utils.ErrInternal(fmt.Errorf("escrow invalid Ethereum address").Error())
	}

	userAddress := common.HexToAddress(params.UserAddress)

	jsonrpc, _ := getChainRpc(params.ChainId)
	client, err := ethclient.Dial(jsonrpc)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Errorf("client connection failed: %v", err).Error())
	}

	var multicallAddress common.Address
	multicallAddress, err = getMulticallAddress(params.ChainId) ///////////////////// need to create a func to deternibne multicall
	fmt.Printf("\nmulticall address: %v", params.ChainId)
	fmt.Printf("\nmulticall address: %v", multicallAddress)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	parsedMulticallABI, _ := abi.JSON(strings.NewReader(`[{
			"type":"function",
			"name":"getExtcodesize",
			"inputs":[{"name":"address_","type":"address","internalType":"address"}],
			"outputs":[{"name":"size_","type":"uint256","internalType":"uint256"}],
			"stateMutability":"view"
		},
		{
			"type":"function",
			"name":"multicallView",
			"inputs":[{
				"name":"calls",
				"type":"tuple[]",
				"internalType":"struct Multicall.Call[]",
				"components":[
					{"name":"target","type":"address","internalType":"address"},
					{"name":"callData","type":"bytes","internalType":"bytes"}
			]}],
			"outputs":[{
				"name":"",
				"type":"tuple[]",
				"internalType":"struct Multicall.Result[]",
				"components":[
					{"name":"success","type":"bool","internalType":"bool"},
					{"name":"returnData","type":"bytes","internalType":"bytes"}
			]}],
			"stateMutability":"view"
	}]`))

	assetAddress := common.HexToAddress(params.AssetAddress)

	parsedErc20ABI, _ := abi.JSON(strings.NewReader(`[{
			"type": "function",
			"name": "name",
			"inputs": [],
			"outputs": [{"name":"","type":"string","internalType":"string"}],
			"stateMutability":"view"
		},
  	{
			"type": "function",
			"name": "symbol",
			"inputs": [],
			"outputs": [{"name":"","type":"string","internalType":"string"}],
			"stateMutability":"view"
		},
		{
			"type": "function",
			"name": "decimals",
			"inputs": [],
			"outputs": [{"name":"","type": "uint8","internalType":"uint8"}],
			"stateMutability":"view"
		},
		{
			"type": "function",
			"name": "totalSupply",
			"inputs": [],
			"outputs": [{"name":"","type": "uint256","internalType":"uint256"}],
			"stateMutability":"view"
		},
		{
			"type": "function",
			"name": "balanceOf",
			"inputs": [{"name":"account","type": "address","internalType":"address"}],
			"outputs": [{"name":"","type": "uint256","internalType":"uint256"}],
			"stateMutability":"view"
	}]`))

	calls = []Calls{
		{contractAddress: assetAddress, abi: parsedErc20ABI, method: "name", params: nil},
		{contractAddress: assetAddress, abi: parsedErc20ABI, method: "symbol", params: nil},
		{contractAddress: assetAddress, abi: parsedErc20ABI, method: "decimals", params: nil},
		{contractAddress: assetAddress, abi: parsedErc20ABI, method: "totalSupply", params: nil},
		{contractAddress: assetAddress, abi: parsedErc20ABI, method: "balanceOf", params: userAddress},
	}

	results, err = MulticallView(client, multicallAddress, calls)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Errorf("multicall view failed: %v", err).Error())
	}

	assetName, _ := parsedErc20ABI.Unpack("name", results[0].ReturnData)
	assetSymbol, _ := parsedErc20ABI.Unpack("symbol", results[1].ReturnData)
	assetDecimals, _ := parsedErc20ABI.Unpack("decimals", results[2].ReturnData)
	assetTotalSupply, _ := parsedErc20ABI.Unpack("totalSupply", results[3].ReturnData)
	userBalance, _ := parsedErc20ABI.Unpack("balanceOf", results[4].ReturnData)

	fmt.Print("\ngot this far2\n")

	response.Asset = struct {
		Address     string `json:"address"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Decimal     string `json:"decimal"`
		TotalSupply string `json:"total-supply"`
		Supply      string `json:"supply"`
		Description string `json:"description"`
	}{
		Address:     assetAddress.Hex(),
		Name:        assetName[0].(string),
		Symbol:      assetSymbol[0].(string),
		Decimal:     strconv.Itoa(int(assetDecimals[0].(uint8))),
		TotalSupply: assetTotalSupply[0].(*big.Int).String(),
		Supply:      "",
		Description: "",
	}
	response.User = struct {
		Balance string "json:\"balance\""
	}{
		Balance: userBalance[0].(*big.Int).String(),
	}

	if params.EscrowAddress != "" {
		if !common.IsHexAddress(params.EscrowAddress) {
			return nil, fmt.Errorf("escrow invalid Ethereum address")
		}

		escrowAddress := common.HexToAddress(params.EscrowAddress)

		parsedEscrowABI, _ := abi.JSON(strings.NewReader(`[{
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
		}]`))

		calls = []Calls{
			{contractAddress: multicallAddress, abi: parsedMulticallABI, method: "getExtcodesize", params: escrowAddress},
			{contractAddress: escrowAddress, abi: parsedEscrowABI, method: "getAssetInfo", params: assetAddress},
		}

		results, err = MulticallView(client, multicallAddress, calls)
		if err != nil {
			return nil, utils.ErrInternal(fmt.Errorf("multicall view failed: %v", err).Error())
		}

		escrowSize, _ := parsedMulticallABI.Unpack("getExtcodesize", results[0].ReturnData)
		escrowInfo, _ := parsedEscrowABI.Unpack("getAssetInfo", results[1].ReturnData)

		response.Escrow = struct {
			Init         bool   `json:"init"`
			Balance      string `json:"balance"`
			LockBalance  string `json:"lock-balance"`
			LockDeadline string `json:"lock-deadline"`
		}{
			Init:         escrowSize[0].(*big.Int).Int64() > 0,
			Balance:      escrowInfo[0].(*big.Int).String(),
			LockBalance:  escrowInfo[1].(*big.Int).String(),
			LockDeadline: escrowInfo[2].(*big.Int).String(),
		}
	}

	if params.AccountAddress != "" {
		if !common.IsHexAddress(params.AccountAddress) {
			return nil, utils.ErrInternal(fmt.Errorf("account invalid Ethereum address").Error())
		}

		accountAddress := common.HexToAddress(params.AccountAddress)

		calls = []Calls{
			{contractAddress: multicallAddress, abi: parsedMulticallABI, method: "getExtcodesize", params: accountAddress},
			{contractAddress: assetAddress, abi: parsedErc20ABI, method: "balanceOf", params: accountAddress},
		}

		results, err = MulticallView(client, multicallAddress, calls)
		if err != nil {
			return nil, utils.ErrInternal(fmt.Errorf("multicall view failed: %v", err).Error())
		}

		accountSize, _ := parsedMulticallABI.Unpack("getExtcodesize", results[0].ReturnData)
		accountInfo, _ := parsedErc20ABI.Unpack("balanceOf", results[1].ReturnData)

		response.Account = struct {
			Init    bool   `json:"init"`
			Balance string `json:"balance"`
		}{
			Init:    accountSize[0].(*big.Int).Int64() > 0,
			Balance: accountInfo[0].(*big.Int).String(),
		}
	}

	return response, nil
}

// It accepts an optional query parameter for internal calls.
func UnsignedEscrowRequest(r *http.Request, parameters ...*UnsignedEscrowRequestParams) (interface{}, error) {
	var params *UnsignedEscrowRequestParams
	//var err Error

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &UnsignedEscrowRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
	}

	var errorStr string
	params.Header.ChainId, params.Header.ChainType, params.Header.ChainName, errorStr = utils.CheckChainPartialType(params.Header.ChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}

	salt := common.Hex2Bytes("0x0000000000000000000000000000000000000000000000000000000000000037")
	signer := common.HexToAddress("19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A") // should be from params
	client, err := ethclient.Dial("https://rpc2.sepolia.org")                 // should be from inputs but ignored
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
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
		utils.ErrInternal(err.Error())
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
			return nil, utils.ErrInternal(fmt.Errorf("failed extendNonce call: %v", err).Error())
		}

		parsedResults, err := parsedJSON.Unpack("extendNonce", response)
		if err != nil {
			return nil, utils.ErrInternal(fmt.Errorf("failed extendNonce parse: %v", err).Error())
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

func UnsignedEntryPointRequest(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	var params *UnsignedEntryPointRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &UnsignedEntryPointRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, err
		}
	}

	var errorStr string
	params.Header.FromChainId, params.Header.FromChainType, params.Header.FromChainName, errorStr = utils.CheckChainPartialType(params.Header.FromChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}
	params.Header.ToChainId, params.Header.ToChainType, params.Header.ToChainName, errorStr = utils.CheckChainPartialType(params.Header.ToChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}

	// payload, err := utils.Str2Bytes(params.Payload)
	// if err != nil {
	// 	utils.ErrMalformedRequest(w, err.Error())
	// }
	// fmt.Print(payload)

	// return data; priceGwei is redundant but left up to the user if the user wants to input a different escrow payout in the paymaster and data
	// type MessageOpEvm struct {
	// 	UserOp           PackedUserOperationResponse `json:"op-packed-data"` // parsed data, recommended to validate data
	// 	PaymasterAndData PaymasterAndData            `json:"op-paymaster"`
	// 	UserOpHash       string                      `json:"op-hash"`
	// 	PriceGwei        string                      `json:"op-price"`
	// }

	// need to create the userop, but to make it real we need to use proper tools to value the gas estimate and over estimate
	//		for now use fixed values from forge tests

	// todo
	//	create default values for calldata (this should be done by the protocol api since we don't want to delegate using a specifc wallet architecture)
	//		test data will be using an empty value sent as if it were thorugh signer -> simpleAccount proxy
	//	combine the transaction gas and cost for execution then multiply by 0.1%, this should be our crosschain fee + bid fee
	// 		add this value to the paymaster and data AND PriceGwei

	packedUserOperation := GenerateTestPackedUserOperation()
	paymasterAndData := PaymasterAndData{}

	// empty data for basic testing
	packedUserOperationResponse, _ := ToPackedUserOperationResponse(packedUserOperation)
	paymasterAndDataResponse, _ := ToPaymasterAndDataResponse(paymasterAndData)
	return MessageOpEvm{
		UserOp:           packedUserOperationResponse,
		PaymasterAndData: paymasterAndDataResponse,
		UserOpHash:       "0x0000000000000000000000000000000000000000000000000000000000000000",
		PriceGwei:        "0",
	}, nil
}

// normally this is generated by the wallet
// our client will verify the gas
func GenerateTestPackedUserOperation() PackedUserOperation {
	return PackedUserOperation{
		Sender:             common.Address{},
		Nonce:              big.NewInt(0),
		InitCode:           []byte{},
		CallData:           []byte{},
		AccountGasLimits:   [32]byte{},
		PreVerificationGas: big.NewInt(20000000),
		GasFees:            [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		PaymasterAndData:   []byte{},
		Signature:          []byte{},
	}
}

// ToPackedUserOperationResponse converts a PackedUserOperation to PackedUserOperationResponse.
//
// Needs to be tested
func ToPackedUserOperationResponse(packedUserOperation PackedUserOperation) (PackedUserOperationResponse, error) {
	return PackedUserOperationResponse{
		Sender:             utils.ToHexAddress(packedUserOperation.Sender),
		Nonce:              packedUserOperation.Nonce.String(),
		InitCode:           utils.ToHexBytes(packedUserOperation.InitCode),
		CallData:           utils.ToHexBytes(packedUserOperation.CallData),
		AccountGasLimits:   utils.ToHexBytes(packedUserOperation.AccountGasLimits[:]),
		PreVerificationGas: packedUserOperation.PreVerificationGas.String(),
		GasFees:            utils.ToHexBytes(packedUserOperation.GasFees[:]),
		PaymasterAndData:   utils.ToHexBytes(packedUserOperation.PaymasterAndData),
		Signature:          utils.ToHexBytes(packedUserOperation.Signature),
	}, nil
}

// ToPaymasterAndDataResponse converts a PaymasterAndData to PaymasterAndDataResponse.
//
// Needs to be tested
func ToPaymasterAndDataResponse(pad PaymasterAndData) (PaymasterAndDataResponse, error) {
	return PaymasterAndDataResponse{
		Paymaster:                     utils.ToHexAddress(pad.Paymaster),
		PaymasterVerificationGasLimit: utils.ToHexBytes(pad.PaymasterVerificationGasLimit[:]),
		PaymasterPostOpGasLimit:       utils.ToHexBytes(pad.PaymasterPostOpGasLimit[:]),
		Signer:                        utils.ToHexAddress(pad.Signer),
		DestinationDomain:             utils.Uint32ToString(pad.DestinationDomain),
		MessageType:                   utils.Uint8ToString(pad.MessageType),
		AssetAddress:                  utils.ToHexAddress(pad.AssetAddress),
		AssetAmount:                   pad.AssetAmount.String(),
	}, nil
}

// FromPackedUserOperationResponse converts a PackedUserOperationResponse to PackedUserOperation.
//
// Needs to be written.
func FromPackedUserOperationResponse(packedUserOperationResponse PackedUserOperationResponse) (PackedUserOperation, error) {
	return PackedUserOperation{}, nil
}

// FromPaymasterAndDataResponse converts a PaymasterAndDataResponse to PaymasterAndData.
//
// Needs to be written.
func FromPaymasterAndDataResponse(pad PaymasterAndDataResponse) (PaymasterAndData, error) {
	return PaymasterAndData{}, nil
}

// don't need to gen PaymasterAndData{} suffices
// func GenerateTestPaymasterAndData(paymasterAddress common.Address) PaymasterAndData {
// 	// type PaymasterAndData struct {
// 	// 	Paymaster                     common.Address
// 	// 	PaymasterVerificationGasLimit [32]byte
// 	// 	PaymasterPostOpGasLimit       [32]byte
// 	// 	Signer                        common.Address
// 	// 	DestinationDomain             [4]byte
// 	// 	MessageType                   byte
// 	// 	AssetAddress                  common.Address
// 	// 	AssetAmount                   *big.Int
// 	// }

// 	// type PaymasterAndDataResponse struct {
// 	// 	Paymaster                     string `json:"pad-paymaster"`
// 	// 	PaymasterVerificationGasLimit string `json:"pad-verification-gas-limit"`
// 	// 	PaymasterPostOpGasLimit       string `json:"pad-post-op-gas-limit"`
// 	// 	Signer                        string `json:"pad-signer"`
// 	// 	DestinationDomain             string `json:"pad-destination-domain"`
// 	// 	MessageType                   string `json:"pad-message-type"`
// 	// 	AssetAddress                  string `json:"pad-asset-address"`
// 	// 	AssetAmount                   string `json:"pad-asset-amount"`
// 	// }
// 	return PaymasterAndData{
// 		Paymaster                     common.Address{}
// 		PaymasterVerificationGasLimit [32]byte
// 		PaymasterPostOpGasLimit       [32]byte
// 		Signer                        common.Address
// 		DestinationDomain             [4]byte
// 		MessageType                   byte
// 		AssetAddress                  common.Address
// 		AssetAmount                   *big.Int

// 		Sender:             common.Address{},
// 		Nonce:              big.NewInt(0),
// 		InitCode:           []byte{},
// 		CallData:           []byte{},
// 		AccountGasLimits:   [32]byte{},
// 		PreVerificationGas: big.NewInt(20000000),
// 		GasFees:            [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
// 		PaymasterAndData:   []byte{},
// 		Signature:          []byte{},
// 	}

// var paymasterAndData []byte
// paymasterPrefix := append(common.FromHex(paymasterAddress), common.FromHex("0x0000000000000000000000000098968000000000000000000000000000989680")...)
// paymasterSigner := common.FromHex(signer)
// someint, err = strconv.ParseInt(originId, 10, 64)
// if err != nil {
// 	fmt.Println(err)
// 	utils.ErrInternal(w)
// 	return
// }
// paymasterOrigin := padLeftHex(int(someint))
// paymasterAsset := common.FromHex(assetAddress)
// someint, err = strconv.ParseInt(assetAmount, 10, 64)
// if err != nil {
// 	fmt.Println(err)
// 	utils.ErrInternal(w)
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
// }
