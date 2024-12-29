package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	EvmHandler "github.com/laminafinance/crosschain-api/api/evm"
	InfoHandler "github.com/laminafinance/crosschain-api/api/info"
	Handler "github.com/laminafinance/crosschain-api/api/main"
	RequestHandler "github.com/laminafinance/crosschain-api/api/request"
	SvmHandler "github.com/laminafinance/crosschain-api/api/svm"
	TvmHandler "github.com/laminafinance/crosschain-api/api/tvm"
	"github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

func main() {
	serverEnv := flag.String("server", "production", "Specify the server environment (local/production)")
	flag.Parse()

	var envFile string
	if *serverEnv == "local" {
		envFile = ".env.local"
	} else {
		envFile = ".env"
	}

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	debugMode := os.Getenv("DEBUG_MODE_ENABLED")
	if debugMode == "true" || debugMode == "1" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	http.HandleFunc("/api/main", Handler.Handler)
	http.HandleFunc("/api/info", InfoHandler.Handler)
	http.HandleFunc("/api/evm", EvmHandler.Handler)
	http.HandleFunc("/api/svm", SvmHandler.Handler)
	http.HandleFunc("/api/tvm", TvmHandler.Handler)
	http.HandleFunc("/api/request", RequestHandler.Handler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
