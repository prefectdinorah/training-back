package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}

type RegisterResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Message   string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var db *sql.DB

func init() {
	var err error
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL not set")
	}

	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}
}

func validateEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle CORS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username, email, and password are required"})
		return
	}

	if !validateEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid email format"})
		return
	}

	if len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Password must be at least 6 characters"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to process password"})
		return
	}

	// Insert user
	var userID string
	err = db.QueryRowContext(
		context.Background(),
		`INSERT INTO users (username, email, password_hash, first_name, last_name, phone)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Phone,
	).Scan(&userID)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Username already taken"})
			return
		}
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Email already registered"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to register user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Message:  "User registered successfully",
	})
}
