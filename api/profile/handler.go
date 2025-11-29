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

type ProfileResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
	TrainerID *string `json:"trainer_id"`
	IsTrainer bool   `json:"is_trainer"`
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
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

	// Get user profile
	var profile ProfileResponse
	err = db.QueryRowContext(
		context.Background(),
		`SELECT id, username, email, first_name, last_name, phone, trainer_id, is_trainer
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&profile.ID, &profile.Username, &profile.Email, &profile.FirstName, &profile.LastName, &profile.Phone, &profile.TrainerID, &profile.IsTrainer)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}
