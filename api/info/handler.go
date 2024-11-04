package infoHandler

import (
	"encoding/json"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Query().Get("query") {
	case "version":
		Version(w)
		return
	default:
		version := "Hello, World!"
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
