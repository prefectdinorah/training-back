package handler

import (
	"encoding/json"
	"net/http"
)

// Response structure for JSON responses
type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create response
	response := Response{
		Message: "Hello from Go backend on Vercel! 🚀",
		Status:  "success",
	}

	// Send JSON response
	json.NewEncoder(w).Encode(response)
}
