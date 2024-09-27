package main

import (
	"fmt"
	EvmHandler "handler/api/evm"
	Handler "handler/api/main"
	SvmHandler "handler/api/svm"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	http.HandleFunc("/api/main", Handler.Handler)
	http.HandleFunc("/api/evm", EvmHandler.Handler)
	http.HandleFunc("/api/svm", SvmHandler.Handler)
	// http.HandleFunc("/", Handler.Handler) // tvm is currently in ts
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
