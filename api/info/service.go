package info

import (
	"encoding/json"
	"net/http"

	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func Version(w http.ResponseWriter) {
	versionResponse := utils.VersionResponse{Version: "Lamina Crosschain API v0.0.5b"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(versionResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
