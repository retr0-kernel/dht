package models

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type APIKey struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	KeyHash    string     `json:"-"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type APIKeyService struct {
	db *pgxpool.Pool
}

func NewAPIKeyService(db *pgxpool.Pool) *APIKeyService {
	return &APIKeyService{db: db}
}

// GenerateAPIKey generates a random API key
func (s *APIKeyService) generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashAPIKey hashes an API key using bcrypt
func (s *APIKeyService) hashKey(apiKey string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CreateAPIKey creates a new API key for a user
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID int64, name string, scopes []string, expiresInDays int) (*APIKey, string, error) {
	// Generate API key
	plainKey, err := s.generateKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the key
	keyHash, err := s.hashKey(plainKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}

	// Create key prefix (first 8 characters)
	keyPrefix := plainKey[:8]

	// Format the full key with prefix
	fullKey := fmt.Sprintf("ydht_%s", plainKey)

	// Calculate expiration
	var expiresAt *time.Time
	if expiresInDays > 0 {
		expiry := time.Now().AddDate(0, 0, expiresInDays)
		expiresAt = &expiry
	}

	// Default scopes if none provided
	if len(scopes) == 0 {
		scopes = []string{"read", "write"}
	}

	// Insert API key
	query := `
		INSERT INTO api_keys (user_id, key_hash, key_prefix, name, scopes, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING id, user_id, key_prefix, name, scopes, is_active, expires_at, created_at, updated_at
	`

	var apiKey APIKey
	err = s.db.QueryRow(ctx, query, userID, keyHash, keyPrefix, name, scopes, expiresAt).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyPrefix,
		&apiKey.Name,
		&apiKey.Scopes,
		&apiKey.IsActive,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	return &apiKey, fullKey, nil
}

// ListAPIKeys lists all API keys for a user
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID int64) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, key_prefix, name, scopes, is_active, last_used_at, expires_at, created_at, updated_at, revoked_at
		FROM api_keys
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.KeyPrefix,
			&apiKey.Name,
			&apiKey.Scopes,
			&apiKey.IsActive,
			&apiKey.LastUsedAt,
			&apiKey.ExpiresAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
			&apiKey.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, nil
}

// VerifyAPIKey verifies an API key and returns the associated user ID
func (s *APIKeyService) VerifyAPIKey(ctx context.Context, plainKey string) (int64, error) {
	// Strip the "ydht_" prefix if present
	if len(plainKey) > 5 && plainKey[:5] == "ydht_" {
		plainKey = plainKey[5:]
	}

	// Get key prefix
	keyPrefix := plainKey[:8]

	// Find all keys with this prefix
	query := `
		SELECT id, user_id, key_hash, is_active, expires_at
		FROM api_keys
		WHERE key_prefix = $1 AND is_active = true AND revoked_at IS NULL
	`

	rows, err := s.db.Query(ctx, query, keyPrefix)
	if err != nil {
		return 0, fmt.Errorf("failed to find API key: %w", err)
	}
	defer rows.Close()

	// Try to match the hash
	for rows.Next() {
		var id, userID int64
		var keyHash string
		var isActive bool
		var expiresAt *time.Time

		err := rows.Scan(&id, &userID, &keyHash, &isActive, &expiresAt)
		if err != nil {
			continue
		}

		// Check expiration
		if expiresAt != nil && expiresAt.Before(time.Now()) {
			continue
		}

		// Verify hash
		if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(plainKey)); err == nil {
			// Update last used
			updateQuery := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
			s.db.Exec(ctx, updateQuery, id)

			return userID, nil
		}
	}

	return 0, fmt.Errorf("invalid API key")
}
