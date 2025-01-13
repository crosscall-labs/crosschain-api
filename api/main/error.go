package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/crosscall-labs/crosschain-api/pkg/utils"
)

func errUnsupportedChain(chainId string) string {
	return fmt.Sprintf("Chain ID %s not currently supported", chainId)
}

func errPaymasterAndDataMismatch(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    7,
		Message: "PaymasterAndData mismatch",
	})
}

func errRpcFailed(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    501,
		Message: "Internal server error: RPC connection failed",
	})
}

func errEscrowNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    1000,
		Message: "Escrow address not exist",
	})
}

func errInsufficientEscrowBalance(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&utils.Error{
		Code:    1001,
		Message: "Insufficient escrow balance",
	})
}

func GetOrigin() string {
	return utils.GetOrigin()
}
