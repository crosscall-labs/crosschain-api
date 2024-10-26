package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
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

func ParseAndValidateParams(w http.ResponseWriter, r *http.Request, params interface{}) bool {
	val := reflect.ValueOf(params).Elem() // Dereference the pointer to access the underlying struct
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
				if !ParseAndValidateParams(w, r, nestedParams) {
					return false
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
		http.Error(w, "Missing fields: "+strings.Join(missingFields, ", "), http.StatusBadRequest)
		return false
	}

	return true
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
		} else {
			fmt.Printf("\n%s: %v", fieldName, field.Interface()) // Print field value
		}
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("Error (Code: %d, Message: %s)", e.Code, e.Message)
}

func ErrMalformedRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	json.NewEncoder(w).Encode(&Error{
		Code:    400,
		Message: "Malformed request",
		Details: message,
	})
}

func ErrInternal(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(&Error{
		Code:    500,
		Message: "Internal server error",
		Details: message,
	})
}

func EnvKey2Ecdsa() (*ecdsa.PrivateKey, common.Address, error) {
	return PrivateKey2Sepc256k1(os.Getenv("RELAY_PRIVATE_KEY"))
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
		fmt.Printf("Method: %s, URL: %s", r.Method, r.URL)

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
		return "1667471769", "evm", "tonTvmTestnet", []int{2}, []int{0, 1, 2}, ""
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

// Helper function to convert []byte to hex string prefixed with "0x".
func ToHexBytes(data []byte) string {
	if len(data) == 0 {
		return "0x"
	}
	return "0x" + hex.EncodeToString(data)
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
