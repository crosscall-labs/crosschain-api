package tvmHandler

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/crosscall-labs/crosschain-api/api/tvm/utils/entrypoint"
	"github.com/crosscall-labs/crosschain-api/api/tvm/utils/proxyWallet"
	"github.com/crosscall-labs/crosschain-api/pkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func CreateUnsignedMintCall(
	to *address.Address,
	query_id uint64,
	jetton_amount uint64,
	forward_ton_amount uint64,
	assetAddress *address.Address,
	total_ton_amount uint64) (proxyWallet.ExecutionData, []byte) {

	messageRegime := uint64(0)
	messageDestination := assetAddress
	messageValue := total_ton_amount + tlb.MustFromTON("0.01").Nano().Uint64()
	messageBody := JettonMintMessage(*to, query_id, jetton_amount, forward_ton_amount, *assetAddress, total_ton_amount)

	executionData := proxyWallet.ExecutionData{
		Regime:      byte(messageRegime),
		Destination: messageDestination,
		Value:       messageValue,
		Body:        messageBody,
	}

	/// start with  a5b70ca6d63a158f2d5ef3965f1fac9eb23498040e0066896cd18c0c5f7e670c
	/// end with    61aad896921377300b1b735cd47a8939e2cddabf700fc6c7e9fe391660b35474
	// qwerty, _ := hex.DecodeString("a5b70ca6d63a158f2d5ef3965f1fac9eb23498040e0066896cd18c0c5f7e670c")
	// qwerty2, _ := hashCellWithEthereumPrefix(qwerty)
	// fmt.Printf("\nmessage hash before eth: %v", hex.EncodeToString(qwerty))
	// fmt.Printf("\nmessage hash after eth:  %v", hex.EncodeToString(qwerty2))
	// fmt.Printf("\nsupposed to be:  %v", "61aad896921377300b1b735cd47a8939e2cddabf700fc6c7e9fe391660b35474")

	messageHash := proxyWallet.ExecutionDataToCell(executionData).Hash()

	return executionData, messageHash
}

