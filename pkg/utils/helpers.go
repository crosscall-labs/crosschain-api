package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
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

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		queryTag := fieldType.Tag.Get("query")
		optionalTag := fieldType.Tag.Get("optional")

		if queryTag != "" {
			queryValue := r.URL.Query().Get(queryTag)

			// If the field is required (i.e., optional is not set to "true")
			if queryValue == "" && optionalTag != "true" {
				missingFields = append(missingFields, queryTag)
			} else if queryValue != "" {
				field.SetString(queryValue)
			}
		}
	}

	if len(missingFields) > 0 {
		ErrMalformedRequest(w, fmt.Sprintf("Missing parameters: %v", missingFields))
		return false
	}

	return true
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
