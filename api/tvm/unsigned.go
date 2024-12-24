package tvmHandler

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type ExecutionData struct {
	Regime      byte             `json:"regime"`
	Destination *address.Address `json:"target"`
	Value       uint64           `json:"value"`
	Body        *cell.Cell       `json:"body"`
}

type Signature struct {
	V uint64
	R uint64
	S uint64
}

type ProxyWalletMessage struct {
	QueryId   uint64
	Signature Signature
	Data      ExecutionData
}

type EntrypointMessage struct {
	Destination address.Address
	Body        cell.Cell
}

func ExecutionDataToCell(message ExecutionData) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(uint8(message.Regime)), 8).
		MustStoreAddr(message.Destination).
		MustStoreUInt(message.Value, 64).
		MustStoreRef(message.Body).
		EndCell()
}

func SignatureToCell(signature Signature) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(signature.V, 8).
		MustStoreUInt(signature.R, 256).
		MustStoreUInt(signature.S, 256).
		EndCell()
}

func ProxyWalletMessageToCell(message ProxyWalletMessage) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(11, 32).
		MustStoreUInt(message.QueryId, 64).
		MustStoreRef(SignatureToCell(message.Signature)).
		MustStoreRef(ExecutionDataToCell(message.Data)).
		EndCell()
}

func EntrypointMessageToCell(message EntrypointMessage) *cell.Cell {
	return cell.BeginCell().
		MustStoreAddr(&message.Destination).
		MustStoreRef(&message.Body).
		EndCell()
}

func CreateUnsignedMintCall(
	to address.Address,
	query_id uint64,
	jetton_amount uint64,
	forward_ton_amount uint64,
	assetAddress address.Address,
	total_ton_amount uint64) (ExecutionData, []byte, error) {

	messageRegime := uint64(0)
	messageDestination := assetAddress
	messageValue := total_ton_amount + tlb.MustFromTON("0.01").Nano().Uint64()
	messageBody := MintMessage(to, query_id, jetton_amount, forward_ton_amount, assetAddress, total_ton_amount)

	executionData := ExecutionData{
		Regime:      byte(messageRegime),
		Destination: &messageDestination,
		Value:       messageValue,
		Body:        messageBody,
	}

	messageHash := executionData.Body.Hash()
	messageHashEth, err := hashCellWithEthereumPrefix(messageHash)

	return executionData, messageHashEth, err
}

type SignedEntryPointRequestParams struct {
	EvmAddress   string `query:"evm-address"`
	TvmAddress   string `query:"tvm-address"`
	AssetAddress string `query:"asset-address"`
	AssetAmount  string `query:"asset-amount"`
	Message      struct {
		QueryId   string `query:"msg-query-id"`
		Signature struct {
			V string `query:"sig-v"`
			R string `query:"sig-r"`
			S string `query:"sig-s"`
		} `query:"msg-signature"`
		Data struct {
			MessageRegime      string `query:"data-regime"`
			MessageDestination string `query:"data-destination"`
			MessageValue       string `query:"data-value"`
			MessageBody        string `query:"data-body"`
		} `query:"msg-data"`
	} `query:"message"`
}

