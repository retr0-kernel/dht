package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "strconv"
	"strings"
	"time"

	"dht/internal/config"
)

// AuthMiddleware validates API keys against the usermanager service
func AuthMiddleware(cfg *config.Config, rls *RateLimiterStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Get API key from X-API-Key header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				respondError(w, http.StatusUnauthorized, "Missing X-API-Key header")
				return
			}

			// Validate API key with usermanager service
			userID, err := validateAPIKey(cfg, apiKey)
			if err != nil {
				log.Printf("API key validation failed: %v\n", err)
				respondError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// Check rate limit for this user
			if !rls.AllowRequest(userID) {
				respondError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateAPIKey validates an API key against the usermanager service
func validateAPIKey(cfg *config.Config, apiKey string) (int64, error) {
	// Create request to usermanager
	url := fmt.Sprintf("http://localhost:%s/validate-key", cfg.UserManagerPort)

	reqBody := map[string]string{"api_key": apiKey}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API key validation failed with status %d", resp.StatusCode)
	}

	var result struct {
		UserID int64 `json:"user_id"`
		Valid  bool  `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if !result.Valid {
		return 0, fmt.Errorf("invalid API key")
	}

	return result.UserID, nil
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Consistency")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
