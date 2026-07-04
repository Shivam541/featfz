package handlers

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(healthResponse{
		Success: true,
		Status:  "ok",
	})
}
