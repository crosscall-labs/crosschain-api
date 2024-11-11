package tvmHandler

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func InitClient() (context.Context, ton.APIClientWrapped, *wallet.Wallet, error) {
	mnemonic := os.Getenv("TON_BACKEND_WALLET_MNEMONIC")
	ctx, api, err := ConnectToTestnetClient()
	if err != nil {
		return nil, nil, nil, utils.ErrInternal(fmt.Sprintf("Failed to connect to client: %s", err.Error()))
	}
	backendWallet, err := wallet.FromSeed(api, strings.Split(mnemonic, " "), wallet.V3R2)
	if err != nil {
		return nil, nil, nil, utils.ErrInternal(fmt.Sprintf("FromSeed err: %s", err.Error()))
	}

	return ctx, api, backendWallet, nil
}

func UnsignedEscrowRequest(r *http.Request, parameters ...*UnsignedEscrowRequestParams) (interface{}, error) {
	// var params *UnsignedEscrowRequestParams
	// //var err Error

	// if len(parameters) > 0 {
	// 	params = parameters[0]
	// } else {
	// 	params = &UnsignedEscrowRequestParams{}
	// }

	// if r != nil {
	// 	if err := utils.ParseAndValidateParams(r, &params); err != nil {
	// 		return nil, err
	// 	}
	// }

	// var errorStr string
	// params.Header.ChainId, params.Header.ChainType, params.Header.ChainName, errorStr = utils.CheckChainPartialType(params.Header.ChainId, "escrow", params.Header.TxType)
	// if errorStr != "" {
	// 	return nil, utils.ErrMalformedRequest(errorStr)
	// }

	return MessageEscrowTvm{}, nil
}

func UnsignedEntryPointRequest(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	var params *UnsignedEntryPointRequestParams
	// var err error

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

	// response := MessageOpTvm{}

	// response.Header = params.Header

	ownerEvmAddressBytes, err := utils.HexToBytes(params.ProxyParams.ProxyHeader.OwnerEvmAddress)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("Unable to parse owner evm address: %v", err.Error()))
	}
	ownerEvmAddress := binary.BigEndian.Uint64(ownerEvmAddressBytes)
	fmt.Print(ownerEvmAddress)

	ownerTvmAddress, err := address.ParseAddr(params.ProxyParams.ProxyHeader.OwnerTvmAddress)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("Unable to parse owner tvm address: %v", err.Error()))
	}
	fmt.Print(ownerTvmAddress)

	workChain, err := strconv.Atoi(params.ProxyParams.WorkChain)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("Unable to parse workchain: %v", err.Error()))
	}
	fmt.Print(workChain)

	/// dumps nothing
	proxyAddress, err := CalculateWallet(ownerEvmAddress, ownerTvmAddress, entryPointAddress, workChain)
	if err != nil {
		return nil, err
	}
	fmt.Print(proxyAddress)

	// response.ProxyAddress = proxyAddress.String()

	// func CalculateWallet(
	// 	evmAddress uint64,
	// 	tvmAddress *address.Address,
	// 	entrypointAddress *address.Address,
	// 	workchain int,
	// )

	// params.Header.ToChainId, params.Header.ToChainType, params.Header.ToChainName, errorStr = utils.CheckChainPartialType(params.Header.ToChainId, "escrow", params.Header.TxType)
	// if errorStr != "" {
	// 	return nil, utils.ErrMalformedRequest(errorStr)
	// }

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

	// packedUserOperation := GenerateTestTvmOperation()
	// paymasterAndData := PaymasterAndData{}

	// // empty data for basic testing
	// packedUserOperationResponse, _ := ToPackedUserOperationResponse(packedUserOperation)
	// paymasterAndDataResponse, _ := ToPaymasterAndDataResponse(paymasterAndData)
	return MessageOpTvm{
		Header: params.Header,
		ProxyParams: ProxyParams{
			ProxyHeader: ProxyHeaderParams{
				Nonce:           "0",
				EntryPoint:      entryPointAddress.Dump(),
				PayeeAddress:    "",
				OwnerEvmAddress: params.ProxyParams.ProxyHeader.OwnerEvmAddress,
				OwnerTvmAddress: params.ProxyParams.ProxyHeader.OwnerTvmAddress,
			},
			ExecutionData: ExecutionDataParams{
				Regime:      "",
				Destination: "",
				Value:       "",
				Body:        "",
			},
			WithProxyInit:   "true",
			ProxyWalletCode: "",
			WorkChain:       params.ProxyParams.WorkChain,
		},
		ProxyAddress: "proxyAddress.Dump()",
	}, nil
	// return &MessageOpTvm{}, nil
}

// ToPackedUserOperationResponse converts a PackedUserOperation to PackedUserOperationResponse.
//
// Needs to be tested
// func ToTestTvmOperationResponse(packedUserOperation PackedUserOperation) (PackedUserOperationResponse, error) {
// 	return PackedUserOperationResponse{
// 		Sender:             utils.ToHexAddress(packedUserOperation.Sender),
// 		Nonce:              packedUserOperation.Nonce.String(),
// 		InitCode:           utils.ToHexBytes(packedUserOperation.InitCode),
// 		CallData:           utils.ToHexBytes(packedUserOperation.CallData),
// 		AccountGasLimits:   utils.ToHexBytes(packedUserOperation.AccountGasLimits[:]),
// 		PreVerificationGas: packedUserOperation.PreVerificationGas.String(),
// 		GasFees:            utils.ToHexBytes(packedUserOperation.GasFees[:]),
// 		PaymasterAndData:   utils.ToHexBytes(packedUserOperation.PaymasterAndData),
// 		Signature:          utils.ToHexBytes(packedUserOperation.Signature),
// 	}, nil
// }
