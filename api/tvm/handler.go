package tvmHandler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/laminafinance/crosschain-api/pkg/db"
	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/supabase-community/supabase-go"
)

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

	handlerWithCORS := utils.EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		var response interface{}
		var err error
		supabaseUrl := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		supabaseClient, err := supabase.NewClient(supabaseUrl, supabaseKey, nil)
		if err != nil {
			http.Error(w, "Failed to create Supabase client", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		switch query.Get("query") {
		case "unsigned-escrow-request":
			response, err = UnsignedEscrowRequest(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "unsigned-entrypoint-request":
			response, err = UnsignedEntryPointRequest(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "asset-info":
			response, err = AssetInfoRequest(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "asset-mint":
			response, err = AssetMintRequest(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		// case "user-info":
		// 	response, err = UnsignedEntryPointRequest(r)
		// 	HandleResponse(w, r, supabaseClient, response, err)
		// 	return
		case "test": // deploy
			response, err = TestRequest(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "test2": // view
			response, err = Test2Request(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "test3": // execute
			response, err = Test3Request(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "test4": // deploy + execute
			response, err = Test4Request(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "test5": // tonx view
			response, err = Test5Request(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		case "test6": // depply jetton via tonutils-go
			response, err = Test6Request(r)
			HandleResponse(w, r, supabaseClient, response, err)
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(utils.ErrMalformedRequest("Invalid query parameter"))
			return
		}

	}))

	handlerWithCORS.ServeHTTP(w, r)
}

func HandleResponse(w http.ResponseWriter, r *http.Request, supabaseClient *supabase.Client, response interface{}, err error) {
	if err != nil {
		// if logErr := db.LogError(supabaseClient, err, r.URL.Query().Get("query"), response); logErr != nil {
		// 	fmt.Printf("Failed to log error: %v\n", logErr.Error())
		// }

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ErrInternal(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}
