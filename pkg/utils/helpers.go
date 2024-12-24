package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func WriteJSONResponse(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ParseAndValidateParams(r *http.Request, params interface{}) error {
	val := reflect.ValueOf(params).Elem() // Dereference the pointer to access the underlying struct
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	typ := val.Type()

	missingFields := []string{}
	allowedFields := make(map[string]struct{})

	fmt.Printf("\nfieldType.Type: %v", typ)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		queryTag := fieldType.Tag.Get("query")
		optionalTag := fieldType.Tag.Get("optional")

		if queryTag != "" {
			allowedFields[queryTag] = struct{}{}
		}

		// fmt.Printf("\nfieldType.Name: %v", fieldType.Name)
		if _, exists := typ.FieldByName(fieldType.Name); exists {
			if field.Kind() == reflect.Struct {
				// Recursively parse nested struct fields
				nestedParams := reflect.New(fieldType.Type).Interface()
				if err := ParseAndValidateParams(r, nestedParams); err != nil {
					return err
				}
				// After recursion, set the original struct's field value
				field.Set(reflect.ValueOf(nestedParams).Elem())
			} else if queryTag != "" {
				queryValue := r.URL.Query().Get(queryTag)

				// If the field is required (i.e., optional is not set to "true")
				if queryValue == "" && optionalTag != "true" {
					missingFields = append(missingFields, queryTag)
				} else if queryValue != "" {
					field.SetString(queryValue)
				}
			}
		}
	}

	// If there are missing fields, return an error response
	if len(missingFields) > 0 {
		return ErrMalformedRequest(fmt.Sprint("Missing fields: " + strings.Join(missingFields, ", ")))
	}

	return nil
}

func PrintStructFields(params interface{}) {
	val := reflect.ValueOf(params)

	// Ensure the value is a struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		fmt.Println("Expected a struct")
		return
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		fieldName := fieldType.Name

		// Check if it's a nested struct
		if field.Kind() == reflect.Struct {
			fmt.Printf("\n%s:\n", fieldName)
			PrintStructFields(field.Interface()) // Recursively print nested struct fields
			fmt.Println()
		} else {
			fmt.Printf("\n%s: %v", fieldName, field.Interface()) // Print field value
		}
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("Error (Code: %d, Message: %s)", e.Code, e.Message)
}

// func ErrMalformedRequest(w http.ResponseWriter, message string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusBadRequest)

// 	json.NewEncoder(w).Encode(&Error{
// 		Code:    400,
// 		Message: "Malformed request",
// 		Details: message,
// 	})
// }

func GetOrigin() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	funcName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(funcName, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return "unknown"
}

func ErrMalformedRequest(message string) error {
	origin := GetOrigin()

	return Error{
		Code:    400,
		Message: "Malformed request",
		Details: message,
		Origin:  origin,
	}
}

func ErrInternal(message string) Error {
	origin := GetOrigin()

	return Error{
		Code:    500,
		Message: "Internal server error",
		Details: message,
		Origin:  origin,
	}
}

func EnvKey2Ecdsa() (*ecdsa.PrivateKey, common.Address, error) {
	return PrivateKey2Sepc256k1(os.Getenv("RELAY_PRIVATE_KEY"))
}

func Key2Ecdsa(key string) (*ecdsa.PrivateKey, common.Address, error) {
	return PrivateKey2Sepc256k1(key)
}

func PrivateKey2Sepc256k1(privateKeyString string) (privateKey *ecdsa.PrivateKey, publicAddress common.Address, err error) {
	privateKey, err = crypto.HexToECDSA(privateKeyString)
	if err != nil {
		err = fmt.Errorf("Error converting private key: %v", err)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err = fmt.Errorf("Error casting public key to ECDSA")
		return
	}

	publicAddress = crypto.PubkeyToAddress(*publicKeyECDSA)
	return
}

