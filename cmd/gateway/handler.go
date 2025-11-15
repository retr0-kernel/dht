package main

import (
	"bytes"
	_ "context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"dht/internal/config"
	"dht/internal/hashring"
)

type Handler struct {
	config           *config.Config
	ring             *hashring.HashRing
	rateLimiterStore *RateLimiterStore
	httpClient       *http.Client
}

func NewHandler(cfg *config.Config, ring *hashring.HashRing, rls *RateLimiterStore) *Handler {
	return &Handler{
		config:           cfg,
		ring:             ring,
		rateLimiterStore: rls,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PutKey handles PUT /v1/kv/:key
func (h *Handler) PutKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Read request body (the value to store)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Get consistency level from header (default: eventual)
	consistency := r.Header.Get("X-Consistency")
	if consistency == "" {
		consistency = "eventual"
	}

	// Validate consistency level
	if consistency != "strong" && consistency != "eventual" {
		respondError(w, http.StatusBadRequest, "Invalid consistency level. Must be 'strong' or 'eventual'")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(int64)

	// Use hash ring to determine which node should handle this key
	nodeURL := h.ring.GetNode(key)
	log.Printf("PUT key=%s routed to node=%s (user=%d, consistency=%s)\n", key, nodeURL, userID, consistency)

	// Forward request to DHT node
	reqURL := fmt.Sprintf("%s/store/%s", nodeURL, key)
	req, err := http.NewRequestWithContext(r.Context(), "PUT", reqURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	// Forward headers
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.Header.Set("X-Consistency", consistency)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	// Send request to DHT node
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to DHT node: %v\n", err)
		respondError(w, http.StatusServiceUnavailable, "DHT node unavailable")
		return
	}
	defer resp.Body.Close()

	// Read response from DHT node
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading DHT node response: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to read response")
		return
	}

	// Forward DHT node response to client
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)
}

// GetKey handles GET /v1/kv/:key
func (h *Handler) GetKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Get consistency level from header (default: eventual)
	consistency := r.Header.Get("X-Consistency")
	if consistency == "" {
		consistency = "eventual"
	}

	// Validate consistency level
	if consistency != "strong" && consistency != "eventual" {
		respondError(w, http.StatusBadRequest, "Invalid consistency level. Must be 'strong' or 'eventual'")
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id").(int64)

	// Use hash ring to determine which node should handle this key
	nodeURL := h.ring.GetNode(key)
	log.Printf("GET key=%s routed to node=%s (user=%d, consistency=%s)\n", key, nodeURL, userID, consistency)

	// Forward request to DHT node
	reqURL := fmt.Sprintf("%s/store/%s", nodeURL, key)
	req, err := http.NewRequestWithContext(r.Context(), "GET", reqURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	// Forward headers
	req.Header.Set("X-Consistency", consistency)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	// Send request to DHT node
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to DHT node: %v\n", err)
		respondError(w, http.StatusServiceUnavailable, "DHT node unavailable")
		return
	}
	defer resp.Body.Close()

	// Read response from DHT node
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading DHT node response: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to read response")
		return
	}

	// Forward DHT node response to client
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)
}

// DeleteKey handles DELETE /v1/kv/:key
func (h *Handler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Get consistency level from header (default: eventual)
	consistency := r.Header.Get("X-Consistency")
	if consistency == "" {
		consistency = "eventual"
	}

	// Validate consistency level
	if consistency != "strong" && consistency != "eventual" {
		respondError(w, http.StatusBadRequest, "Invalid consistency level. Must be 'strong' or 'eventual'")
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id").(int64)

	// Use hash ring to determine which node should handle this key
	nodeURL := h.ring.GetNode(key)
	log.Printf("DELETE key=%s routed to node=%s (user=%d, consistency=%s)\n", key, nodeURL, userID, consistency)

	// Forward request to DHT node
	reqURL := fmt.Sprintf("%s/store/%s", nodeURL, key)
	req, err := http.NewRequestWithContext(r.Context(), "DELETE", reqURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	// Forward headers
	req.Header.Set("X-Consistency", consistency)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	// Send request to DHT node
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to DHT node: %v\n", err)
		respondError(w, http.StatusServiceUnavailable, "DHT node unavailable")
		return
	}
	defer resp.Body.Close()

	// Read response from DHT node
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading DHT node response: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to read response")
		return
	}

	// Forward DHT node response to client
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)
}

// Health check endpoint
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "gateway",
		"nodes":   h.ring.GetAllNodes(),
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
