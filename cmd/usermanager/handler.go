package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dht/internal/auth"
	"dht/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	userService   *models.UserService
	apiKeyService *models.APIKeyService
	authService   *auth.AuthService
	db            *pgxpool.Pool
}

func NewHandler(userService *models.UserService, apiKeyService *models.APIKeyService, authService *auth.AuthService, db *pgxpool.Pool) *Handler {
	return &Handler{
		userService:   userService,
		apiKeyService: apiKeyService,
		authService:   authService,
		db:            db,
	}
}

// Signup handles user registration
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email, username, and password are required")
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		respondError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Create user
	user, err := h.userService.CreateUser(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			respondError(w, http.StatusConflict, "Email or username already exists")
			return
		}
		log.Printf("Error creating user: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Return user response
	respondJSON(w, http.StatusCreated, models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	})
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Authenticate user
	user, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Extract IP and User-Agent
	ipAddress := getIPAddress(r)
	userAgent := r.UserAgent()

	// Create session
	_, err = h.userService.CreateSession(r.Context(), user.ID, ipAddress, userAgent)
	if err != nil {
		log.Printf("Error creating session: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Generate JWT tokens
	accessToken, err := h.authService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Error generating access token: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Printf("Error generating refresh token: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return login response
	respondJSON(w, http.StatusOK, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
		},
	})
}

// CreateAPIKey handles API key creation
func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "API key name is required")
		return
	}

	// Create API key
	apiKey, plainKey, err := h.apiKeyService.CreateAPIKey(r.Context(), userID, req.Name, req.Scopes, req.ExpiresInDays)
	if err != nil {
		log.Printf("Error creating API key: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	// Return API key response (including plain key - only time it's shown)
	respondJSON(w, http.StatusCreated, models.APIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		KeyPrefix: apiKey.KeyPrefix,
		Key:       plainKey, // Only returned on creation
		Scopes:    apiKey.Scopes,
		IsActive:  apiKey.IsActive,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
	})
}

// ListAPIKeys handles listing user's API keys
func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get API keys
	apiKeys, err := h.apiKeyService.ListAPIKeys(r.Context(), userID)
	if err != nil {
		log.Printf("Error listing API keys: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to list API keys")
		return
	}

	// Convert to response format
	var responses []models.APIKeyResponse
	for _, key := range apiKeys {
		responses = append(responses, models.APIKeyResponse{
			ID:         key.ID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scopes:     key.Scopes,
			IsActive:   key.IsActive,
			LastUsedAt: key.LastUsedAt,
			ExpiresAt:  key.ExpiresAt,
			CreatedAt:  key.CreatedAt,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"api_keys": responses,
		"count":    len(responses),
	})
}

// ValidateAPIKey validates an API key and returns user ID
func (h *Handler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.APIKey == "" {
		respondError(w, http.StatusBadRequest, "API key is required")
		return
	}

	// Verify API key
	userID, err := h.apiKeyService.VerifyAPIKey(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid API key")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":   true,
		"user_id": userID,
	})
}

// Health check endpoint
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "usermanager",
	})
}

// Helper function to extract user ID from JWT token
func (h *Handler) extractUserIDFromToken(r *http.Request) (int64, error) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, http.ErrNoCookie
	}

	// Extract token (format: "Bearer <token>")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, http.ErrNoCookie
	}

	token := parts[1]

	// Validate and extract claims
	claims, err := h.authService.ValidateAccessToken(token)
	if err != nil {
		return 0, err
	}

	// Parse user ID
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// ListUsageRecords lists usage records for authenticated user
func (h *Handler) ListUsageRecords(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `
		SELECT id, user_id, api_key_id, operation, key_accessed, 
		       request_size_bytes, response_size_bytes, status_code, 
		       duration_ms, ip_address, user_agent, error_message, created_at
		FROM usage_records
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := h.db.Query(context.Background(), query, userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch usage records")
		return
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var record struct {
			ID                int64
			UserID            int64
			APIKeyID          *int64
			Operation         string
			KeyAccessed       *string
			RequestSizeBytes  int64
			ResponseSizeBytes int64
			StatusCode        int
			DurationMs        int
			IPAddress         *string
			UserAgent         *string
			ErrorMessage      *string
			CreatedAt         time.Time
		}

		err := rows.Scan(
			&record.ID, &record.UserID, &record.APIKeyID, &record.Operation,
			&record.KeyAccessed, &record.RequestSizeBytes, &record.ResponseSizeBytes,
			&record.StatusCode, &record.DurationMs, &record.IPAddress,
			&record.UserAgent, &record.ErrorMessage, &record.CreatedAt,
		)
		if err != nil {
			continue
		}

		records = append(records, map[string]interface{}{
			"id":                  record.ID,
			"user_id":             record.UserID,
			"api_key_id":          record.APIKeyID,
			"operation":           record.Operation,
			"key_accessed":        record.KeyAccessed,
			"request_size_bytes":  record.RequestSizeBytes,
			"response_size_bytes": record.ResponseSizeBytes,
			"status_code":         record.StatusCode,
			"duration_ms":         record.DurationMs,
			"ip_address":          record.IPAddress,
			"user_agent":          record.UserAgent,
			"error_message":       record.ErrorMessage,
			"created_at":          record.CreatedAt,
		})
	}

	respondJSON(w, http.StatusOK, records)
}

// GetUsageStats returns usage statistics for authenticated user
func (h *Handler) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(*) FILTER (WHERE status_code >= 200 AND status_code < 300) as successful_requests,
			COUNT(*) FILTER (WHERE status_code >= 400) as failed_requests,
			COALESCE(SUM(request_size_bytes + response_size_bytes), 0) as total_bytes_transferred,
			COALESCE(AVG(duration_ms), 0) as average_latency_ms
		FROM usage_records
		WHERE user_id = $1
	`

	var stats struct {
		TotalRequests         int64
		SuccessfulRequests    int64
		FailedRequests        int64
		TotalBytesTransferred int64
		AverageLatencyMs      float64
	}

	err = h.db.QueryRow(context.Background(), query, userID).Scan(
		&stats.TotalRequests,
		&stats.SuccessfulRequests,
		&stats.FailedRequests,
		&stats.TotalBytesTransferred,
		&stats.AverageLatencyMs,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch statistics")
		return
	}

	// Get requests by operation
	operationQuery := `
		SELECT operation, COUNT(*) as count
		FROM usage_records
		WHERE user_id = $1
		GROUP BY operation
	`

	rows, err := h.db.Query(context.Background(), operationQuery, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch operation stats")
		return
	}
	defer rows.Close()

	requestsByOperation := make(map[string]int64)
	for rows.Next() {
		var operation string
		var count int64
		if err := rows.Scan(&operation, &count); err == nil {
			requestsByOperation[operation] = count
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_requests":          stats.TotalRequests,
		"successful_requests":     stats.SuccessfulRequests,
		"failed_requests":         stats.FailedRequests,
		"total_bytes_transferred": stats.TotalBytesTransferred,
		"average_latency_ms":      stats.AverageLatencyMs,
		"requests_by_operation":   requestsByOperation,
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func isValidEmail(email string) bool {
	// Simple email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(forwarded, ",")
		ip := strings.TrimSpace(ips[0])
		return cleanIPAddress(ip)
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return cleanIPAddress(realIP)
	}

	// Fall back to RemoteAddr (strip port)
	return cleanIPAddress(r.RemoteAddr)
}

// cleanIPAddress removes port from IP address
func cleanIPAddress(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// No port present, return as-is
		return addr
	}
	return host
}
