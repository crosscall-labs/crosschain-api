package tvmHandler

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/laminafinance/crosschain-api/pkg/tonx"
	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
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

func calcJettonWalletAddress(userAddress address.Address, assetAddress address.Address, workchain int64) *cell.Cell {
	//jettonMinterCode, _ := cell.FromBOC(jettonMinterBocBuffer)
	jettonWalletCode, _ := cell.FromBOC(jettonWalletBocBuffer)
	// jettonWalletCodeDict, _ := ByteArrayToCellDictionary(jettonWalletCode.ToRawUnsafe().Data)

	packJettonWalletData := cell.BeginCell().
		MustStoreCoins(0).
		MustStoreSlice(userAddress.Data(), userAddress.BitsLen()).
		MustStoreSlice(assetAddress.Data(), assetAddress.BitsLen()).
		MustStoreRef(jettonWalletCode).
		EndCell()

	// packJettonWalletDataDict, _ := ByteArrayToCellDictionary(packJettonWalletData.ToRawUnsafe().Data)

	jettonWalletStateInit := cell.BeginCell().
		MustStoreUInt(0, 2).
		MustStoreDict(jettonWalletCode.AsDict(256)).
		MustStoreDict(packJettonWalletData.AsDict(256)).
		MustStoreUInt(0, 1).
		EndCell()

	hashBytes := jettonWalletStateInit.Hash()
	hashInt := new(big.Int).SetBytes(hashBytes[:])

	jettonWalletAddress := cell.BeginCell().
		MustStoreUInt(4, 3).
		MustStoreInt(workchain, 8).
		MustStoreUInt(hashInt.Uint64(), 256).
		EndCell()

	return jettonWalletAddress
}

type TxBlockResponse struct {
	Tx    ParsedTransaction `json:"tx"`
	Block ParsedBlock       `json:"block"`
	Error string            `json:"error,omitempty"`
}

type ParsedTransaction struct {
	LT   string `json:"lt"`
	Type string `json:"type"`
	Hash string `json:"hash"`
	Out  string `json:"out"`
	To   string `json:"to"`
}

type ParsedBlock struct {
	Workchain string `json:"workchain"`
	Shard     string `json:"shard"`
	SeqNum    string `json:"seq_num"`
	RootHash  string `json:"root_hash"`
	FileHash  string `json:"file_hash"`
}

func ParseTxBlockResponse(tx *tlb.Transaction, block *ton.BlockIDExt, err error) TxBlockResponse {
	return TxBlockResponse{
		Tx: ParsedTransaction{
			LT:   strconv.FormatUint(tx.LT, 10),
			Type: string(tx.IO.In.MsgType),
			Hash: hex.EncodeToString(tx.Hash),
			Out:  tx.TotalFees.Coins.String(),
			To:   hex.EncodeToString(tx.AccountAddr),
		},
		Block: ParsedBlock{
			Workchain: strconv.Itoa(int(block.Workchain)),
			Shard:     strconv.FormatInt(block.Shard, 10),
			SeqNum:    strconv.FormatUint(uint64(block.SeqNo), 10),
			RootHash:  hex.EncodeToString(block.RootHash),
			FileHash:  hex.EncodeToString(block.FileHash),
		},
	}
}

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

	ctx, _, w, err := InitClient()
	if err != nil {
		return nil, err
	}

	contractAddressRaw := params.AssetAddress
	contractAddress := address.MustParseAddr(contractAddressRaw)

	userAddressRaw := params.UserAddress
	userAddress := address.MustParseAddr(userAddressRaw)
	jettonAmount, err := strconv.ParseUint(params.AssetAmount, 10, 64)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Errorf("failed to parse jetton amount: %v", err).Error())
	}

	queryId := uint64(71)
	forwardTonAmount := uint64(5000000)
	totalTonAmount := uint64(10000000)

	msgBody := MintMessage(*userAddress, queryId, jettonAmount, forwardTonAmount, *contractAddress, totalTonAmount)
	amount := tlb.MustFromTON("0.01")

	tx, block, err := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     contractAddress,
			Amount:      amount,
			Body:        msgBody,
		},
	})

	return ParseTxBlockResponse(tx, block, err), nil
}

