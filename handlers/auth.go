package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"healthy-lifestyle-api/models"
)

var (
	db        *sql.DB
	jwtSecret string
)

func Init(database *sql.DB, secret string) {
	db = database
	jwtSecret = secret
}

// RegisterHandler creates a new user
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Login == "" || req.Email == "" || req.Password == "" {
		jsonError(w, "Login, email, and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert user
	var name, surname, phone *string
	if req.Name != "" {
		name = &req.Name
	}
	if req.Surname != "" {
		surname = &req.Surname
	}
	if req.Phone != "" {
		phone = &req.Phone
	}

	userID := uuid.New()
	_, err = db.Exec(
		"INSERT INTO users (id, login, name, surname, email, phone, password_hash) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		userID, req.Login, name, surname, req.Email, phone, string(hashedPassword),
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			jsonError(w, "Login or email already exists", http.StatusConflict)
		} else {
			log.Printf("Error creating user: %v\n", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Generate JWT token
	user := models.User{
		ID:    userID,
		Login: req.Login,
		Email: req.Email,
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		jsonError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// LoginHandler authenticates a user
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		jsonError(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Get user from database
	var user models.User
	var passwordHash string

	err := db.QueryRow(
		"SELECT id, login, name, surname, email, phone, trainer, is_trainer, created_at, updated_at FROM users WHERE login = $1",
		req.Login,
	).Scan(&user.ID, &user.Login, &user.Name, &user.Surname, &user.Email, &user.Phone, &user.Trainer, &user.IsTrainer, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, "Invalid login or password", http.StatusUnauthorized)
		} else {
			log.Printf("Error querying user: %v\n", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Get password hash
	err = db.QueryRow("SELECT password_hash FROM users WHERE id = $1", user.ID).Scan(&passwordHash)
	if err != nil {
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		jsonError(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID)
	if err != nil {
		jsonError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfileHandler returns current user profile
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(uuid.UUID)

	var user models.User
	err := db.QueryRow(
		"SELECT id, login, name, surname, email, phone, trainer, is_trainer, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Login, &user.Name, &user.Surname, &user.Email, &user.Phone, &user.Trainer, &user.IsTrainer, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateProfileHandler updates user profile
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(uuid.UUID)

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updateFields = append(updateFields, "name = $"+string(rune(argIndex)))
		args = append(args, req.Name)
		argIndex++
	}
	if req.Surname != nil {
		updateFields = append(updateFields, "surname = $"+string(rune(argIndex)))
		args = append(args, req.Surname)
		argIndex++
	}
	if req.Email != nil {
		updateFields = append(updateFields, "email = $"+string(rune(argIndex)))
		args = append(args, req.Email)
		argIndex++
	}
	if req.Phone != nil {
		updateFields = append(updateFields, "phone = $"+string(rune(argIndex)))
		args = append(args, req.Phone)
		argIndex++
	}
	if req.Trainer != nil {
		updateFields = append(updateFields, "trainer = $"+string(rune(argIndex)))
		args = append(args, req.Trainer)
		argIndex++
	}
	if req.IsTrainer != nil {
		updateFields = append(updateFields, "is_trainer = $"+string(rune(argIndex)))
		args = append(args, req.IsTrainer)
		argIndex++
	}

	if len(updateFields) == 0 {
		jsonError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	updateFields = append(updateFields, "updated_at = $"+string(rune(argIndex)))
	args = append(args, time.Now())
	args = append(args, userID)

	query := "UPDATE users SET " + strings.Join(updateFields, ", ") + " WHERE id = $" + string(rune(len(args)))
	_, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating user: %v\n", err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return updated user
	var user models.User
	err = db.QueryRow(
		"SELECT id, login, name, surname, email, phone, trainer, is_trainer, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Login, &user.Name, &user.Surname, &user.Email, &user.Phone, &user.Trainer, &user.IsTrainer, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		jsonError(w, "Failed to retrieve updated user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteProfileHandler deletes user account
func DeleteProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(uuid.UUID)

	result, err := db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		log.Printf("Error deleting user: %v\n", err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions
func generateJWT(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID.String(),
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func validateJWT(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	userIDStr := claims["userID"].(string)
	return uuid.Parse(userIDStr)
}

func jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: message})
}