func SignedEntryPointRequest(r *http.Request, parameters ...*SignedEntryPointRequestParams) (interface{}, error) {
	var params *SignedEntryPointRequestParams
	var isInit bool

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &SignedEntryPointRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, err
		}
	}

	ctx, api, w, err := InitClient()
	if err != nil {
		return nil, err
	}
	nonce := 0
	entrypointAddress := address.MustParseAddr("EQAGJK50PW_a1ZbQWK0yldegu56FlX0nXKQIa7xzoWCzQp78")
	b, err := api.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}

	if ok := common.IsHexAddress(params.EvmAddress); !ok {
		return nil, utils.ErrInternal("evm-address is not a valid hex")
	}
	evmAddress := common.HexToAddress(params.EvmAddress)
	evmAddressBigInt := new(big.Int).SetBytes(evmAddress.Bytes())
	if evmAddressBigInt.BitLen() > 160 {
		return nil, utils.ErrInternal(fmt.Errorf("invalid address: exceeds 160 bits").Error())
	}

	tvmAddress := address.MustParseAddr(params.TvmAddress)

	proxyWalletAddress, state := calculateProxyWalletAddress(uint64(nonce), *entrypointAddress, evmAddressBigInt, *tvmAddress, byte(b.Workchain))
	// call get_wallet_info, if fails, wallet is not init
	if _, err := getWalletInfo(proxyWalletAddress.String()); err != nil {
		isInit = false
	}

	assetAddress := address.MustParseAddr(params.AssetAddress)
	assetAmount, err := strconv.ParseUint(params.AssetAmount, 10, 64) // amount to be receieved
	if err != nil {
		return nil, utils.ErrInternal(fmt.Errorf("failed to parse jetton amount: %v", err).Error())
	}
	queryId := uint64(72)
	forwardTonAmount := uint64(5000000)
	totalTonAmount := uint64(10000000)
	executionData, messageHash, err := CreateUnsignedMintCall(*tvmAddress, queryId, assetAmount, forwardTonAmount, *assetAddress, totalTonAmount)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	// we are calling the mint function to mint tokens to user tvm address
	// we check if user sa is initialize

	var signature Signature
	signature.V, err = strconv.ParseUint(params.Message.Signature.V, 10, 64)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("invalid sig-v value: %v", err.Error()))
	}

	signatureR, err := hex.DecodeString(params.Message.Signature.R)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("invalid sig-r value: %v", err.Error()))
	}
	signature.R, _ = utils.BytesToUint64(signatureR)

	signatureS, err := hex.DecodeString(params.Message.Signature.S)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("invalid sig-s value: %v", err.Error()))
	}
	signature.S, _ = utils.BytesToUint64(signatureR)

	signatureBytes := append(signatureR, signatureS...)
	signatureBytes = append(signatureBytes, byte(signature.V))
	if ok, err := ValidateEvmEcdsaSignature(messageHash, signatureBytes, evmAddress); !ok || err != nil {
		if err != nil {
			return nil, utils.ErrInternal(fmt.Sprintf("Error validating signature: %v\n", err))
		} else {
			return nil, utils.ErrInternal("Signature validation failed: invalid signature")
		}
	}

	proxyWalletMessage := ProxyWalletMessage{
		QueryId:   0,
		Signature: signature,
		Data:      executionData,
	}

	proxyMessage := &tlb.InternalMessage{}
	if isInit {
		proxyMessage = &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     proxyWalletAddress,
			Amount:      tlb.FromNanoTONU(executionData.Value),
			Body:        ProxyWalletMessageToCell(proxyWalletMessage),
		}
	} else {
		proxyMessage = &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     proxyWalletAddress,
			Amount:      tlb.FromNanoTONU(executionData.Value),
			Body:        ProxyWalletMessageToCell(proxyWalletMessage),
			StateInit:   state,
		}
	}

	entrypointMessage := EntrypointMessage{
		Destination: *proxyWalletAddress,
		Body:        *proxyMessage.Payload(),
	}

	tx, block, err := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     entrypointAddress,
			Amount:      tlb.MustFromTON("0.05"),
			Body:        EntrypointMessageToCell(entrypointMessage),
		},
	})

	return ParseTxBlockResponse(tx, block, err), nil
}

type UnsignedEntryPointRequestParams struct {
	Header      utils.MessageHeader `query:"header"`
	ProxyParams ProxyParams         `query:"proxy"`
}

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
	OwnerTvmAddress string `query:"p-tvm"`                        // our social login SHOULD generate this
}

type ExecutionDataParams struct {
	Regime      string `query:"exe-regime" optional:"true"`
	Destination string `query:"exe-target" optional:"true"`
	Value       string `query:"exe-value" optional:"true"`
	Body        string `query:"exe-body" optional:"true"`
}

// UnsignedEntryPointRequestResponse:
type MessageOpTvm struct {
	Header       utils.MessageHeader `json:"header"`
	ProxyParams  ProxyParams         `json:"proxy"`
	ProxyAddress string              `json:"proxy-address"`
	ValueNano    string              `json:"value"`
	MessageHash  string              `json:"hash"`
}

type UnsignedMintToRequestParams struct {
	ChainId      string `query:"chain-id"`
	VM           string `query:"vm" optional:"true"`
	UserAddress  string `query:"user-address"`
	AssetAddress string `query:"asset-address"`
	AssetAmount  string `query:"asset-amount"`
}

func UnsignedMintToRequest(r *http.Request, parameters ...*UnsignedMintToRequestParams) (interface{}, error) {
	var params *UnsignedMintToRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &UnsignedMintToRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, err
		}
	}

	assetAddress := address.MustParseAddr(params.ProxyParams)
	assetAmount, err := strconv.ParseUint(params.AssetAmount, 10, 64) // amount to be receieved
	if err != nil {
		return nil, utils.ErrInternal(fmt.Errorf("failed to parse jetton amount: %v", err).Error())
	}
	queryId := uint64(72)
	forwardTonAmount := uint64(5000000)
	totalTonAmount := uint64(10000000)
	executionData, messageHash, err := CreateUnsignedMintCall(*tvmAddress, queryId, assetAmount, forwardTonAmount, *assetAddress, totalTonAmount)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	// type ExecutionData struct {
	// 	Regime      byte             `json:"regime"`
	// 	Destination *address.Address `json:"target"`
	// 	Value       uint64           `json:"value"`
	// 	Body        *cell.Cell       `json:"body"`
	// }
	hex.EncodeToString(executionData.Regime)
	executionData.Destination.String()
	strconv.ParseUint(executionData.Value, 10, 64)
	executionData.Body.ToBOC()
	hex.EncodeToString(messageHash)

	return nil, nil
}

