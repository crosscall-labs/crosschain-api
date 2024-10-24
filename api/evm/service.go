package evmHandler

import (
	"net/http"

	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func UnsignedEscrowRequest(w http.ResponseWriter, r *http.Request) {
	params := &UnsignedEscrowRequestParams{}

	if !utils.ParseAndValidateParams(w, r, params) {
		return
	}

	var errorStr string
	params.Header.FromChainId, params.Header.FromChainType, params.Header.FromChainName, errorStr = utils.CheckChainPartialType(params.Header.FromChainId, "escrow", params.Header.TxType)
	if errorStr != "" {
		utils.ErrMalformedRequest(w, errorStr)
		return
	}
}

func UnsignedEntryPointRequest(w http.ResponseWriter, r *http.Request) {
	params := &UnsignedEntryPointRequestParams{}

	if !utils.ParseAndValidateParams(w, r, params) {
		return
	}

	var errorStr string
	params.Header.FromChainId, params.Header.FromChainType, params.Header.FromChainName, errorStr = utils.CheckChainPartialType(params.Header.FromChainId, "entrypoint", params.Header.TxType)
	if errorStr != "" {
		utils.ErrMalformedRequest(w, errorStr)
		return
	}
}
