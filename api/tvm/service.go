package tvmHandler

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
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
	withProxyInit, err := strconv.ParseBool(params.ProxyParams.WithProxyInit)
	if err != nil {
		return nil, err
	}

	params.Header.FromChainId, params.Header.FromChainType, params.Header.FromChainName, errorStr = utils.CheckChainPartialType(params.Header.FromChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}
	params.Header.ToChainId, params.Header.ToChainType, params.Header.ToChainName, errorStr = utils.CheckChainPartialType(params.Header.ToChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}

	ownerEvmAddressBytes, err := utils.HexToBytes(params.ProxyParams.ProxyHeader.OwnerEvmAddress)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("Unable to parse owner evm address: %v", err.Error()))
	}

	ownerEvmAddress := binary.BigEndian.Uint64(ownerEvmAddressBytes)
	ownerTvmAddress, err := address.ParseAddr(params.ProxyParams.ProxyHeader.OwnerTvmAddress)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("Unable to parse owner tvm address: %v", err.Error()))
	}

	workChain, err := strconv.Atoi(params.ProxyParams.WorkChain)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("Unable to parse workchain: %v", err.Error()))
	}

	proxyAddress, proxyInit, err := CalculateWallet(ownerEvmAddress, ownerTvmAddress, entryPointAddress, workChain)
	if err != nil {
		return nil, err
	}

	messageHash, err := ExecutionDataHash(ExecutionDataParams{
		Regime:      "0",
		Destination: params.ProxyParams.ExecutionData.Destination,
		Value:       params.ProxyParams.ExecutionData.Value,
		Body:        params.ProxyParams.ExecutionData.Body,
	})
	if err != nil {
		return nil, err
	}

	value, _ := strconv.Atoi(params.ProxyParams.ExecutionData.Value)
	if withProxyInit {
		value += 100000000 + 100000000
	} else {
		value += 100000000
	}

	fmt.Print(params.Header)
	return MessageOpTvm{
		Header: params.Header,
		ProxyParams: ProxyParams{
			ProxyHeader: ProxyHeaderParams{
				Nonce:           "0",
				EntryPoint:      entryPointAddress.String(),
				PayeeAddress:    "",
				OwnerEvmAddress: params.ProxyParams.ProxyHeader.OwnerEvmAddress,
				OwnerTvmAddress: params.ProxyParams.ProxyHeader.OwnerTvmAddress,
			},
			ExecutionData: ExecutionDataParams{
				Regime:      "0",
				Destination: params.ProxyParams.ExecutionData.Destination,
				Value:       params.ProxyParams.ExecutionData.Value,
				Body:        params.ProxyParams.ExecutionData.Body,
			},
			WithProxyInit:   "true",
			ProxyWalletCode: hex.EncodeToString(proxyInit.ToBOC()),
			WorkChain:       params.ProxyParams.WorkChain,
		},
		ProxyAddress: proxyAddress.String(),
		ValueNano:    big.NewInt(int64(value)).String(), // default to 0.1 ton
		MessageHash:  hex.EncodeToString(messageHash),
	}, nil
}

//http://localhost:8080/api/tvm?query=unsigned-entrypoint-request&txtype=1&fid=11155111&fsigner=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&tid=1667471769&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&p-init=false&p-workchain=-1&p-evm=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&exe-target=EQAW3iupIDrCICc7SbcY_SBP6jCNO-F8v91dG9XNLHw-lE9k&exe-value=200000000&exe-body=

/*
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

*/
