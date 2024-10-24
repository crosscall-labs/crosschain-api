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
