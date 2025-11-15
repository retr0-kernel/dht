package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"dht/internal/auth"
	"dht/internal/models"
)

type Handler struct {
	userService   *models.UserService
	apiKeyService *models.APIKeyService
	authService   *auth.AuthService
}

func NewHandler(userService *models.UserService, apiKeyService *models.APIKeyService, authService *auth.AuthService) *Handler {
	return &Handler{
		userService:   userService,
		apiKeyService: apiKeyService,
		authService:   authService,
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
