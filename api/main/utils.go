// utils.go
package handler

import (
	"encoding/json"
	"net/http"
)

// WriteJSONResponse writes a message to the JSON response
func WriteJSONResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": message,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
