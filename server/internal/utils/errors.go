package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, clientMsg string, err error) {
	log.Printf("request %s %s -> %d: %v", r.Method, r.URL.Path, status, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: clientMsg})
}
