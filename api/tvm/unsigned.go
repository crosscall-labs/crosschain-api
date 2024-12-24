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
	Destination *address.Address
	Body        *cell.Cell
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

func EntrypointMessageToCell(message EntrypointMessage, queryId uint64) *cell.Cell {
	msgBody := cell.BeginCell().
		MustStoreAddr(message.Destination).
		MustStoreRef(message.Body).
		EndCell()

	return cell.BeginCell().
		MustStoreUInt(1, 32).
		MustStoreUInt(queryId, 64).
		MustStoreRef(msgBody).
		EndCell()
}

func CreateUnsignedMintCall(
	to *address.Address,
	query_id uint64,
	jetton_amount uint64,
	forward_ton_amount uint64,
	assetAddress *address.Address,
	total_ton_amount uint64) (ExecutionData, []byte) {

	messageRegime := uint64(0)
	messageDestination := assetAddress
	messageValue := total_ton_amount + tlb.MustFromTON("0.01").Nano().Uint64()
	messageBody := MintMessage(*to, query_id, jetton_amount, forward_ton_amount, *assetAddress, total_ton_amount)

	executionData := ExecutionData{
		Regime:      byte(messageRegime),
		Destination: messageDestination,
		Value:       messageValue,
		Body:        messageBody,
	}

	/// start with  a5b70ca6d63a158f2d5ef3965f1fac9eb23498040e0066896cd18c0c5f7e670c
	/// end with    61aad896921377300b1b735cd47a8939e2cddabf700fc6c7e9fe391660b35474
	qwerty, _ := hex.DecodeString("a5b70ca6d63a158f2d5ef3965f1fac9eb23498040e0066896cd18c0c5f7e670c")
	qwerty2, _ := hashCellWithEthereumPrefix(qwerty)
	fmt.Printf("\nmessage hash before eth: %v", hex.EncodeToString(qwerty))
	fmt.Printf("\nmessage hash after eth:  %v", hex.EncodeToString(qwerty2))
	fmt.Printf("\nsupposed to be:  %v", "61aad896921377300b1b735cd47a8939e2cddabf700fc6c7e9fe391660b35474")

	messageHash := executionData.Body.Hash()

	return executionData, messageHash
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
	fmt.Print((isInit))

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
	entrypointAddress := address.MustParseAddr("EQCjyLHEhFEAVFSaoCEcSLx95yuEA4Eh1z1Ge7woXBYwx3i4")
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
	proxyWalletAddress, _ := calculateProxyWalletAddress(uint64(nonce), entrypointAddress, evmAddressBigInt, tvmAddress, byte(b.Workchain))
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
	executionData, messageHash := CreateUnsignedMintCall(tvmAddress, queryId, assetAmount, forwardTonAmount, assetAddress, totalTonAmount)
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
	proxyMessage = &tlb.InternalMessage{
		IHRDisabled: true,
		Bounce:      false,
		DstAddr:     proxyWalletAddress,
		Amount:      tlb.FromNanoTONU(executionData.Value),
		Body:        ProxyWalletMessageToCell(proxyWalletMessage),
	}
	fmt.Print(proxyMessage)
	// if isInit {
	// 	proxyMessage = &tlb.InternalMessage{
	// 		IHRDisabled: true,
	// 		Bounce:      false,
	// 		DstAddr:     proxyWalletAddress,
	// 		Amount:      tlb.FromNanoTONU(executionData.Value),
	// 		Body:        ProxyWalletMessageToCell(proxyWalletMessage),
	// 	}
	// } else {
	// 	proxyMessage = &tlb.InternalMessage{
	// 		IHRDisabled: true,
	// 		Bounce:      false,
	// 		DstAddr:     proxyWalletAddress,
	// 		Amount:      tlb.FromNanoTONU(executionData.Value),
	// 		Body:        ProxyWalletMessageToCell(proxyWalletMessage),
	// 		StateInit:   state,
	// 	}
	// }

	// txproxy, blockproxy, _ := w.SendWaitTransaction(ctx, &wallet.Message{
	// 	Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
	// 	InternalMessage: &tlb.InternalMessage{
	// 		IHRDisabled: true,
	// 		Bounce:      false,
	// 		DstAddr:     proxyWalletAddress,
	// 		Amount:      tlb.MustFromTON("0.1"),
	// 		Body:        cell.BeginCell().EndCell(),
	// 		StateInit:   state,
	// 	},
	// })

	entrypointMessage := EntrypointMessage{
		Destination: proxyWalletAddress,
		Body:        ProxyWalletMessageToCell(proxyWalletMessage),
	}

	tx, block, _ := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     entrypointAddress,
			Amount:      tlb.MustFromTON("0.2"),
			Body:        EntrypointMessageToCell(entrypointMessage, queryId),
		},
	})

	return struct {
		Tx []TxBlockResponse
	}{
		Tx: []TxBlockResponse{
			// ParseTxBlockResponse(txproxy, blockproxy),
			ParseTxBlockResponse(tx, block),
		},
	}, nil
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

