package svmHandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	handlerWithCORS := EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		switch query.Get("query") {
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

func getSymbol(asset0, asset1 string) string {
	if strings.ToUpper(asset0+asset1) == "BTCETH" {
		return "ETHBTC"
	}
	return strings.ToUpper(asset0 + asset1)
}

func errMalformedRequest(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&Error{
		Code:    400,
		Message: "Malformed request",
	})
}

func errInternal(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&Error{
		Code:    500,
		Message: "Internal server error",
	})
}

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Printf("Method: %s, URL: %s", r.Method, r.URL)

		next.ServeHTTP(w, r)
	})
}