// now that we have a way to execute, deploy + execute, view, we can formulate and execute the escrow request
// the current issue is we need a way to execute the transaction on the users behalf, ie they sign
// escrow we do not need to do this
// buuut we need a single call for the generation a store of value in the escrow
// to do this we re

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
			return nil, err
		}
	}

	var errorStr string
	params.Header.ChainId, params.Header.ChainType, params.Header.ChainName, errorStr = utils.CheckChainPartialType(params.Header.ChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		return nil, utils.ErrMalformedRequest(errorStr)
	}

	return MessageEscrowTvm{}, nil
}

func calculateProxyWalletAddress(nonce uint64, entrypointAddress address.Address, evmAddressBigInt *big.Int, tvmAddress address.Address, workchain byte) (*address.Address, *tlb.StateInit) {
	proxyWalletConfigCell := cell.BeginCell().
		MustStoreUInt(nonce, 64).
		MustStoreAddr(&entrypointAddress).
		MustStoreUInt(evmAddressBigInt.Uint64(), 160).
		MustStoreAddr(&tvmAddress).
		EndCell()

	proxyWalletCodeCell, _ := cell.FromBOC(proxyWalletBocBuffer)

	state := &tlb.StateInit{
		Data: proxyWalletConfigCell,
		Code: proxyWalletCodeCell,
	}

	stateCell, _ := tlb.ToCell(state)

	return address.NewAddress(0, workchain, stateCell.Hash()), state
}

func TestRequest(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	// doesn't use params
	// the goal is to use the counter contract code

	// 1) call the deployed counter contract
	// 2) deploy the counter contract
	// 3) both call deploy and then verify the counter contract
	// 4) listen to changes on the counter contract (tbd)

	ctx, _, w, err := InitClient()
	if err != nil {
		return nil, err
	}

	countercodehex := "b5ee9c7241010a01008b000114ff00f4a413f4bcf2c80b0102016202070202ce03060201200405006b1b088831c02456f8007434c0cc1c6c244c383c0074c7f4cfcc74c7cc3c008060841fa1d93beea63e1080683e18bc00b80c2103fcbc20001d3b513434c7c07e1874c7c07e18b46000194f842f841c8cb1fcb1fc9ed54802016e0809000db5473e003f0830000db63ffe003f08500171db07"
	countercodebytes, _ := hex.DecodeString(countercodehex)

	counterConfigCell := cell.BeginCell().
		MustStoreUInt(10, 32).
		MustStoreUInt(10, 32).
		EndCell()

	//counterCodeCell, _ := ByteArrayToCellDictionary(countercodebytes)
	counterCodeCell, _ := cell.FromBOC(countercodebytes)

	// stateInitCell := cell.BeginCell().
	// 	MustStoreUInt(0, 2).
	// 	MustStoreDict(counterCodeCell).
	// 	MustStoreDict(counterConfigCell.AsDict(256)).
	// 	MustStoreUInt(0, 1).
	// 	EndCell()

	// toaddress := calculate_contract_address(stateInitCell, 0) // so now we can calculate the address

	state := &tlb.StateInit{
		Data: counterConfigCell,
		Code: counterCodeCell,
	}

	stateCell, _ := tlb.ToCell(state)
	// if err != nil {
	// 	return nil, nil, nil, err
	// }

	addr := address.NewAddress(0, 0, stateCell.Hash())

	msgBody := cell.BeginCell().EndCell() // this will be the increment code

	amount := tlb.MustFromTON("0.02")

	// w is of type wallet
	// determine how to calculate w
	tx, block, err := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     addr,
			Amount:      amount,
			Body:        msgBody,
			StateInit:   state,
		},
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("\ntx return data: \n%v\n\ntx return block: \n%v\n", tx, block)

	//func calculate_contract_address(state_init *cell.Cell, workchain int) *cell.Cell {
	return nil, nil
}

