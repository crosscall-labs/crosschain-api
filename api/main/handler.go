package handler

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/laminafinance/crosschain-api/pkg/db"
	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/supabase-community/supabase-go"
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
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("\nRecovered from panic: %v", rec)

			supabaseUrl := os.Getenv("SUPABASE_URL")
			supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
			supabaseClient, err := supabase.NewClient(supabaseUrl, supabaseKey, nil)
			if err == nil {
				logErr := db.LogPanic(supabaseClient, fmt.Sprintf("%v", rec), nil)
				if logErr != nil {
					log.Printf("\nFailed to log panic to Supabase: %v", logErr)
				}
			} else {
				log.Printf("\nFailed to create Supabase client for panic logging: %v", err)
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	var response interface{}
	var err error
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	supabaseClient, err := supabase.NewClient(supabaseUrl, supabaseKey, nil)
	if err != nil {
		http.Error(w, "Failed to create Supabase client", http.StatusInternalServerError)
		return
	}

	saltHash := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000037")

	// Cast common.Hash to [32]byte
	var SALT [32]byte
	copy(SALT[:], saltHash[:]) // 55

	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Query().Get("query") {
	case "unsigned-message":
		response, err = UnsignedRequest(r)
		HandleResponse(w, r, supabaseClient, response, err)
		return
	case "unsigned-bytecode":
		response, err = UnsignedBytecode(r)
		HandleResponse(w, r, supabaseClient, response, err)
		return
	case "signed-bytecode":
		response, err = SignedBytecode(r)
		HandleResponse(w, r, supabaseClient, response, err)
		return
	case "signed-escrow-payout":
		// will add env restriction on origin later
		response, err = SignedEscrowPayout(r)
		HandleResponse(w, r, supabaseClient, response, err)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrMalformedRequest("Invalid query parameter"))
		return
	}
}

func HandleResponse(w http.ResponseWriter, r *http.Request, supabaseClient *supabase.Client, response interface{}, err error) {
	if err != nil {
		// if logErr := db.LogError(supabaseClient, err, r.URL.Query().Get("query"), response); logErr != nil {
		// 	fmt.Printf("Failed to log error: %v\n", logErr.Error())
		// }

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}
