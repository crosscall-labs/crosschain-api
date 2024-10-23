package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
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

	// Iterate through the fields in the struct
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		queryTag := fieldType.Tag.Get("query")
		optionalTag := fieldType.Tag.Get("optional")

		if field.Kind() == reflect.Struct {
			// Recursively parse nested struct fields
			nestedParams := reflect.New(fieldType.Type).Interface()
			if !ParseAndValidateParams(w, r, nestedParams) {
				return false // If the nested struct has missing fields, return false
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

	// If there are missing fields, return an error response
	if len(missingFields) > 0 {
		http.Error(w, "Missing fields: "+strings.Join(missingFields, ", "), http.StatusBadRequest)
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
