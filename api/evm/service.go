package evmHandler

import (
	"fmt"
	"net/http"

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
	params.Header.ChainId, params.Header.ChainType, params.Header.ChainName, errorStr = utils.CheckChainPartialType(params.Header.ChainId, "entrypoint", params.Header.TxType)
	if errorStr != "" {
		if r == nil {
			return nil, fmt.Errorf("%s", errorStr)
		} else {
			utils.ErrMalformedRequest(w, errorStr)
			return nil, nil
		}
	}

	return nil, nil
}
