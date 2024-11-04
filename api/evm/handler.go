package evmHandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/laminafinance/crosschain-api/pkg/db"
	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/supabase-community/supabase-go"
)

func Handler(w http.ResponseWriter, r *http.Request) {
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
		case "unsigned-entrypoint-request":
			response, err = UnsignedEntryPointRequest(r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(utils.ErrMalformedRequest("Invalid query parameter"))
			return
		}

		if err != nil {
			if logErr := db.LogError(supabaseClient, err, r.URL.Query().Get("query"), response); logErr != nil {
				fmt.Printf("Failed to log error: %v\n", logErr.Error())
			}

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	}))

	handlerWithCORS.ServeHTTP(w, r)
}
