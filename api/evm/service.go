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

	return nil, nil
}

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