func Str2Bytes(hexStr string) ([]byte, error) {
	if hexStr == "" {
		return []byte{}, nil // Return empty byte slice for empty input
	}

	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("invalid payload: length odd")
	}

	for _, r := range hexStr {
		if _, err := strconv.ParseUint(string(r), 16, 8); err != nil {
			return nil, fmt.Errorf("invalid payload: hex only 0123456789abcdefABCDEF")
		}
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func HasInt(inputArray []int, input int) error {
	for _, value := range inputArray {
		if value == input {
			return nil
		}
	}
	return fmt.Errorf("int not found in input array")
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
		fmt.Printf("\nMethod: %s, URL: %s", r.Method, r.URL)

		next.ServeHTTP(w, r)
	})
}

func CheckChainPartialType(chainId, partialType, txType string) (string, string, string, string) {
	chainIdOut, chainType, chainName, entrypointTypes, escrowTypes, errorStr := CheckChainType(chainId)
	if errorStr != "" {
		return "", "", "", errorStr
	}
	txTypeInt, err := strconv.Atoi(txType)
	if err != nil {
		return "", "", "", "invalid txtype"
	}

	partialTypeMap := map[string][]int{
		"escrow":     escrowTypes,
		"entrypoint": entrypointTypes,
	}

	types, ok := partialTypeMap[partialType]
	if !ok {
		return "", "", "", "partialType not set"
	}
	if err := HasInt(types, txTypeInt); err != nil {
		return "", "", "", fmt.Sprintf("Chain %s missing type %s for %s", chainIdOut, txType, partialType)
	}

	return chainIdOut, chainType, chainName, ""
}

// case "0x310C5", "200901": // bitlayer mainnet
// case "0xE35", "3637": // botanix mainnet
// case "0xC4", "196": // x layer mainnet
// case "0xA4EC", "42220": // celo mainnet
// case "0x82750", "534352": // scroll mainnet
// case "0xA", "10": // op mainnet
// case "0xA4B1", "42161": // arbitrum one
// case "0x2105", "8453": // base mainnet
// case "0x13A", "314": // filecoin mainnet
// case "0x63630000", "1667432448": // tvm workchain_id == 0
// case "0x53564D0001", "357930172419": // solana mainnet
// case "0xBF04", "48900": // zircuit mainnet
func CheckChainType(chainId string) (string, string, string, []int, []int, string) { // out: id, vm, name, escrowType, entrypointType, error
	disabled := fmt.Sprintf("unsupported chain ID: %s", chainId)
	switch chainId {
	case "0x3106A", "200810": // bitlayer testnet
		return "200810", "evm", "bitlayerTestnet", []int{0, 1}, []int{0, 1, 2}, ""
	case "0x4268", "17000": // holesky
		return "17000", "evm", "ethereumHoleskyTestnet", []int{0, 1}, []int{0, 1, 2}, ""
	case "0xAA36A7", "11155111": // sepolia
		return "11155111", "evm", "ethereumSepoliaTestnet", []int{0, 1}, []int{0, 1, 2}, ""
	case "0xE34", "3636": // botanix testnet
		return "3636", "evm", "botanixTestnet", []int{0, 1}, []int{0, 1, 2}, disabled
	case "0xF35A", "62298": // citrea testnet
		return "62298", "evm", "citreaTestnet", []int{0, 1}, []int{0, 1, 2}, ""
	case "0x13881", "80001": // matic mumbai
		return "80001", "evm", "maticMumbai", nil, nil, disabled
	case "0x13882", "80002": // matic amoy
		return "80002", "evm", "maticAmoy", nil, nil, disabled
	case "0xC3", "195": // x layer testnet
		return "195", "evm", "xLayerEvmTestnet", nil, nil, disabled
	case "0xAEF3", "44787": // celo alfajores
		return "44787", "evm", "celoAlforesTestnet", nil, nil, disabled
	case "0x5E9", "1513": // story testnet
		return "1513", "evm", "storyEvmTestnet", nil, nil, disabled
	case "0x8274F", "534351": // scroll testnet
		return "534351", "evm", "scrollEvmTestnet", nil, nil, disabled
	case "0xAA37DC", "11155420": // op sepolia
		return "11155420", "evm", "optimismSepoliaTestnet", nil, nil, disabled
	case "0x66EEE", "421614": // arbitrum sepolia
		return "421614", "evm", "arbitrumSepoliaTestnet", nil, nil, disabled
	case "0x14A34", "84532": // base sepolia
		return "84532", "evm", "baseSepoliaTestnet", nil, nil, disabled
	case "0x4CB2F", "314159": // filecoin calibration
		return "314159", "evm", "filecoinEvmTestnet", nil, nil, disabled
	case "0xBF03", "48899": // zircuit testnet
		return "48899", "evm", "zircuitTestnet", nil, nil, disabled
	case "0x63639999", "1667471769": // tvm workchain_id == -1 ton testnet
		return "1667471769", "tvm", "tonTvmTestnet", []int{2}, []int{0, 1, 2}, ""
	case "0x53564D0002", "357930172418": // solana devnet
		return "357930172418", "svm", "solanaSvmDevnet", nil, nil, disabled
	case "0x53564D0003", "357930172419": // solana testnet
		return "357930172419", "svm", "solanaSvmTestnet", nil, nil, disabled
	case "0x53564D0004", "357930172420": // eclipse (solana) testnet
		return "357930172420", "svm", "eclipseSvmTestnet", nil, nil, disabled
	default:
		return "", "", "", nil, nil, disabled
	}
}