func SignedEntryPointRequest(r *http.Request, parameters ...*SignedEntryPointRequestParams) (interface{}, error) {
	utils.LogNotice("SignedEntryPointRequest called!")
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
		fmt.Print(w)
		return nil, err
	}

	// ##################### PARSE AND VALIDATE PARAMS ##########################
	utils.LogNotice("Begin parse & validate parameters")
	entrypointAddress := address.MustParseAddr("kQD-99yAg5IsTqRt2qiw1IDkEexCXcaaGmMU-MPhL74cHwCM")
	b, err := api.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}
	if ok := common.IsHexAddress(params.EvmAddress); !ok {
		utils.LogError("evm-address is not a valid hex", params.EvmAddress)
		return nil, utils.ErrInternal("evm-address is not a valid hex")
	}
	evmAddress := common.HexToAddress(params.EvmAddress)
	evmAddressBigInt := new(big.Int).SetBytes(evmAddress.Bytes())
	if evmAddressBigInt.BitLen() > 160 {
		utils.LogError("invalid evm-address", "exceeds 160 bits")
		return nil, utils.ErrInternal("invalid address: exceeds 160 bits")
	}
	tvmAddress := address.MustParseAddr(params.TvmAddress)
	initNonce := 0 // should be taking entrypoint from params by static for now
	proxyWalletAddress, state := calculateProxyWalletAddress(uint64(initNonce), entrypointAddress, evmAddressBigInt, tvmAddress, byte(b.Workchain))
	if _, err := getWalletInfo(proxyWalletAddress.String()); err != nil { // don't need any data from proxy wallet, just if valid
		utils.LogInfoSimple(fmt.Sprintf("proxy wallet %+v status: INITIALIZED", proxyWalletAddress.String()))
		isInit = false
	} else {
		utils.LogInfoSimple(fmt.Sprintf("proxy wallet %+v status: NOT INITIALIZED", proxyWalletAddress.String()))
		isInit = true
	}
	executionData, err := ToExecutionData(params.Message.Data)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	messageHash := proxyWallet.ExecutionDataToCell(executionData).Hash()

	// ######################### VALIDATE SIGNATURE #############################
	utils.LogNotice("Begin signature vaidation")
	var signature proxyWallet.Signature
	signature.V, err = strconv.ParseUint(params.Message.Signature.V, 10, 64)
	if err != nil {
		utils.LogError("invalid sig-v value", err.Error())
		return nil, utils.ErrInternal(fmt.Sprintf("invalid sig-v value: %v", err.Error()))
	}

	signatureR, err := hex.DecodeString(params.Message.Signature.R)
	if err != nil {
		utils.LogError("invalid sig-s value", err.Error())
		return nil, utils.ErrInternal(fmt.Sprintf("invalid sig-r value: %v", err.Error()))
	}
	signature.R, _ = utils.BytesToUint64(signatureR)

	signatureS, err := hex.DecodeString(params.Message.Signature.S)
	if err != nil {
		utils.LogError("invalid sig-s value", err.Error())
		return nil, utils.ErrInternal(fmt.Sprintf("invalid sig-s value: %v", err.Error()))
	}
	signature.S, _ = utils.BytesToUint64(signatureR)

	if signature.V >= 27 {
		signature.V -= 27
	}

	signatureBytes := append(signatureR, signatureS...)
	signatureBytes = append(signatureBytes, byte(signature.V))

	utils.LogInfo("Signature details", utils.FormatKeyValueLogs([][2]string{
		{"address", evmAddress.Hex()},
		{"hash", hex.EncodeToString(messageHash)},
		{"signature", hex.EncodeToString(signatureBytes)},
		{"module", "signature-validation"},
	}))

	if ok, err := ValidateEvmEcdsaSignature(messageHash, signatureBytes, evmAddress); !ok || err != nil {
		if err != nil {
			utils.LogError("error validating signature", err.Error())
			return nil, utils.ErrInternal(fmt.Sprintf("error validating signature: %v", err.Error()))
		} else {
			utils.LogError("signature validation failed", "invaid signature")
			return nil, utils.ErrInternal("Signature validation failed: invalid signature")
		}
	}

	// ############################ CALL BUILDER ################################
	utils.LogNotice("Begin call builder")
	proxyWalletMessage := proxyWallet.ProxyWalletMessage{
		QueryId:   1,
		Signature: signature,
		Data:      executionData,
	}

	proxyMessage := &tlb.InternalMessage{}
	proxyMessage = &tlb.InternalMessage{
		IHRDisabled: true,
		Bounce:      false,
		DstAddr:     proxyWalletAddress,
		Amount:      tlb.FromNanoTONU(executionData.Value),
		Body:        proxyWallet.ProxyWalletMessageToCell(proxyWalletMessage),
	}
	fmt.Print(proxyMessage)
	if isInit {
		proxyMessage = &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     proxyWalletAddress,
			Amount:      tlb.FromNanoTONU(executionData.Value),
			Body:        proxyWallet.ProxyWalletMessageToCell(proxyWalletMessage),
		}
	} else {
		proxyMessage = &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     proxyWalletAddress,
			Amount:      tlb.FromNanoTONU(executionData.Value),
			Body:        proxyWallet.ProxyWalletMessageToCell(proxyWalletMessage),
			StateInit:   state,
		}
	}

	var txproxy *tlb.Transaction
	var blockproxy *ton.BlockIDExt
	var tx *tlb.Transaction
	var block *ton.BlockIDExt
	if !isInit {
		fmt.Print(txproxy)
		fmt.Print(blockproxy)
		fmt.Print(tx)
		fmt.Print(block)
		txproxy, blockproxy, err = w.SendWaitTransaction(ctx, &wallet.Message{
			Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
			InternalMessage: &tlb.InternalMessage{
				IHRDisabled: true,
				Bounce:      false,
				DstAddr:     proxyWalletAddress,
				Amount:      tlb.MustFromTON("0.15"),
				Body:        proxyWallet.ProxyWalletMessageToCell(proxyWalletMessage),
				StateInit:   state,
			},
		})
		if err != nil {
			fmt.Printf("\nI got this error: %v", err)
		}
	}

	entrypointMessage := entrypoint.EntrypointMessage{
		Destination: proxyWalletAddress,
		Body:        proxyWallet.ProxyWalletMessageToCell(proxyWalletMessage),
	}
	fmt.Print(entrypointMessage)

	queryId := uint64(0)
	tx, block, _ = w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     entrypointAddress,
			Amount:      tlb.MustFromTON("0.2"),
			Body:        entrypoint.EntrypointMessageToCell(entrypointMessage, queryId),
		},
	})

	return struct {
		Tx []TxBlockResponse
	}{
		Tx: []TxBlockResponse{
			ParseTxBlockResponse(txproxy, blockproxy),
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

// absolutely works!!!
//newest jetton (jetton minter) address: EQBMNUHziiUIUT_rHOpcAGF_IMT8flnnQQxfwu0jQLXVJJIW

//http://localhost:8080/api/tvm?query=swap-to-data-info&chain-id=1667471769&user-address=kQAqU-Wt4oIYMD-NRg803-rAoUaEHfdaUVZXY1fLVe0CoTlG&asset-address=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&asset-amount=100000000000

// newest
//http://localhost:8080/api/tvm?query=swap-to-data-info&chain-id=1667471769&user-address=kQAqU-Wt4oIYMD-NRg803-rAoUaEHfdaUVZXY1fLVe0CoTlG&asset-address=EQBMNUHziiUIUT_rHOpcAGF_IMT8flnnQQxfwu0jQLXVJJIW&asset-amount=100000000000

//http://localhost:8080/api/tvm?query=unsigned-entrypoint-request&txtype=1&fid=11155111&fsigner=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&tid=1667471769&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&p-init=false&p-workchain=-1&p-evm=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&exe-target=EQAW3iupIDrCICc7SbcY_SBP6jCNO-F8v91dG9XNLHw-lE9k&exe-value=200000000&p-nonce=0&exe-regime=0&exe-target=kQCF-z9-RXDesmTrzHRzGGr8XegUHG8MD7IeTFMbgUGC43vy&exe-value=20000000&exe-body=b5ee9c7241010301009700016d00000015000000000000004880054a7cb5bc50430607f1a8c1e69bfd581428d083beeb4a2acaec6af96abda05427312d00a2e90edd00100101af178d451900000000000000485174876e800800f5c76f48954433b8e0b6b0830f0d361d72d5ca440d546794baeb211a6865f7f700217ecfdf915c37ac993af31d1cc61abf177a05071bc303ec879314c6e05060b8cd312d03020000ba199643

// newest
//http://localhost:8080/api/tvm?query=unsigned-entrypoint-request&txtype=1&fid=11155111&fsigner=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&tid=1667471769&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&p-init=false&p-workchain=-1&p-evm=ED81eFe292283f031CF49d5aBC24A988dABC0e2B&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&&p-nonce=0&exe-regime=0&exe-target=EQBMNUHziiUIUT_rHOpcAGF_IMT8flnnQQxfwu0jQLXVJJIW&exe-value=200000000&exe-body=b5ee9c7241010301009700016d00000015000000000000004880054a7cb5bc50430607f1a8c1e69bfd581428d083beeb4a2acaec6af96abda05427312d00a2e90edd00100101af178d451900000000000000485174876e800800f5c76f48954433b8e0b6b0830f0d361d72d5ca440d546794baeb211a6865f7f700130d507ce28942144ffac73a9700185fc8313f1f9679d04317f0bb48d02d75490d312d0302000014bf99a5

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

	// fmt.Printf("\n body boc:  %v", hex.EncodeToString(executionData.Body.ToBOC()))

	// no we want to try to call ToExecutionData
	// t, _ := cell.FromBOC(executionData.Body.ToBOC())
	// fmt.Printf("\n body boc2: %v", hex.EncodeToString(t.ToBOC()))
	// //to/from boc same

	// var testRegime []byte
	// testRegime = append(testRegime, executionData.Regime)
	// executionData2 := ExecutionDataParams{
	// 	Regime:      hex.EncodeToString(testRegime),
	// 	Destination: executionData.Destination.String(),
	// 	Value:       strconv.FormatUint(executionData.Value, 10),
	// 	Body:        hex.EncodeToString(executionData.Body.ToBOC()),
	// }

	// executionData3, err := ToExecutionData(executionData2) // we don't want to use the body of the proxyparams
	// hash1 := proxyWallet.ExecutionDataToCell(executionData3).Hash()
	// hash2, _ := ExecutionDataHash(executionData2)
	// fmt.Printf("\n hash via ExecutionDataToCell: %v", hex.EncodeToString(hash1))
	// fmt.Printf("\n hash via ExecutionDataHash:   %v", hex.EncodeToString(hash2))

	return UnsignedMintToRequestResponse{
		Regime:      fmt.Sprint(executionData.Regime),
		Destination: executionData.Destination.String(),
		Value:       fmt.Sprint(executionData.Value),
		Body:        hex.EncodeToString(executionData.Body.ToBOC()),
		Hash:        hex.EncodeToString(messageHash),
	}, nil
}

func UnsignedBurnFromRequest(r *http.Request, parameters ...*UnsignedMintToRequestParams) (interface{}, error) {
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
	proxyWalletAddress, _ := calculateProxyWalletAddress(nonce, entrypointAddress, evmAddressBigInt, tvmAddress, byte(b.Workchain))
	// call get_wallet_info, if fails, wallet is not init
	if _, err := getWalletInfo(proxyWalletAddress.String()); err != nil {
		isInit = false
	}
	executionData, err := ToExecutionData(params.ProxyParams.ExecutionData) // we don't want to use the body of the proxyparams
	if err != nil {
		fmt.Printf("\nerr is inside of here: %v", err)
		return nil, utils.ErrInternal(err.Error())
	}
	messageHash := proxyWallet.ExecutionDataToCell(executionData).Hash()
	// messageHashEth, err := hashCellWithEthereumPrefix(messageHash)
	// if err != nil {
	// 	return nil, utils.ErrInternal(err.Error())
	// } // this is used to create the exact format but already auto performed by EVM wallets
	value := executionData.Value + tlb.MustFromTON("0.02").Nano().Uint64()
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
		MessageHash:  hex.EncodeToString(messageHash),
	}, nil
}
