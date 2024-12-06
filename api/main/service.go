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
	evmHandler "github.com/laminafinance/crosschain-api/api/evm"
	tvmHandler "github.com/laminafinance/crosschain-api/api/tvm"
	"github.com/laminafinance/crosschain-api/pkg/utils"
	"golang.org/x/crypto/sha3"
)

func UnsignedRequest(r *http.Request) (interface{}, error) {
	params := &UnsignedRequestParams{}

	if err := utils.ParseAndValidateParams(r, params); err != nil {
		return nil, err
	}

	// payload, err := utils.Str2Bytes(params.Payload)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Print(payload)

	// finish validating input parameters
	// func checkChainType(chainId string) (string, string, []int, []int, error) { // out: vm, name, entrypointType, escrowType, error
	var errorStr string
	params.Header.FromChainId, params.Header.FromChainType, params.Header.FromChainName, errorStr = utils.CheckChainPartialType(params.Header.FromChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}
	params.Header.ToChainId, params.Header.ToChainType, params.Header.ToChainName, errorStr = utils.CheckChainPartialType(params.Header.ToChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}

	//utils.PrintStructFields(params)

	unsignedDataResponse := &UnsignedDataResponse{}
	unsignedDataResponse.Header = params.Header
	// var toResponse interface{}
	// var fromResponse interface{}
	// var err error
	fmt.Print("I got this far\n")
	switch params.Header.ToChainType {
	case "evm":
		fmt.Print("I got this far2\n")
		response, err := evmHandler.UnsignedEntryPointRequest(nil, &evmHandler.UnsignedEntryPointRequestParams{
			Header:  params.Header,
			Payload: params.Payload,
		})
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
		unsignedDataResponse.ToMessage = response.(MessageResponse)
		// utils.PrintStructFields(response)
	case "tvm":
		response, err := tvmHandler.UnsignedEntryPointRequest(nil, &tvmHandler.UnsignedEntryPointRequestParams{
			Header: params.Header,
			ProxyParams: tvmHandler.ProxyParams{
				ProxyHeader: tvmHandler.ProxyHeaderParams{
					OwnerEvmAddress: params.Header.FromChainSigner, // we cannot manage tvm<>tvm txs
					OwnerTvmAddress: params.Header.ToChainSigner,
				},
				ExecutionData: tvmHandler.ExecutionDataParams{
					Destination: params.Target,
					Value:       params.Value,
					Body:        params.Payload,
				},
				WithProxyInit: "true", // we should be checking if init
				WorkChain:     "-1",   // this should be set but testnet default
			},
		})
		/*
			// now we need to build our rest api request
			// tvmL
			// http://localhost:8080/api/tvm?
			query=unsigned-entrypoint-request&
			txtype=1&fid=11155111&
			fsigner=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&
			tid=1667471769&
			tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&
			p-init=false&
			p-workchain=-1&
			p-evm=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&
			p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&
			exe-target=EQAW3iupIDrCICc7SbcY_SBP6jCNO-F8v91dG9XNLHw-lE9k&
			exe-value=200000000&
			exe-body=
			// main so far:
			// http://localhost:8080/api/main?
			query=unsigned-message&
			txtype=1&
			fid=11155111&
			fsigner=0x19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A&
			tid=1667471769&
			tsigner=0x19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A&
			payload=00
		*/
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
		unsignedDataResponse.ToMessage = response.(MessageResponse)
		/*
			type ProxyParams struct {
				ProxyHeader     ProxyHeaderParams   `query:"p-header"`
				ExecutionData   ExecutionDataParams `query:"p-exe"`
				WithProxyInit   string              `query:"p-init"` // Required: Initalize the proxy wallet
				ProxyWalletCode string              `query:"p-code" optional:"true"`
				WorkChain       string              `query:"p-workchain" optional:"true"` // assume 0 for testnet atm
			}

			type ProxyHeaderParams struct {
				Nonce           string `query:"p-nonce" optional:"true"`
				EntryPoint      string `query:"p-entrypoint" optional:"true"` // possible that a better one is accepted in the future
				PayeeAddress    string `query:"p-payee" optional:"true"`      // solver is us for now
				OwnerEvmAddress string `query:"p-evm"`                        // easy to derive
				OwnerTvmAddress string `query:"p-tvm" optional:"true"`        // our social login SHOULD generate this
			}

			type ExecutionDataParams struct {
				Regime      string `query:"exe-regime" optional:"true"`
				Destination string `query:"exe-target" optional:"true"`
				Value       string `query:"exe-value" optional:"true"`
				Body        string `query:"exe-body" optional:"true"`
			}


						// test tx
						http://localhost:8080/api/tvm?query=unsigned-entrypoint-request
						&txtype=1
						&fid=11155111
						&fsigner=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266
						&tid=1667471769
						&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf

						&p-init=false
						&p-workchain=-1
						&p-evm=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266
						&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf
						&exe-target=EQAW3iupIDrCICc7SbcY_SBP6jCNO-F8v91dG9XNLHw-lE9k
						&exe-value=200000000
						&exe-body=
						byte
						uint8
						evm 20 bytes


		*/
	case "svm":
	default:
		return nil, utils.ErrInternal(fmt.Sprintf("%s type chains are not yet supported", params.Header.FromChainType))
	}

	switch params.Header.FromChainType {
	case "evm":
		response, err := evmHandler.UnsignedEscrowRequest(nil, &evmHandler.UnsignedEscrowRequestParams{
			Header: utils.PartialHeader{
				TxType:      params.Header.TxType,
				ChainName:   params.Header.FromChainName,
				ChainType:   params.Header.FromChainType,
				ChainId:     params.Header.FromChainId,
				ChainSigner: params.Header.FromChainSigner,
			},
		})
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
		unsignedDataResponse.FromMessage = response.(MessageResponse)
		// utils.PrintStructFields(response)
	case "tvm":
	case "svm":
	default:
		return nil, utils.ErrInternal(fmt.Sprintf("%s type chains are not yet supported", params.Header.FromChainType))
	}

	return unsignedDataResponse, nil
}