var disabled = func(chainId string) error {
	return ErrMalformedRequest(fmt.Sprintf("unsupported chain ID: %s", chainId))
}

// this will need to later be added to db
// var chainInfoMap = map[string]struct {
// 	Info  ChainInfo
// 	Error error
// }{
// 	"0x3106A":      {Info: ChainInfo{"200810", "evm", "Bitlayer Testnet", "0000000000000000000000000000000000000000", "BTC", "18"}, Error: nil},
// 	"200810":       {Info: ChainInfo{"200810", "evm", "Bitlayer Testnet", "0000000000000000000000000000000000000000", "BTC", "18"}, Error: nil},
// 	"0x4268":       {Info: ChainInfo{"17000", "evm", "Holesky Testnet", "0000000000000000000000000000000000000000", "ETH", "18"}, Error: nil},
// 	"17000":        {Info: ChainInfo{"17000", "evm", "Holesky Testnet", "0000000000000000000000000000000000000000", "ETH", "18"}, Error: nil},
// 	"0xAA36A7":     {Info: ChainInfo{"11155111", "evm", "Sepolia Testnet", "0000000000000000000000000000000000000000", "ETH", "18"}, Error: nil},
// 	"11155111":     {Info: ChainInfo{"11155111", "evm", "Sepolia Testnet", "0000000000000000000000000000000000000000", "ETH", "18"}, Error: nil},
// 	"0xE34":        {Info: ChainInfo{"3636", "evm", "Botanix Testnet", "0000000000000000000000000000000000000000", "BTC", "18"}, Error: disabled("3636")},
// 	"3636":         {Info: ChainInfo{"3636", "evm", "Botanix Testnet", "0000000000000000000000000000000000000000", "BTC", "18"}, Error: disabled("3636")},
// 	"0xF35A":       {Info: ChainInfo{"62298", "evm", "Citrea Testnet", "0000000000000000000000000000000000000000", "BTC", "18"}, Error: nil},
// 	"62298":        {Info: ChainInfo{"62298", "evm", "Citrea Testnet", "0000000000000000000000000000000000000000", "BTC", "18"}, Error: nil},
// 	"0x13881":      {Info: ChainInfo{"80001", "evm", "Matic Mumbai", "0000000000000000000000000000000000000000", "MATIC", "18"}, Error: disabled("80001")},
// 	"80001":        {Info: ChainInfo{"80001", "evm", "Matic Mumbai", "0000000000000000000000000000000000000000", "MATIC", "18"}, Error: disabled("80001")},
// 	"0x13882":      {Info: ChainInfo{"80002", "evm", "Matic Amoy", "0000000000000000000000000000000000000000", "MATIC", "18"}, Error: disabled("80002")},
// 	"80002":        {Info: ChainInfo{"80002", "evm", "Matic Amoy", "0000000000000000000000000000000000000000", "MATIC", "18"}, Error: disabled("80002")},
// 	"0x63639999":   {Info: ChainInfo{"1667471769", "tvm", "Ton Testnet", "0000000000000000000000000000000000000000", "TON", "9"}, Error: nil},
// 	"1667471769":   {Info: ChainInfo{"1667471769", "tvm", "Ton Testnet", "0000000000000000000000000000000000000000", "TON", "9"}, Error: nil},
// 	"0x53564D0002": {Info: ChainInfo{"357930172418", "svm", "Solana Devnet", "0000000000000000000000000000000000000000", "SOL", "9"}, Error: disabled("357930172418")},
// 	"357930172418": {Info: ChainInfo{"357930172418", "svm", "Solana Devnet", "0000000000000000000000000000000000000000", "SOL", "18"}, Error: disabled("357930172418")},
// 	"0x53564D0003": {Info: ChainInfo{"357930172419", "svm", "Solana Testnet", "0000000000000000000000000000000000000000", "SOL", "18"}, Error: disabled("357930172419")},
// 	"357930172419": {Info: ChainInfo{"357930172419", "svm", "Solana Testnet", "0000000000000000000000000000000000000000", "SOL", "18"}, Error: disabled("357930172419")},
// 	"0x53564D0004": {Info: ChainInfo{"357930172420", "svm", "Eclipse Testnet", "0000000000000000000000000000000000000000", "SOL", "18"}, Error: disabled("357930172420")},
// 	"357930172420": {Info: ChainInfo{"357930172420", "svm", "Eclipse Testnet", "0000000000000000000000000000000000000000", "SOL", "18"}, Error: disabled("357930172420")},
// }

