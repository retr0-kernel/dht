package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dht/internal/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	IsActive     bool       `json:"is_active"`
	IsVerified   bool       `json:"is_verified"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type Session struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	SessionToken string     `json:"session_token"`
	RefreshToken *string    `json:"refresh_token,omitempty"`
	IPAddress    *string    `json:"ip_address,omitempty"`
	UserAgent    *string    `json:"user_agent,omitempty"`
	IsActive     bool       `json:"is_active"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

type UserService struct {
	db          *pgxpool.Pool
	authService *auth.AuthService
}

func NewUserService(db *pgxpool.Pool, authService *auth.AuthService) *UserService {
	return &UserService{
		db:          db,
		authService: authService,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, email, username, password string) (*User, error) {
	// Hash password
	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	query := `
		INSERT INTO users (email, username, password_hash, is_active, is_verified)
		VALUES ($1, $2, $3, true, false)
		RETURNING id, email, username, is_active, is_verified, created_at, updated_at
	`

	var user User
	err = s.db.QueryRow(ctx, query, email, username, hashedPassword).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// AuthenticateUser authenticates a user by email and password
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, is_active, is_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Verify password
	if err := s.authService.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	updateQuery := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err = s.db.Exec(ctx, updateQuery, user.ID)
	if err != nil {
		// Log error but don't fail authentication
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, email, username, is_active, is_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user User
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// CreateSession creates a new session for a user
func (s *UserService) CreateSession(ctx context.Context, userID int64, ipAddress, userAgent string) (*Session, error) {
	// Generate session token
	sessionToken, err := s.authService.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.authService.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Insert session
	query := `
		INSERT INTO sessions (user_id, session_token, refresh_token, ip_address, user_agent, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW() + INTERVAL '7 days')
		RETURNING id, user_id, session_token, refresh_token, ip_address, user_agent, is_active, expires_at, created_at, updated_at
	`

	var session Session
	err = s.db.QueryRow(ctx, query, userID, sessionToken, refreshToken, ipAddress, userAgent).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionToken,
		&session.RefreshToken,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}