func UnsignedBytecode(r *http.Request) (interface{}, error) {
	privateKey, relayAddress, err := utils.EnvKey2Ecdsa()
	fmt.Print(privateKey, relayAddress)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	params := &UnsignedBytecodeParams{}

	if err := utils.ParseAndValidateParams(r, params); err != nil {
		return nil, err
	}
	/*
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
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// if !results[0].Success {
		// 	fmt.Printf("Escrow: getEscrowAddress failed for chain chain %s\n", chainInfo.ChainId)
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// parsedResults, err := parsedABIs["EscrowFactory"].Unpack("getEscrowAddress", results[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// if !results2[0].Success {
		// 	fmt.Printf("SCW: getAddress failed for chain chain %s\n", chainInfo2.ChainId)
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// parsedResults2, err := parsedABIs["SimpleAccountFactory"].Unpack("getAddress", results2[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results2[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
		// 	return
		// }
		// extcodesize2 := parsedResults[0].(*big.Int)

		// parsedResults2, err = parsedABIs["Entrypoint"].Unpack("getNonce", results2[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }
		// unsignedDataResponse.EscrowPayload = "0x" + common.Bytes2Hex(escrowPayload)
		// unsignedDataResponse.EscrowTarget = chainInfo.AddressMulticall
		// unsignedDataResponse.EscrowValue = assetAmount // should gas and paymaster costs
		// fmt.Println("got this far8")
		// initcodeCall, err := GetViewCallBytes(*client2, parsedABIs["SimpleAccountFactory"], "createAccount", common.HexToAddress(signer), new(big.Int).SetBytes(SALT[:]))
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }
		// paymasterTarget := padLeftHex(someint)
		// paymasterAsset := common.FromHex(assetAddress) // need to check for badd addresses
		// someint, err = strconv.Atoi(assetAmount)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }
		// unsignedDataResponse.UserOpHash = "0x" + common.Bytes2Hex(returnData)

		// //w.WriteHeader(http.StatusOK)
		// w.Header().Set("Content-Type", "application/json")
		// if err := json.NewEncoder(w).Encode(unsignedDataResponse); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	*/
	return nil, nil
}

func SignedBytecode(r *http.Request) (interface{}, error) {
	privateKey, relayAddress, err := utils.EnvKey2Ecdsa()
	fmt.Print(privateKey, relayAddress)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	params := &SignedBytecodeParams{}

	if err := utils.ParseAndValidateParams(r, &params); err != nil {
		return nil, err
	}
	/*
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
		// 	utils.ErrInternal(w)
		// 	return
		// }
		// packedUserOperation.Nonce = big.NewInt(someint)
		// // parse perverificationgas to proper format
		// someint, err = strconv.ParseInt(useropPreVerificationGas, 10, 64)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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

		// initializerBytes, err := GetViewCallBytes(*client, parsedABIs["Escrow"], "initialize", common.HexToAddress(signer), common.HexToAddress(chainInfo.AddressEscrow))
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// if !results[0].Success {
		// 	fmt.Printf("Escrow: getEscrowAddress failed for chain chain %s\n", chainInfo.ChainId)
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// parsedResults, err := parsedABIs["EscrowFactory"].Unpack("getEscrowAddress", results[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results[0].ReturnData)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
		// 	return
		// }

		// someint, err = strconv.ParseInt(assetAmount, 10, 64)
		// if err != nil {
		// 	fmt.Println(err)
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
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
		// 	utils.ErrInternal(w)
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
	*/
	return nil, nil
}

func UnsignedEscrowPayout(r *http.Request) (interface{}, error) {
	privateKey, relayAddress, err := utils.EnvKey2Ecdsa()
	fmt.Print(privateKey, relayAddress)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	// params := &UnsignedEscrowPayoutParams{}

	// if !utils.ParseAndValidateParams(w, r, params) {
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
	return nil, nil
}

func SignedEscrowPayout(r *http.Request) (interface{}, error) {
	privateKey, relayAddress, err := utils.EnvKey2Ecdsa()
	fmt.Print(privateKey, relayAddress)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	params := &SignedEscrowPayoutParams{}

	if err := utils.ParseAndValidateParams(r, &params); err != nil {
		return nil, err
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
	return nil, nil
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
// 		utils.ErrInternal(w)
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
// 		utils.ErrInternal(w)
// 		return
// 	}

// 	if !results[0].Success {
// 		fmt.Printf("Escrow: getEscrowAddress failed for chain chain %s\n", chainInfo.ChainId)
// 		utils.ErrInternal(w)
// 		return
// 	}

// 	parsedResults, err := parsedABIs["EscrowFactory"].Unpack("getEscrowAddress", results[0].ReturnData)
// 	if err != nil {
// 		fmt.Println(err)
// 		utils.ErrInternal(w)
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
// 		utils.ErrInternal(w)
// 		return
// 	}

// 	parsedResults, err = parsedABIs["Multicall"].Unpack("getExtcodesize", results[0].ReturnData)
// 	if err != nil {
// 		fmt.Println(err)
// 		utils.ErrInternal(w)
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

func checkClient(chainId string) (*ethclient.Client, *Chain, error) {
	client, chainInfo, err := checkChainStatus(chainId)
	if err != nil {
		return nil, nil, err
	}
	if client == nil {
		return nil, nil, utils.ErrInternal(errUnsupportedChain(chainId))
	}
	return client, chainInfo, nil
}