func Test2Request(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	// just the view function
	ctx, api, _, err := InitClient()
	if err != nil {
		return nil, err
	}

	//b, err := api.CurrentMasterchainInfo(ctx)
	b := TestnetInfo

	//b, err := api.LookupBlock(ctx, -1, -0x8000000000000000, 27531166) // last init is genesis block
	// if err != nil {
	// 	return nil, err
	// }

	contractAddress := "EQDuTkPoaFG8V6KZP0SVsaDF5nzYRxLfPn9o_9WdROMmqseY"
	addr := address.MustParseAddr(contractAddress)
	props, err := api.RunGetMethod(ctx, b, addr, "get_counter")
	if err != nil {
		return nil, err
	}

	value, _ := props.Int(0)

	fmt.Printf("\nthis is the returned data from test2: \n%v\n", value)

	//func calculate_contract_address(state_init *cell.Cell, workchain int) *cell.Cell {
	return nil, nil
}

func Test3Request(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	// just the increment function
	ctx, _, w, err := InitClient()
	if err != nil {
		return nil, err
	}

	contractAddress := "EQB_aPrr_z_m11vT0Jz3EAz1G3aIS4UcZGL4dF2M0M4oJ6TV"
	addr := address.MustParseAddr(contractAddress)

	msgBody := cell.BeginCell().
		MustStoreUInt(2122802415, 32). // having trouble converting 4byte hex to number, check how to do later
		MustStoreUInt(69420, 64).
		MustStoreUInt(23, 32).
		EndCell()

	amount := tlb.MustFromTON("0.01")

	fmt.Print("I got this far")
	tx, block, err := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     addr,
			Amount:      amount,
			Body:        msgBody,
		},
	})
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}

	fmt.Printf("\nthis is the returned data from test3: \ntx:\n%v\nblock:\n%v", tx, block)

	return nil, nil
}

func Test4Request(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	// deploy, init, and call function in one call
	ctx, _, w, err := InitClient()
	if err != nil {
		return nil, err
	}

	countercodehex := "b5ee9c7241010a01008b000114ff00f4a413f4bcf2c80b0102016202070202ce03060201200405006b1b088831c02456f8007434c0cc1c6c244c383c0074c7f4cfcc74c7cc3c008060841fa1d93beea63e1080683e18bc00b80c2103fcbc20001d3b513434c7c07e1874c7c07e18b46000194f842f841c8cb1fcb1fc9ed54802016e0809000db5473e003f0830000db63ffe003f08500171db07"
	countercodebytes, _ := hex.DecodeString(countercodehex)

	counterConfigCell := cell.BeginCell().
		MustStoreUInt(10, 32).
		MustStoreUInt(10, 32).
		EndCell()

	//counterCodeCell, _ := ByteArrayToCellDictionary(countercodebytes)
	counterCodeCell, _ := cell.FromBOC(countercodebytes)

	state := &tlb.StateInit{
		Data: counterConfigCell,
		Code: counterCodeCell,
	}

	stateCell, _ := tlb.ToCell(state)

	addr := address.NewAddress(0, 0, stateCell.Hash())

	msgBody := cell.BeginCell().
		MustStoreUInt(2122802415, 32). // having trouble converting 4byte hex to number, check how to do later
		MustStoreUInt(69420, 64).
		MustStoreUInt(23, 32).
		EndCell()

	amount := tlb.MustFromTON("0.02")

	myMsg := &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     addr,
			Amount:      amount,
			Body:        msgBody,
			StateInit:   state,
		},
	}

	myMsg.InternalMessage.Payload()

	tx, block, err := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     addr,
			Amount:      amount,
			Body:        msgBody,
			StateInit:   state,
		},
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("\ntx return data: \n%v\n\ntx return block: \n%v\n", tx, block)

	return nil, nil
}

