package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
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

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Email and password are required"})
		return
	}

	// Get user from database
	var userID, passwordHash string
	err := db.QueryRowContext(
		context.Background(),
		`SELECT id, password_hash FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &passwordHash)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid email or password"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   req.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token:   tokenString,
		UserID:  userID,
		Email:   req.Email,
		Message: "Login successful",
	})
}
