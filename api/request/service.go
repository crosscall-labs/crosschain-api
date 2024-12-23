package requestHandler

import (
	"fmt"
	"net/http"

	evmHandler "github.com/laminafinance/crosschain-api/api/evm"
	tvmHandler "github.com/laminafinance/crosschain-api/api/tvm"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func AssetMintRequest(r *http.Request, parameters ...*utils.AssetMintRequestParams) (interface{}, error) {
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

	var err error
	_, params.VM, err = utils.GetChainType(params.ChainId)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}

	switch params.VM {
	case "evm":
		return evmHandler.AssetMintRequest(r, params)
	case "tvm":
		return tvmHandler.AssetMintRequest(r, params)
	// case "svm":
	// 	// return svmHandler.AssetInfoRequestSvm(params)
	default:
		return nil, fmt.Errorf("Virtual machine %v is unsupported", params.VM)
	}
}

// http://localhost:8080/api/request?query=unsigned-crosschain-request&txtype=1&fid=11155111&fsigner=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&tid=1667471769&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&p-init=false&p-workchain=-1&p-evm=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&exe-target=EQAW3iupIDrCICc7SbcY_SBP6jCNO-F8v91dG9XNLHw-lE9k&exe-value=200000000&exe-body=
func UnsignedCrosschainRequest(r *http.Request) (interface{}, error) {
	params := &UnsignedCrosschainRequestParams{}

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
		//value := unsignedDataResponse.ToMessage.GetType()
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
