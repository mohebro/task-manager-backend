package model

import (
	"time"

	"github.com/google/uuid"
)

// User is a team member who can sign in to Task Manager.
// PasswordHash is never returned in JSON responses; handlers
// return PublicUser instead.
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PublicUser is the safe, outward-facing shape of a User.
type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Public returns the safe shape of this user.
func (u *User) Public() PublicUser {
	return PublicUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// RegisterInput is the body of POST /auth/register.
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput is the body of POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned by register and login:
// the user plus a JWT the client sends as
// Authorization: Bearer <token>.
type AuthResponse struct {
	User  PublicUser `json:"user"`
	Token string     `json:"token"`
}