type UnsignedMintToRequestResponse struct {
	Regime      string `json:"regime"`
	Destination string `json:"destination"`
	Value       string `json:"value"`
	Body        string `json:"body"`
	Hash        string `json:"hash"`
}

//http://localhost:8080/api/tvm?query=swap-to-data-info&chain-id=1667471769&user-address=kQAqU-Wt4oIYMD-NRg803-rAoUaEHfdaUVZXY1fLVe0CoTlG&asset-address=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&asset-amount=100000000000

//http://localhost:8080/api/tvm?query=unsigned-entrypoint-request&txtype=1&fid=11155111&fsigner=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&tid=1667471769&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&p-init=false&p-workchain=-1&p-evm=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&exe-target=EQAW3iupIDrCICc7SbcY_SBP6jCNO-F8v91dG9XNLHw-lE9k&exe-value=200000000&p-nonce=0&exe-regime=0&exe-target=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&exe-value=20000000&exe-body=b5ee9c7241010301009700016d00000015000000000000004880054a7cb5bc50430607f1a8c1e69bfd581428d083beeb4a2acaec6af96abda05427312d00a2e90edd00100101af178d451900000000000000485174876e800800f5c76f48954433b8e0b6b0830f0d361d72d5ca440d546794baeb211a6865f7f700217ecfdf915c37ac993af31d1cc61abf177a05071bc303ec879314c6e05060b8cd312d03020000ba199643

//http://localhost:8080/api/tvm?query=signed-entrypoint-request
//&evm-address=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&tvm-address=kQAqU-Wt4oIYMD-NRg803-rAoUaEHfdaUVZXY1fLVe0CoTlG&asset-address=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&asset-amount=100000000000
//&msg-query-id=0&sig-v=28&sig-r=c331330c63fd18aa3bdf5ddac834468fd15213839003882b588b21f7a6adbaa7&sig-s=714ec4954ddd0a08802a0545fd277055dff384c145d73f2126f799cb36ad0900
//&data-regime=0&data-destination=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&data-value=20000000&data-body=b5ee9c7241010301009700016d00000015000000000000004880054a7cb5bc50430607f1a8c1e69bfd581428d083beeb4a2acaec6af96abda05427312d00a2e90edd00100101af178d451900000000000000485174876e800800f5c76f48954433b8e0b6b0830f0d361d72d5ca440d546794baeb211a6865f7f700217ecfdf915c37ac993af31d1cc61abf177a05071bc303ec879314c6e05060b8cd312d03020000ba199643

//#r
// :
// "0x506deca88a5b8564a35574398ffef18201a1ecca729f0cd6852803807bac1e20"
// #s
// :
// "0x36ea3022b9527e666d4b4f5da4b061ee537b9a6950eec247351547f4107a833b"
// #v
// :
// 27

// as one tx
//http://localhost:8080/api/tvm?query=signed-entrypoint-request&evm-address=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&tvm-address=kQAqU-Wt4oIYMD-NRg803-rAoUaEHfdaUVZXY1fLVe0CoTlG&asset-address=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&asset-amount=100000000000&msg-query-id=0&sig-v=00&sig-r=730617d12b5b9a54267e0cf6f4cf385c2f4a247726bef3c29504b0b8e86c21a5&sig-s=37ec787db8dc684c0b3ad677c9945a2fb558a36ff594a25de49de63c4e991c36&data-regime=0&data-destination=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&data-value=20000000&data-body=b5ee9c7241010301009700016d00000015000000000000004880054a7cb5bc50430607f1a8c1e69bfd581428d083beeb4a2acaec6af96abda05427312d00a2e90edd00100101af178d451900000000000000485174876e800800f5c76f48954433b8e0b6b0830f0d361d72d5ca440d546794baeb211a6865f7f700217ecfdf915c37ac993af31d1cc61abf177a05071bc303ec879314c6e05060b8cd312d03020000ba199643