func Test5Request(r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	// just the view function
	// ctx, api, _, err := InitClient()
	// if err != nil {
	// 	return nil, err
	// }

	url := os.Getenv("TONX_API_BASE_TESTNET_URL")
	apiKey := os.Getenv("TONX_TESTNET_API_KEY_1")
	jsonrpc := os.Getenv("TONX_API_JSONRPC")

	request := tonx.TonRunGetMethod{
		Address: "EQDuTkPoaFG8V6KZP0SVsaDF5nzYRxLfPn9o_9WdROMmqseY",
		//4CDE9B6C823D71C3F9F31A19C78EE8F9B4649370B143BEF1660B2ADDE8362F4B
		//1234567890123456789012345678901234567890123456789012345678901234
		Method: "get_counter",
	}

	response, _ := tonx.SendTonXRequest(url, apiKey, jsonrpc, 1, "runGetMethod", request)
	var parsedResponse tonx.TonRunGetMethodResponse
	if err := json.Unmarshal([]byte(response), &parsedResponse); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return nil, err
	}

	// Print parsed response details
	fmt.Printf("Parsed Response: %+v\n", parsedResponse)
	fmt.Printf("Gas Used: %d\n", parsedResponse.Result.GasUsed)
	fmt.Printf("Exit Code: %d\n", parsedResponse.Result.ExitCode)
	fmt.Println("Stack Pairs:")
	for _, pair := range parsedResponse.Result.Stack {
		fmt.Printf("Type: %s, Value: %s\n", pair[0], pair[1])
	}

	fmt.Printf("\nresponse: \n%v", response)
	// responseBody, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	log.Fatalf("Error reading response body: %v", err)
	// }
	// fmt.Printf("Response Body:\n%s\n", string(responseBody))

	//func calculate_contract_address(state_init *cell.Cell, workchain int) *cell.Cell {
	return nil, nil
}

func Test6Request(r *http.Request, parameters ...interface{}) (interface{}, error) {
	// ctx, api, w, err := InitClient()
	// if err != nil {
	// 	return nil, err
	// }

	// tokenContract := address.MustParseAddr("EQBCFwW8uFUh-amdRmNY9NyeDEaeDYXd9ggJGsicpqVcHq7B")
	// master := jetton.NewJettonMasterClient(api, tokenContract)

	// data, err := master.GetJettonData(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// decimals := 9
	// content := data.Content.(*nft.ContentOnchain)
	// log.Println("total supply:", data.TotalSupply.Uint64())
	// log.Println("mintable:", data.Mintable)
	// log.Println("admin addr:", data.AdminAddr)
	// log.Println("onchain content:")
	// log.Println("	name:", content.Name)
	// log.Println("	symbol:", content.GetAttribute("symbol"))
	// if content.GetAttribute("decimals") != "" {
	// 	decimals, err = strconv.Atoi(content.GetAttribute("decimals"))
	// 	if err != nil {
	// 		log.Fatal("invalid decimals")
	// 	}
	// }
	// log.Println("	decimals:", decimals)
	// log.Println("	description:", content.Description)
	// log.Println()

	// tokenWallet, err := master.GetJettonWallet(ctx, address.MustParseAddr("EQCWdteEWa4D3xoqLNV0zk4GROoptpM1-p66hmyBpxjvbbnn"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// tokenBalance, err := tokenWallet.GetBalance(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("jetton balance:", tlb.MustFromNano(tokenBalance, decimals))

	return nil, nil
}