// func GetChainInfo(chainId string) (ChainInfo, error) {
// 	if info, found := chainInfoMap[chainId]; found {
// 		if info.Error != nil {
// 			return ChainInfo{}, info.Error
// 		}
// 		return info.Info, nil
// 	}

// 	return ChainInfo{}, fmt.Errorf("Unsupporting Chain ID: %s", chainId)
// }

// var chainInfoMap = map[string]ChainInfo{
// 	"0x3106A":      {"200810", "evm", "bitlayerTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"200810":       {"200810", "evm", "bitlayerTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"0x4268":       {"17000", "evm", "ethereumHoleskyTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"17000":        {"17000", "evm", "ethereumHoleskyTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"0xAA36A7":     {"11155111", "evm", "ethereumSepoliaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"11155111":     {"11155111", "evm", "ethereumSepoliaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"0xE34":        {"3636", "evm", "botanixTestnet", []int{0, 1}, []int{0, 1, 2}, disabled("3636")},
// 	"3636":         {"3636", "evm", "botanixTestnet", []int{0, 1}, []int{0, 1, 2}, disabled("3636")},
// 	"0xF35A":       {"62298", "evm", "citreaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"62298":        {"62298", "evm", "citreaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
// 	"0x13881":      {"80001", "evm", "maticMumbai", nil, nil, disabled("80001")},
// 	"80001":        {"80001", "evm", "maticMumbai", nil, nil, disabled("80001")},
// 	"0x13882":      {"80002", "evm", "maticAmoy", nil, nil, disabled("80002")},
// 	"80002":        {"80002", "evm", "maticAmoy", nil, nil, disabled("80002")},
// 	"0xC3":         {"195", "evm", "xLayerEvmTestnet", nil, nil, disabled("195")},
// 	"195":          {"195", "evm", "xLayerEvmTestnet", nil, nil, disabled("195")},
// 	"0xAEF3":       {"44787", "evm", "celoAlforesTestnet", nil, nil, disabled("44787")},
// 	"44787":        {"44787", "evm", "celoAlforesTestnet", nil, nil, disabled("44787")},
// 	"0x5E9":        {"1513", "evm", "storyEvmTestnet", nil, nil, disabled("1513")},
// 	"1513":         {"1513", "evm", "storyEvmTestnet", nil, nil, disabled("1513")},
// 	"0x8274F":      {"534351", "evm", "scrollEvmTestnet", nil, nil, disabled("534351")},
// 	"534351":       {"534351", "evm", "scrollEvmTestnet", nil, nil, disabled("534351")},
// 	"0xAA37DC":     {"11155420", "evm", "optimismSepoliaTestnet", nil, nil, disabled("11155420")},
// 	"11155420":     {"11155420", "evm", "optimismSepoliaTestnet", nil, nil, disabled("11155420")},
// 	"0x66EEE":      {"421614", "evm", "arbitrumSepoliaTestnet", nil, nil, disabled("421614")},
// 	"421614":       {"421614", "evm", "arbitrumSepoliaTestnet", nil, nil, disabled("421614")},
// 	"0x14A34":      {"84532", "evm", "baseSepoliaTestnet", nil, nil, disabled("84532")},
// 	"84532":        {"84532", "evm", "baseSepoliaTestnet", nil, nil, disabled("84532")},
// 	"0x4CB2F":      {"314159", "evm", "filecoinEvmTestnet", nil, nil, disabled("314159")},
// 	"314159":       {"314159", "evm", "filecoinEvmTestnet", nil, nil, disabled("314159")},
// 	"0xBF03":       {"48899", "evm", "zircuitTestnet", nil, nil, disabled("48899")},
// 	"48899":        {"48899", "evm", "zircuitTestnet", nil, nil, disabled("48899")},
// 	"0x63639999":   {"1667471769", "tvm", "tonTvmTestnet", []int{2}, []int{0, 1, 2}, nil},
// 	"1667471769":   {"1667471769", "tvm", "tonTvmTestnet", []int{2}, []int{0, 1, 2}, nil},
// 	"0x53564D0002": {"357930172418", "svm", "solanaSvmDevnet", nil, nil, disabled("357930172418")},
// 	"357930172418": {"357930172418", "svm", "solanaSvmDevnet", nil, nil, disabled("357930172418")},
// 	"0x53564D0003": {"357930172419", "svm", "solanaSvmTestnet", nil, nil, disabled("357930172419")},
// 	"357930172419": {"357930172419", "svm", "solanaSvmTestnet", nil, nil, disabled("357930172419")},
// 	"0x53564D0004": {"357930172420", "svm", "eclipseSvmTestnet", nil, nil, disabled("357930172420")},
// 	"357930172420": {"357930172420", "svm", "eclipseSvmTestnet", nil, nil, disabled("357930172420")},
// }

