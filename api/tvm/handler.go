package tvmHandler

import (
	"encoding/json"
	"net/http"

	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	handlerWithCORS := utils.EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		var response interface{}
		var err error

		w.Header().Set("Content-Type", "application/json")
		switch query.Get("query") {
		// case "unsigned-escrow-request":
		// 	response, err = UnsignedEscrowRequest(r)
		// case "unsigned-entrypoint-request":
		// 	response, err = UnsignedEntryPointRequest(r)
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

		if err := json.NewEncoder(w).Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	}))

	handlerWithCORS.ServeHTTP(w, r)
}
