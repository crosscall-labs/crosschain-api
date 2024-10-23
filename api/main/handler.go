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
			utils.ErrInternal(w, err.Error())
			return
		}
		parsedABIs[name] = parsedABI
	}

	switch r.URL.Query().Get("query") {
	case "unsigned-message":
		UnsignedRequest(w, r)
		return
	case "unsigned-bytecode":
		UnsignedBytecode(w, r)
		return
	case "signed-bytecode":
		SignedBytecode(w, r)
		return
	case "signed-escrow-payout":
		// will add env restriction on origin later
		SignedEscrowPayout(w, r)
		return
	default:
		version := "Hello, World!"
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