func getTestnetTonx() (string, string, string) {
	url := os.Getenv("TONX_API_BASE_TESTNET_URL")
	apiKey := os.Getenv("TONX_TESTNET_API_KEY_1")
	jsonrpc := os.Getenv("TONX_API_JSONRPC")
	return url, apiKey, jsonrpc
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
			return nil, err
		}
	}

	response := &utils.AssetInfoRequestResponse{}
	var result interface{}
	var err error
	response.ChainId = params.ChainId
	response.VM = params.VM

	result, err = getJettonData(params.AssetAddress)
	if err != nil {
		return nil, utils.ErrInternal(fmt.Sprintf("asset could not be found: %v", err.Error()))
	}

	response.Asset = struct {
		Address     string `json:"address"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Decimal     string `json:"decimal"`
		TotalSupply string `json:"total-supply"`
		Supply      string `json:"supply"`
		Description string `json:"description"`
	}{
		Address:     params.AssetAddress,
		Name:        "",
		Symbol:      "",
		Decimal:     "",
		TotalSupply: result.(*GetJettonDataResponse).TotalSupply,
		Supply:      "",
		Description: "",
	}

	result, err = getUserJettonWallet(params.UserAddress, params.AssetAddress)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}
	userJettonWalletRaw := result.(*GetUserJettonWalletResponse)
	result, err = getWalletData(userJettonWalletRaw.JettonWalletAddress)
	if err != nil {
		response.User = struct {
			Balance string "json:\"balance\""
		}{
			Balance: "",
		}
	}

	response.User = struct {
		Balance string "json:\"balance\""
	}{
		Balance: result.(*GetWalletDataResponse).Balance,
	}

	if params.EscrowAddress != "" {
		result, err = getUserJettonWallet(params.EscrowAddress, params.AssetAddress)
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
		userJettonWalletRaw := result.(*GetUserJettonWalletResponse).JettonWalletAddress

		result, err = getWalletData(userJettonWalletRaw)
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}

		response.Escrow = struct {
			Init         bool   `json:"init"`
			Balance      string `json:"balance"`
			LockBalance  string `json:"lock-balance"`
			LockDeadline string `json:"lock-deadline"`
		}{
			Init:         false,
			Balance:      result.(*GetWalletDataResponse).Balance,
			LockBalance:  "",
			LockDeadline: "",
		}
	}

	if params.AccountAddress != "" {
		result, err = getUserJettonWallet(params.AccountAddress, params.AssetAddress)
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}
		userJettonWalletRaw := result.(*GetUserJettonWalletResponse).JettonWalletAddress

		result, err = getWalletData(userJettonWalletRaw)
		if err != nil {
			return nil, utils.ErrInternal(err.Error())
		}

		response.Account = struct {
			Init    bool   `json:"init"`
			Balance string `json:"balance"`
		}{
			Init:    false,
			Balance: result.(*GetWalletDataResponse).Balance,
		}
	}
	return response, nil
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


need to pack the execution data in bytes

since we know raw bytes
buuuuut the toboc and fromboc function are broke in tonutils
so thus we need to rework them
*/

// func FromBOC(input []byte) (*cell.Cell, error) {
// // first lets investigate the tonutils serialization/deserialization
// 	return nil, nil
// }

/*
from cell/parse.go
func FromBOC(data []byte) (*Cell, error) {
	cells, err := FromBOCMultiRoot(data)
	if err != nil {
		return nil, err
	}

	return cells[0], nil
}

func FromBOCMultiRoot(data []byte) ([]*Cell, error) {
	if len(data) < 10 {
		return nil, errors.New("invalid boc")
	}

	r := newReader(data)

	if !bytes.Equal(r.MustReadBytes(4), bocMagic) {
		return nil, errors.New("invalid boc magic header")
	}

	flags, cellNumSizeBytes := parseBOCFlags(r.MustReadByte()) // has_idx:(## 1) has_crc32c:(## 1)  has_cache_bits:(## 1) flags:(## 2) { flags = 0 } size:(## 3) { size <= 4 }
	dataSizeBytes := int(r.MustReadByte())                     // off_bytes:(## 8) { off_bytes <= 8 }

	cellsNum := dynInt(r.MustReadBytes(cellNumSizeBytes)) // cells:(##(size * 8))
	rootsNum := dynInt(r.MustReadBytes(cellNumSizeBytes)) // roots:(##(size * 8)) { roots >= 1 }

	// complete BOCs - ??? (absent:(##(size * 8)) { roots + absent <= cells })
	_ = r.MustReadBytes(cellNumSizeBytes)

	dataLen := dynInt(r.MustReadBytes(dataSizeBytes)) // tot_cells_size:(##(off_bytes * 8))

	// with checksum
	if flags.HasCrc32c {
		crc := crc32.Checksum(data[:len(data)-4], crc32.MakeTable(crc32.Castagnoli))
		if binary.LittleEndian.Uint32(data[len(data)-4:]) != crc {
			return nil, errors.New("checksum not matches")
		}
	}

	rootsIndex := make([]int, rootsNum)
	for i := 0; i < rootsNum; i++ {
		rootsIndex[i] = dynInt(r.MustReadBytes(cellNumSizeBytes))
	}

	if flags.hasCacheBits && !flags.hasIndex {
		return nil, fmt.Errorf("cache flag cant be set without index flag")
	}

	var index []int
	if flags.hasIndex {
		index = make([]int, 0, cellsNum)
		idxData, err := r.ReadBytes(cellsNum * dataSizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read custom index, err: %v", err)
		}

		for i := 0; i < cellsNum; i++ {
			off := i * dataSizeBytes
			val := dynInt(idxData[off : off+dataSizeBytes])
			if flags.hasCacheBits {
				// we don't need a cache, cause our loader uses memory
				val /= 2
			}
			index = append(index, val)
		}
	}

	if cellsNum > dataLen/2 {
		return nil, fmt.Errorf("cells num looks malicious: data len %d, cells %d", dataLen, cellsNum)
	}

	payload, err := r.ReadBytes(dataLen)
	if err != nil {
		return nil, fmt.Errorf("failed to read payload, want %d, has %d", dataLen, r.LeftLen())
	}

	cll, err := parseCells(rootsIndex, cellsNum, cellNumSizeBytes, payload, index)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}

	return cll, nil
}

func parseCells(rootsIndex []int, cellsNum, refSzBytes int, data []byte, index []int) ([]*Cell, error) {
	cells := make([]*Cell, cellsNum)
	for i := 0; i < cellsNum; i++ {
		// initialize them one by one for flexible gc and memory usage
		cells[i] = &Cell{}
	}

	// index = nil
	offset := 0
	for i := 0; i < cellsNum; i++ {
		if len(data)-offset < 2 {
			return nil, errors.New("failed to parse cell header, corrupted data")
		}

		if index != nil {
			// if we have index, then set offset from it, it stores end of each cell
			offset = 0
			if i > 0 {
				offset = index[i-1]
			}
		}

		// len(self.refs) + self.is_special() * 8 + self.level() * 32
		flags := data[offset]
		refsNum := int(flags & 0b111)
		special := (flags & 0b1000) != 0
		withHashes := (flags & 0b10000) != 0
		levelMask := LevelMask{flags >> 5}

		if refsNum > 4 {
			return nil, errors.New("too many refs in cell")
		}

		ln := data[offset+1]
		// round to 1 byte, len in octets
		oneMore := ln % 2
		sz := int(ln/2 + oneMore)

		offset += 2
		if len(data)-offset < sz {
			return nil, errors.New("failed to parse cell payload, corrupted data")
		}

		if withHashes {
			maskBits := int(math.Ceil(math.Log2(float64(levelMask.Mask) + 1)))
			hashesNum := maskBits + 1

			offset += hashesNum*hashSize + hashesNum*depthSize
			// TODO: check depth and hashes
		}

		payload := data[offset : offset+sz]

		offset += sz
		if len(data)-offset < refsNum*refSzBytes {
			return nil, errors.New("failed to parse cell refs, corrupted data")
		}

		refsIndex := make([]int, 0, refsNum)
		for y := 0; y < refsNum; y++ {
			refIndex := data[offset : offset+refSzBytes]

			refsIndex = append(refsIndex, dynInt(refIndex))
			offset += refSzBytes
		}

		refs := make([]*Cell, len(refsIndex))
		for y, id := range refsIndex {
			if i == id {
				return nil, errors.New("recursive reference of cells")
			}
			if id < i && index == nil { // compatibility with c++ implementation
				return nil, errors.New("reference to index which is behind parent cell")
			}
			if id >= len(cells) {
				return nil, errors.New("invalid index, out of scope")
			}

			refs[y] = cells[id]
		}

		bitsSz := uint(int(ln) * 4)

		// if not full byte
		if int(ln)%2 != 0 {
			// find last bit of byte which indicates the end and cut it and next
			for y := uint(0); y < 8; y++ {
				if (payload[len(payload)-1]>>y)&1 == 1 {
					bitsSz += 3 - y
					break
				}
			}
		}

		cells[i].special = special
		cells[i].bitsSz = bitsSz
		cells[i].levelMask = levelMask
		cells[i].data = payload
		cells[i].refs = refs
	}

	roots := make([]*Cell, len(rootsIndex))

	for i := len(cells) - 1; i >= 0; i-- {
		cells[i].calculateHashes()
	}

	for i, idx := range rootsIndex {
		roots[i] = cells[idx]
	}

	return roots, nil
}

func parseBOCFlags(data byte) (bocFlags, int) {
	return bocFlags{
		hasIndex:     data&(1<<7) > 0,
		HasCrc32c:    data&(1<<6) > 0,
		hasCacheBits: data&(1<<5) > 0,
	}, int(data & 0b00000111)
}

func dynInt(data []byte) int {
	tmp := make([]byte, 8)
	copy(tmp[8-len(data):], data)

	return int(binary.BigEndian.Uint64(tmp))
}






first function to jank
function parseBoc(src) {
    let reader = new BitReader_1.BitReader(new BitString_1.BitString(src, 0, src.length * 8));
    let magic = reader.loadUint(32);
    if (magic === 0x68ff65f3) {
        let size = reader.loadUint(8);
        let offBytes = reader.loadUint(8);
        let cells = reader.loadUint(size * 8);
        let roots = reader.loadUint(size * 8); // Must be 1
        let absent = reader.loadUint(size * 8);
        let totalCellSize = reader.loadUint(offBytes * 8);
        let index = reader.loadBuffer(cells * offBytes);
        let cellData = reader.loadBuffer(totalCellSize);
        return {
            size,
            offBytes,
            cells,
            roots,
            absent,
            totalCellSize,
            index,
            cellData,
            root: [0]
        };
    }
    else if (magic === 0xacc3a728) {
        let size = reader.loadUint(8);
        let offBytes = reader.loadUint(8);
        let cells = reader.loadUint(size * 8);
        let roots = reader.loadUint(size * 8); // Must be 1
        let absent = reader.loadUint(size * 8);
        let totalCellSize = reader.loadUint(offBytes * 8);
        let index = reader.loadBuffer(cells * offBytes);
        let cellData = reader.loadBuffer(totalCellSize);
        let crc32 = reader.loadBuffer(4);
        if (!(0, crc32c_1.crc32c)(src.subarray(0, src.length - 4)).equals(crc32)) {
            throw Error('Invalid CRC32C');
        }
        return {
            size,
            offBytes,
            cells,
            roots,
            absent,
            totalCellSize,
            index,
            cellData,
            root: [0]
        };
    }
    else if (magic === 0xb5ee9c72) {
        let hasIdx = reader.loadUint(1);
        let hasCrc32c = reader.loadUint(1);
        let hasCacheBits = reader.loadUint(1);
        let flags = reader.loadUint(2); // Must be 0
        let size = reader.loadUint(3);
        let offBytes = reader.loadUint(8);
        let cells = reader.loadUint(size * 8);
        let roots = reader.loadUint(size * 8);
        let absent = reader.loadUint(size * 8);
        let totalCellSize = reader.loadUint(offBytes * 8);
        let root = [];
        for (let i = 0; i < roots; i++) {
            root.push(reader.loadUint(size * 8));
        }
        let index = null;
        if (hasIdx) {
            index = reader.loadBuffer(cells * offBytes);
        }
        let cellData = reader.loadBuffer(totalCellSize);
        if (hasCrc32c) {
            let crc32 = reader.loadBuffer(4);
            if (!(0, crc32c_1.crc32c)(src.subarray(0, src.length - 4)).equals(crc32)) {
                throw Error('Invalid CRC32C');
            }
        }
        return {
            size,
            offBytes,
            cells,
            roots,
            absent,
            totalCellSize,
            index,
            cellData,
            root
        };
    }
    else {
        throw Error('Invalid magic');
    }
}
*/