//	type SignedEntryPointRequestParams struct {
//		EvmAddress   string `query:"evm-address"`
//		TvmAddress   string `query:"tvm-address"`
//		AssetAddress string `query:"asset-address"`
//		AssetAmount  string `query:"asset-amount"`
//		Message      struct {
//			QueryId   string `query:"msg-query-id"`
//			Signature struct {
//				V string `query:"sig-v"`
//				R string `query:"sig-r"`
//				S string `query:"sig-s"`
//			} `query:"msg-signature"`
//			Data struct {
//				MessageRegime      string `query:"data-regime"`
//				MessageDestination string `query:"data-destination"`
//				MessageValue       string `query:"data-value"`
//				MessageBody        string `query:"data-body"`
//			} `query:"msg-data"`
//		} `query:"message"`
//	}
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

	userAddress := address.MustParseAddr(params.UserAddress)
	assetAddress := address.MustParseAddr(params.AssetAddress)
	assetAmount, err := strconv.ParseUint(params.AssetAmount, 10, 64) // amount to be receieved
	if err != nil {
		return nil, utils.ErrInternal(fmt.Errorf("failed to parse jetton amount: %v", err).Error())
	}
	queryId := uint64(72)
	forwardTonAmount := uint64(5000000)
	totalTonAmount := uint64(10000000)
	executionData, messageHash := CreateUnsignedMintCall(userAddress, queryId, assetAmount, forwardTonAmount, assetAddress, totalTonAmount)

	return UnsignedMintToRequestResponse{
		Regime:      fmt.Sprint(executionData.Regime),
		Destination: executionData.Destination.String(),
		Value:       fmt.Sprint(executionData.Value),
		Body:        hex.EncodeToString(executionData.Body.ToBOC()),
		Hash:        hex.EncodeToString(messageHash),
	}, nil
}

func UnsignedMintFromRequest(r *http.Request, parameters ...*interface{}) (interface{}, error) {
	return nil, nil
}

//#r
// :
// "0xc331330c63fd18aa3bdf5ddac834468fd15213839003882b588b21f7a6adbaa7"
// #s
// :
// "0x714ec4954ddd0a08802a0545fd277055dff384c145d73f2126f799cb36ad0900"
// #v
// :
// 28

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
		return nil, utils.ErrMalformedRequest(fmt.Sprintf("could not parse escrow: %v", errorStr))
	}
	params.Header.ToChainId, params.Header.ToChainType, params.Header.ToChainName, errorStr = utils.CheckChainPartialType(params.Header.ToChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(fmt.Sprintf("could not parse escrow2: %v", errorStr))
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
	proxyWalletAddress, _ := calculateProxyWalletAddress(nonce, *entrypointAddress, evmAddressBigInt, *tvmAddress, byte(b.Workchain))
	// call get_wallet_info, if fails, wallet is not init
	if _, err := getWalletInfo(proxyWalletAddress.String()); err != nil {
		isInit = false
	}
	fmt.Print("\ngot over here")
	executionData, err := ToExecutionData(params.ProxyParams.ExecutionData)
	if err != nil {
		fmt.Printf("\nerr is inside of here: %v", err)
		return nil, utils.ErrInternal(err.Error())
	}
	fmt.Print("\ngot over here")
	messageHash := executionData.Body.Hash()
	fmt.Print("\ngot over here")
	messageHashEth, err := hashCellWithEthereumPrefix(messageHash)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	fmt.Print("\ngot over here")
	value := executionData.Value + tlb.MustFromTON("0.02").Nano().Uint64()
	fmt.Print("\ngot over here")
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
			ExecutionData:   params.ProxyParams.ExecutionData,
			WithProxyInit:   strconv.FormatBool(isInit),
			ProxyWalletCode: "",
			WorkChain:       params.ProxyParams.WorkChain,
		},
		ProxyAddress: proxyWalletAddress.String(),
		ValueNano:    strconv.FormatUint(value, 10),
		MessageHash:  hex.EncodeToString(messageHashEth),
	}, nil
}
