package models

import "time"

// Request models
type SignupRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAPIKeyRequest struct {
	Name          string   `json:"name"`
	Scopes        []string `json:"scopes,omitempty"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
}

// Response models
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type APIKeyResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Key        string     `json:"key,omitempty"` // Only returned on creation
	Scopes     []string   `json:"scopes"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
