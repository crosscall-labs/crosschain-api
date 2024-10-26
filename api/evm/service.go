package evmHandler

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

// It accepts an optional query parameter for internal calls.
func UnsignedEscrowRequest(w http.ResponseWriter, r *http.Request, parameters ...*UnsignedEscrowRequestParams) (interface{}, error) {
	var params *UnsignedEscrowRequestParams
	//var err Error

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &UnsignedEscrowRequestParams{}
	}

	if r != nil {
		if !utils.ParseAndValidateParams(w, r, &params) {
			return nil, fmt.Errorf("%s", "http.Request is required")
		}
		if w == nil {
			utils.ErrInternal(w, "http.Request is required")
			return nil, nil
		}
	}

	var errorStr string
	params.Header.ChainId, params.Header.ChainType, params.Header.ChainName, errorStr = utils.CheckChainPartialType(params.Header.ChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		if r == nil {
			return nil, fmt.Errorf("%s", errorStr)
		} else {
			utils.ErrMalformedRequest(w, errorStr)
			return nil, nil
		}
	}

	// Encode the response and write it to the ResponseWriter
	// if err := json.NewEncoder(w).Encode(unsignedDataResponse); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	return nil, nil
}

func UnsignedEntryPointRequest(w http.ResponseWriter, r *http.Request, parameters ...*UnsignedEntryPointRequestParams) (interface{}, error) {
	var params *UnsignedEntryPointRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &UnsignedEntryPointRequestParams{}
	}

	if r != nil {
		if !utils.ParseAndValidateParams(w, r, &params) {
			return nil, fmt.Errorf("%s", "http.Request is required")
		}
		if w == nil {
			utils.ErrInternal(w, "http.Request is required")
			return nil, nil
		}
	}

	var errorStr string
	params.Header.FromChainId, params.Header.FromChainType, params.Header.FromChainName, errorStr = utils.CheckChainPartialType(params.Header.FromChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		if r == nil {
			return nil, fmt.Errorf("%s", errorStr)
		} else {
			utils.ErrMalformedRequest(w, errorStr)
			return nil, nil
		}
	}
	params.Header.ToChainId, params.Header.ToChainType, params.Header.ToChainName, errorStr = utils.CheckChainPartialType(params.Header.ToChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		if r == nil {
			return nil, fmt.Errorf("%s", errorStr)
		} else {
			utils.ErrMalformedRequest(w, errorStr)
			return nil, nil
		}
	}

	payload, err := utils.Str2Bytes(params.Payload)
	if err != nil {
		utils.ErrMalformedRequest(w, err.Error())
	}
	fmt.Print(payload)

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
	// 	create default values for packed user op
	//	create default values for paymaster and data
	//	create default values for calldata (this should be done by the protocol api since we don't want to delegate using a specifc wallet architecture)
	//		test data will be using an empty value sent as if it were thorugh signer -> simpleAccount proxy
	//	combine the transaction gas and cost for execution then multiply by 0.1%, this should be our crosschain fee + bid fee
	// 		add this value to the paymaster and data AND PriceGwei

	packedUserOperation := GenerateTestPackedUserOperation()
	packedUserOperationResponse, _ := ToPackedUserOperationResponse(packedUserOperation)
	utils.PrintStructFields(packedUserOperationResponse)

	return nil, nil
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