// func CheckChainType2(chainId string) (string, string, string, []int, []int, string) {
// 	if info, found := chainInfoMap[chainId]; found {
// 		return info.ID, info.VM, info.Name, info.EscrowType, info.EntrypointType, ""
// 	}
// 	disabled := fmt.Sprintf("unsupported chain ID: %s", chainId)
// 	return "", "", "", nil, nil, disabled
// }

var chainClientMap = map[string]ChainInfo{
	"0x3106A":      {"200810", "evm", "bitlayerTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"200810":       {"200810", "evm", "bitlayerTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"0x4268":       {"17000", "evm", "ethereumHoleskyTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"17000":        {"17000", "evm", "ethereumHoleskyTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"0xAA36A7":     {"11155111", "evm", "ethereumSepoliaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"11155111":     {"11155111", "evm", "ethereumSepoliaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"0xE34":        {"3636", "evm", "botanixTestnet", []int{0, 1}, []int{0, 1, 2}, disabled("3636")},
	"3636":         {"3636", "evm", "botanixTestnet", []int{0, 1}, []int{0, 1, 2}, disabled("3636")},
	"0xF35A":       {"62298", "evm", "citreaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"62298":        {"62298", "evm", "citreaTestnet", []int{0, 1}, []int{0, 1, 2}, nil},
	"0x13881":      {"80001", "evm", "maticMumbai", nil, nil, disabled("80001")},
	"80001":        {"80001", "evm", "maticMumbai", nil, nil, disabled("80001")},
	"0x13882":      {"80002", "evm", "maticAmoy", nil, nil, disabled("80002")},
	"80002":        {"80002", "evm", "maticAmoy", nil, nil, disabled("80002")},
	"0xC3":         {"195", "evm", "xLayerEvmTestnet", nil, nil, disabled("195")},
	"195":          {"195", "evm", "xLayerEvmTestnet", nil, nil, disabled("195")},
	"0xAEF3":       {"44787", "evm", "celoAlforesTestnet", nil, nil, disabled("44787")},
	"44787":        {"44787", "evm", "celoAlforesTestnet", nil, nil, disabled("44787")},
	"0x5E9":        {"1513", "evm", "storyEvmTestnet", nil, nil, disabled("1513")},
	"1513":         {"1513", "evm", "storyEvmTestnet", nil, nil, disabled("1513")},
	"0x8274F":      {"534351", "evm", "scrollEvmTestnet", nil, nil, disabled("534351")},
	"534351":       {"534351", "evm", "scrollEvmTestnet", nil, nil, disabled("534351")},
	"0xAA37DC":     {"11155420", "evm", "optimismSepoliaTestnet", nil, nil, disabled("11155420")},
	"11155420":     {"11155420", "evm", "optimismSepoliaTestnet", nil, nil, disabled("11155420")},
	"0x66EEE":      {"421614", "evm", "arbitrumSepoliaTestnet", nil, nil, disabled("421614")},
	"421614":       {"421614", "evm", "arbitrumSepoliaTestnet", nil, nil, disabled("421614")},
	"0x14A34":      {"84532", "evm", "baseSepoliaTestnet", nil, nil, disabled("84532")},
	"84532":        {"84532", "evm", "baseSepoliaTestnet", nil, nil, disabled("84532")},
	"0x4CB2F":      {"314159", "evm", "filecoinEvmTestnet", nil, nil, disabled("314159")},
	"314159":       {"314159", "evm", "filecoinEvmTestnet", nil, nil, disabled("314159")},
	"0xBF03":       {"48899", "evm", "zircuitTestnet", nil, nil, disabled("48899")},
	"48899":        {"48899", "evm", "zircuitTestnet", nil, nil, disabled("48899")},
	"0x63639999":   {"1667471769", "tvm", "tonTvmTestnet", []int{2}, []int{0, 1, 2}, nil},
	"1667471769":   {"1667471769", "tvm", "tonTvmTestnet", []int{2}, []int{0, 1, 2}, nil},
	"0x53564D0002": {"357930172418", "svm", "solanaSvmDevnet", nil, nil, disabled("357930172418")},
	"357930172418": {"357930172418", "svm", "solanaSvmDevnet", nil, nil, disabled("357930172418")},
	"0x53564D0003": {"357930172419", "svm", "solanaSvmTestnet", nil, nil, disabled("357930172419")},
	"357930172419": {"357930172419", "svm", "solanaSvmTestnet", nil, nil, disabled("357930172419")},
	"0x53564D0004": {"357930172420", "svm", "eclipseSvmTestnet", nil, nil, disabled("357930172420")},
	"357930172420": {"357930172420", "svm", "eclipseSvmTestnet", nil, nil, disabled("357930172420")},
}

// func checkChainStatus(chainId string) (*ethclient.Client, *Chain, error) {
// 	var client *ethclient.Client
// 	var chain *Chain
// 	var err error

// 	var rpcURL string
// 	var addresses Chain

// 	switch chainId {
// 	case "0x3106A", "200810":
// 		chainId = "200810"
// 		rpcURL = "https://testnet-rpc.bitlayer.org"
// 		addresses = Chain{
// 			AddressEntrypoint:            "0x317bBdFbAe7845648864348A0C304392d0F2925F",
// 			AddressEntrypointSimulations: "0x6960fA06d5119258533B5d715c8696EE66ca4042",
// 			AddressSimpleAccountFactory:  "0xCF730748FcDc78A5AB854B898aC24b6d6001AbF7",
// 			AddressSimpleAccount:         "0xfaAe830bA56C40d17b7e23bfe092f23503464114",
// 			AddressMulticall:             "0x66e4f2437c5F612Ae25e94C1C549cb9f151E0cB3",
// 			AddressHyperlaneMailbox:      "0x2EaAd60F982f7B99b42f30e98B3b3f8ff89C0A46",
// 			AddressHyperlaneIgp:          "0x16e81e1973939bD166FDc61651F731e1658060F3",
// 			AddressPaymaster:             "0xdAE5e7CEBe4872BF0776477EcCCD2A0eFdF54f0e",
// 			AddressEscrow:                "0x9925D4a40ea432A25B91ab424b16c8FC6e0Eec5A",
// 			AddressEscrowFactory:         "0xC531388B2C2511FDFD16cD48f1087A747DC34b33",
// 		}
// 	case "0x4268", "17000":
// 		chainId = "200810"
// 		rpcURL = "https://ethereum-holesky-rpc.publicnode.com"
// 		addresses = Chain{
// 			AddressEntrypoint:            "0xc5Ff094002cdaF36d6a766799eB63Ec82B8C79F1",
// 			AddressEntrypointSimulations: "0x67B9841e9864D394FDc02e787A0Ac37f32B49eC7",
// 			AddressSimpleAccountFactory:  "0x39351b719D044CF6E91DEC75E78e5d128c582bE7",
// 			AddressSimpleAccount:         "0x0983a4e9D9aB03134945BFc9Ec9EF31338AB7465",
// 			AddressMulticall:             "0x98876409cc48507f8Ee8A0CCdd642469DBfB3E21",
// 			AddressHyperlaneMailbox:      "0x913A6477496eeb054C9773843a64c8621Fc46e8C",
// 			AddressHyperlaneIgp:          "0x2Fb9F9bd9034B6A5CAF3eCDB30db818619EbE9f1",
// 			AddressPaymaster:             "0xA5bcda4aA740C02093Ba57A750a8f424BC8B4B13",
// 			AddressEscrow:                "0x686130A96724734F0B6f99C6D32213BC62C1830A",
// 			AddressEscrowFactory:         "0x45d5D46B097870223fDDBcA9a9eDe35A7D37e2A1",
// 		}
// 	case "0xaa36a7", "11155111":
// 		chainId = "11155111"
// 		rpcURL = "https://rpc2.sepolia.org"
// 		addresses = Chain{
// 			AddressEntrypoint:            "0xA6eBc93dA2C99654e7D6BC12ed24362061805C82",
// 			AddressEntrypointSimulations: "0x0d17dE0436b65279c8D7A75847F84626687A1647",
// 			AddressSimpleAccountFactory:  "0x54bed3E354cbF23C2CADaB1dF43399473e38a358",
// 			AddressSimpleAccount:         "0x54bed3E354cbF23C2CADaB1dF43399473e38a358",
// 			AddressMulticall:             "0x6958206f218D8f889ECBb76B89eE9bF1CAe37715",
// 			AddressHyperlaneMailbox:      "0xAc165ff97Dc42d87D858ba8BC4AA27429a8C48e8",
// 			AddressHyperlaneIgp:          "0x00eb6D45afac57E708eC3FA6214BFe900aFDb95D",
// 			AddressPaymaster:             "0x31aCA626faBd9df61d24A537ecb9D646994b4d4d",
// 			AddressEscrow:                "0xea8D264dF67c9476cA80A24067c2F3CF7726aC4d",
// 			AddressEscrowFactory:         "0xd9842E241B7015ea1E1B5A90Ae20b6453ADF2723",
// 		}
// 	case "0xe34", "3636":
// 		chainId = "3636"
// 		rpcURL = "https://node.botanixlabs.dev"
// 		addresses = Chain{
// 			AddressEntrypoint:            "0xF7B12fFBC58dd654aeA52f1c863bf3f4731f848F",
// 			AddressEntrypointSimulations: "0x1db7F1263FbfBe5d91548B3422563179f6bE8d99",
// 			AddressSimpleAccountFactory:  "0xFB23dB8098Faf2dB307110905dC3698Fe27E136d",
// 			AddressSimpleAccount:         "0x15aA997cC02e103a7570a1C26F09996f6FBc1829",
// 			AddressMulticall:             "0x6cB50ee0241C7AE6Ebc30A34a9F3C23A96098bBf",
// 			AddressHyperlaneMailbox:      "0xd2DB8440B7dC1d05aC2366b353f1cF205Cf875EA",
// 			AddressHyperlaneIgp:          "0x8439DBdca66C9F72725f1B2d50dFCdc7c6CBBbEb",
// 			AddressPaymaster:             "0xbbfb649f42Baf44729a150464CBf6B89349A634a",
// 			AddressEscrow:                "0xCD77545cA802c4B05ff359f7b10355EC220E7476",
// 			AddressEscrowFactory:         "0xA6eBc93dA2C99654e7D6BC12ed24362061805C82",
// 		}
// 	default:
// 		return nil, nil, fmt.Errorf("unsupported chain ID: %s", chainId)
// 	}

// 	client, err = ethclient.Dial(rpcURL)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	domain, err := strconv.ParseUint(chainId, 0, 32)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	chain = &Chain{
// 		ChainId:                      chainId,
// 		Domain:                       uint32(domain),
// 		AddressEntrypoint:            addresses.AddressEntrypoint,
// 		AddressEntrypointSimulations: addresses.AddressEntrypointSimulations,
// 		AddressSimpleAccountFactory:  addresses.AddressSimpleAccountFactory,
// 		AddressMulticall:             addresses.AddressMulticall,
// 		AddressHyperlaneMailbox:      addresses.AddressHyperlaneMailbox,
// 		AddressHyperlaneIgp:          addresses.AddressHyperlaneIgp,
// 		AddressPaymaster:             addresses.AddressPaymaster,
// 		AddressEscrow:                addresses.AddressEscrow,
// 		AddressEscrowFactory:         addresses.AddressEscrowFactory,
// 	}

// 	return client, chain, nil
// }

// Helper function to convert []byte to hex string prefixed with "0x".
func ToHexBytes(data []byte) string {
	if len(data) == 0 {
		return "0x"
	}
	return "0x" + hex.EncodeToString(data)
}

func HexToBytes(hexStr string) ([]byte, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Helper function to convert common.Address to hex string prefixed with "0x".
func ToHexAddress(addr common.Address) string {
	return "0x" + hex.EncodeToString(addr[:])
}

// Helper function to convert [4]byte to a uint32 string.
func Uint32ToString(data [4]byte) string {
	value := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	return fmt.Sprintf("%d", value)
}

// Helper function to convert a byte to string.
func Uint8ToString(data byte) string {
	return fmt.Sprintf("%d", data)
}

func Bytes32PadRight(data []byte) [32]byte {
	var result [32]byte
	if len(data) >= 32 {
		copy(result[:], data[:32])
	} else {
		copy(result[:], data)
	}
	return result
}

func Bytes32PadLeft(data []byte) [32]byte {
	var result [32]byte
	if len(data) >= 32 {
		copy(result[:], data[:32])
	} else {
		start := 32 - len(data)
		copy(result[start:], data)
	}
	return result
}

var chainTypeMap = map[string]struct {
	ChainId string
	VM      string
}{
	"0x3106A":    {ChainId: "200810", VM: "evm"},
	"200810":     {ChainId: "200810", VM: "evm"},
	"0x4268":     {ChainId: "17000", VM: "evm"},
	"17000":      {ChainId: "17000", VM: "evm"},
	"0xAA36A7":   {ChainId: "11155111", VM: "evm"},
	"11155111":   {ChainId: "11155111", VM: "evm"},
	"0xF35A":     {ChainId: "62298", VM: "evm"},
	"62298":      {ChainId: "62298", VM: "evm"},
	"0x63639999": {ChainId: "1667471769", VM: "tvm"},
	"1667471769": {ChainId: "1667471769", VM: "tvm"},
	"998":        {ChainId: "998", VM: "evm"},
}

func GetChainType(chainId string) (string, string, error) {
	if data, found := chainTypeMap[chainId]; found {
		return data.ChainId, data.VM, nil
	}

	return "", "", fmt.Errorf("unsupporting chain id: %s", chainId)
}

func HexToUint64(hexStr string) (uint64, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("failed to decode hex: %w", err)
	}

	if len(bytes) > 8 {
		return 0, fmt.Errorf("hex value too large for uint64")
	}

	var value uint64
	for _, b := range bytes {
		value = (value << 8) | uint64(b)
	}
	return value, nil
}

func BytesToUint64(bytes []byte) (uint64, error) {
	if len(bytes) > 8 {
		return 0, fmt.Errorf("hex value too large for uint64")
	}

	var value uint64
	for _, b := range bytes {
		value = (value << 8) | uint64(b)
	}
	return value, nil
}
