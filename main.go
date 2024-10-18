package main

import (
	"fmt"
	"log"
	"net/http"

	EvmHandler "github.com/laminafinance/crosschain-api/api/evm"
	InfoHandler "github.com/laminafinance/crosschain-api/api/info"
	Handler "github.com/laminafinance/crosschain-api/api/main"
	SvmHandler "github.com/laminafinance/crosschain-api/api/svm"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	http.HandleFunc("/api/main", Handler.Handler)
	http.HandleFunc("/api/info", InfoHandler.Handler)
	http.HandleFunc("/api/evm", EvmHandler.Handler)
	http.HandleFunc("/api/svm", SvmHandler.Handler)
	// http.HandleFunc("/", Handler.Handler) // tvm is currently in ts
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
