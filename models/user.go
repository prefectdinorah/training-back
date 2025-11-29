package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Name      *string   `json:"name"`
	Surname   *string   `json:"surname"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone"`
	Trainer   *uuid.UUID `json:"trainer"`
	IsTrainer bool      `json:"is_trainer"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	Surname  string `json:"surname,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Name     *string    `json:"name"`
	Surname  *string    `json:"surname"`
	Email    *string    `json:"email"`
	Phone    *string    `json:"phone"`
	Trainer  *uuid.UUID `json:"trainer"`
	IsTrainer *bool    `json:"is_trainer"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