func UnsignedMintFromRequest(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	return nil, nil
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
	/////////////////////////////////////////////////////////

	var isInit bool

	ctx, api, _, err := InitClient()
	if err != nil {
		return nil, err
	}
	entrypointAddress := address.MustParseAddr("EQAGJK50PW_a1ZbQWK0yldegu56FlX0nXKQIa7xzoWCzQp78")
	b, err := api.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nvalue of params.ProxyParams.WithProxyInit: %v\n", params.ProxyParams.WithProxyInit)

	withProxyInit, err := strconv.ParseBool(params.ProxyParams.WithProxyInit)
	if err != nil {
		return nil, err
	}

	nonce, err := strconv.ParseUint(params.ProxyParams.ProxyHeader.Nonce, 10, 64)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("nonce invalid: %v", err.Error()))
	}

	if ok := common.IsHexAddress(params.ProxyParams.ProxyHeader.OwnerEvmAddress); !ok {
		return nil, utils.ErrInternal("evm-address is not a valid hex")
	}
	evmAddress := common.HexToAddress(params.ProxyParams.ProxyHeader.OwnerEvmAddress)
	evmAddressBigInt := new(big.Int).SetBytes(evmAddress.Bytes())
	if evmAddressBigInt.BitLen() > 160 {
		return nil, utils.ErrInternal(fmt.Errorf("invalid address: exceeds 160 bits").Error())
	}

	tvmAddress := address.MustParseAddr(params.ProxyParams.ProxyHeader.OwnerTvmAddress)

	proxyWalletAddress, state := calculateProxyWalletAddress(nonce, *entrypointAddress, evmAddressBigInt, *tvmAddress, byte(b.Workchain))
	// call get_wallet_info, if fails, wallet is not init
	if _, err := getWalletInfo(proxyWalletAddress.String()); err != nil {
		isInit = false
	}

	executionData, err := ToExecutionData(params.ProxyParams.ExecutionData)
	messageHash := executionData.Body.Hash()
	messageHashEth, err := hashCellWithEthereumPrefix(messageHash)

	// assetAddress := address.MustParseAddr(params.ProxyParams.)
	// assetAmount, err := strconv.ParseUint(params.AssetAmount, 10, 64) // amount to be receieved
	// if err != nil {
	// 	return nil, utils.ErrInternal(fmt.Errorf("failed to parse jetton amount: %v", err).Error())
	// }
	// queryId := uint64(72)
	// forwardTonAmount := uint64(5000000)
	// totalTonAmount := uint64(10000000)
	// executionData, messageHash, err := CreateUnsignedMintCall(*tvmAddress, queryId, assetAmount, forwardTonAmount, *assetAddress, totalTonAmount)
	// if err != nil {
	// 	return nil, utils.ErrInternal(err.Error())
	// }

	value, _ := strconv.Atoi(params.ProxyParams.ExecutionData.Value)
	if withProxyInit {
		// fmt.Print("\ninside of withproxyinit")
		// url := os.Getenv("TONX_API_BASE_TESTNET_URL")
		// apiKey := os.Getenv("TONX_TESTNET_API_KEY_1")
		// jsonrpc := os.Getenv("TONX_API_JSONRPC")

		// proxyWalletData := PackProxyWalletData(0, entryPointAddress, ownerEvmAddress, ownerTvmAddress)

		// cell.BeginCell().EndCell().ToBOC()

		// request := tonx.TonEstimateFee{
		// 	Address: proxyAddress.String(),
		// 	Body:    hex.EncodeToString(hex.EncodeToString(proxyInit.ToBOC())),
		// 	InitCode: proxyWalletCodeHex,
		// 	InitData: proxyWalletData,
		// }

		// response, _ := tonx.SendTonXRequest(url, apiKey, jsonrpc, 1, "estimateFee", request)
		// var parsedResponse tonx.TonEstimateFeeResponse
		// if err := json.Unmarshal([]byte(response), &parsedResponse); err != nil {
		// 	fmt.Printf("Error parsing response: %v\n", err)
		// 	return nil, err
		// }

		// fmt.Printf("\nfee estimation: \n%v\n", request)

		value += 100000000 + 100000000
	} else {
		value += 100000000
	}

	// thingy := cell.BeginCell().EndCell()
	// fmt.Printf("cell thing\n%v\n", hex.EncodeToString(thingy.ToBOC())) //-> b5ee9c724101010100020000004cacb9cd

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
				Value: func() string {
					if params.ProxyParams.ExecutionData.Value == "" {
						return "0"
					}
					return params.ProxyParams.ExecutionData.Value
				}(),
				Body: func() string {
					if params.ProxyParams.ExecutionData.Body == "" || params.ProxyParams.ExecutionData.Body == "00" {
						return "b5ee9c724101010100020000004cacb9cd"
					}
					return params.ProxyParams.ExecutionData.Body
				}(),
			},
			WithProxyInit:   params.ProxyParams.WithProxyInit,
			ProxyWalletCode: hex.EncodeToString(proxyInit.ToBOC()),
			WorkChain:       params.ProxyParams.WorkChain,
		},
		ProxyAddress: proxyWalletAddress.String(),
		ValueNano:    big.NewInt(int64(value)).String(), // default to 0.1 ton
		MessageHash:  hex.EncodeToString(messageHashEth),
	}, nil
}
