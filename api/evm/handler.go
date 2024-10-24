package evmHandler

import (
	"encoding/json"
	"net/http"

	"github.com/laminafinance/crosschain-api/pkg/utils"
)

// Error data structure
type Error struct {
	Code    uint64 `json:"code"`
	Message string `json:"message"`
}

type SpotResponse struct {
	Asset0   string `json:"asset0"`
	Asset1   string `json:"asset1"`
	AmountIn string `json:"amount-in"`
	Spot     string `json:"spot-price"`
	SpotMin  string `json:"spot-price-min"`
	SpotMax  string `json:"spot-price-max"`
}

type OrderbookResponse struct {
	RetCode    int           `json:"retCode"`
	RetMsg     string        `json:"retMsg"`
	Result     OrderbookData `json:"result"`
	RetExtInfo struct{}      `json:"retExtInfo"`
	Time       int64         `json:"time"`
}

type OrderbookData struct {
	Symbol    string     `json:"s"`
	Asks      [][]string `json:"a"`
	Bids      [][]string `json:"b"`
	Timestamp int64      `json:"ts"`
	UpdateID  int        `json:"u"`
	CrossSeq  int        `json:"seq"`
}

type HelloWorld struct {
	Test string `json:"test"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	handlerWithCORS := utils.EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		switch query.Get("query") {
		case "unsigned-escrow-request":
			UnsignedEscrowRequest(w, r)
			return
		case "test":
			var hellowWorld HelloWorld
			hellowWorld.Test = "Hello, World!"

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(hellowWorld); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Invalid query parameter", http.StatusBadRequest)
			return
		}
	}))

	handlerWithCORS.ServeHTTP(w, r)
}
