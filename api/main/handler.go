package handler

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

// need query for creating an escrow lock
// this means that the query needs to call the escrow contract with the correct initializer data and salt
// need salt variable
// need initializer {eoaowner and delegate}, delegate is stored at:
//	storage slot 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc of the escrow contract

// need query for creating userop + scw initcode + paymasteranddata

var privateKey *ecdsa.PrivateKey
var relayAddress common.Address

func Handler(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	var err error

	saltHash := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000037")

	// Cast common.Hash to [32]byte
	var SALT [32]byte
	copy(SALT[:], saltHash[:]) // 55

	contractABIs := map[string]string{
		"Entrypoint":           contractAbiEntrypoint,
		"SimpleAccount":        contractAbiSimpleAccount,
		"SimpleAccountFactory": contractAbiSimpleAccountFactory,
		"Multicall":            contractAbiMulticall,
		"HyperlaneMailbox":     contractAbiHyperlaneMailbox,
		"HyperlaneIgp":         contractAbiHyperlaneIgp,
		"Paymaster":            contractAbiPaymaster,
		"Escrow":               contractAbiEscrow,
		"EscrowFactory":        contractAbiEscrowFactory,
	}

	parsedABIs := make(map[string]abi.ABI)

	for name, abiStr := range contractABIs {
		parsedABI, err := abi.JSON(strings.NewReader(abiStr))
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parsedABIs[name] = parsedABI
	}

	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Query().Get("query") {
	case "unsigned-message":
		response, err = UnsignedRequest(r)
	case "unsigned-bytecode":
		response, err = UnsignedBytecode(r)
	case "signed-bytecode":
		response, err = SignedBytecode(r)
	case "signed-escrow-payout":
		// will add env restriction on origin later
		response, err = SignedEscrowPayout(r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrMalformedRequest("Invalid query parameter"))
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	// Write successful JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}
