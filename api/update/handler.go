package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type UpdateRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
	TrainerID *string `json:"trainer_id"`
	IsTrainer *bool   `json:"is_trainer"`
}

type UpdateResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var db *sql.DB
var jwtSecret string

func init() {
	var err error
	databaseURL := os.Getenv("DATABASE_URL")
	jwtSecret = os.Getenv("JWT_SECRET")

	if databaseURL == "" || jwtSecret == "" {
		panic("DATABASE_URL and JWT_SECRET must be set")
	}

	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}
}

func parseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}

	return userID, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Extract token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing authorization header"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization header format"})
		return
	}

	userID, err := parseToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid token"})
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Build update query
	updateQuery := "UPDATE users SET"
	params := []interface{}{}
	paramCount := 1

	if req.FirstName != nil {
		updateQuery += fmt.Sprintf(" first_name = $%d,", paramCount)
		params = append(params, *req.FirstName)
		paramCount++
	}
	if req.LastName != nil {
		updateQuery += fmt.Sprintf(" last_name = $%d,", paramCount)
		params = append(params, *req.LastName)
		paramCount++
	}
	if req.Phone != nil {
		updateQuery += fmt.Sprintf(" phone = $%d,", paramCount)
		params = append(params, *req.Phone)
		paramCount++
	}
	if req.TrainerID != nil {
		updateQuery += fmt.Sprintf(" trainer_id = $%d,", paramCount)
		params = append(params, *req.TrainerID)
		paramCount++
	}
	if req.IsTrainer != nil {
		updateQuery += fmt.Sprintf(" is_trainer = $%d,", paramCount)
		params = append(params, *req.IsTrainer)
		paramCount++
	}

	// Remove trailing comma and add WHERE clause
	updateQuery = strings.TrimSuffix(updateQuery, ",")
	updateQuery += fmt.Sprintf(" WHERE id = $%d", paramCount)
	params = append(params, userID)

	_, err = db.ExecContext(context.Background(), updateQuery, params...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update profile"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UpdateResponse{Message: "Profile updated successfully"})
}
