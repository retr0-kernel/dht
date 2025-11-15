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
	"dht/internal/models"
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

	// Get TTL from query parameter
	ttl := time.Duration(0)
	if ttlStr := r.URL.Query().Get("ttl"); ttlStr != "" {
		ttlDuration, err := time.ParseDuration(ttlStr)
		if err == nil {
			ttl = ttlDuration
		}
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(int64)

	// Use hash ring to determine primary and replica nodes
	nodes := h.ring.LocateKey(key, 3) // Get 3 nodes (1 primary + 2 replicas)
	if len(nodes) == 0 {
		respondError(w, http.StatusServiceUnavailable, "No nodes available")
		return
	}

	primaryNode := nodes[0]
	replicaNodes := nodes[1:]

	log.Printf("PUT key=%s primary=%s replicas=%v (user=%d, consistency=%s)\n",
		key, primaryNode, replicaNodes, userID, consistency)

	// Write to primary node first
	reqURL := fmt.Sprintf("%s/store/%s", primaryNode, key)
	if ttl > 0 {
		reqURL = fmt.Sprintf("%s?ttl=%s", reqURL, ttl.String())
	}

	req, err := http.NewRequestWithContext(r.Context(), "PUT", reqURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	// Send request to primary DHT node
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to primary node: %v\n", err)
		respondError(w, http.StatusServiceUnavailable, "Primary node unavailable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		w.Write(responseBody)
		return
	}

	// Trigger replication if there are replica nodes
	if len(replicaNodes) > 0 {
		replReq := models.ReplicationRequest{
			Key:          key,
			Value:        body,
			Operation:    "SET",
			TTL:          ttl,
			Consistency:  consistency,
			PrimaryNode:  primaryNode,
			ReplicaNodes: replicaNodes,
			UserID:       userID,
		}

		h.triggerReplication(&replReq, consistency)
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"key":          key,
		"primary_node": primaryNode,
		"replicas":     len(replicaNodes),
	})
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

	// Use hash ring to determine primary and replica nodes
	nodes := h.ring.LocateKey(key, 3)
	if len(nodes) == 0 {
		respondError(w, http.StatusServiceUnavailable, "No nodes available")
		return
	}

	primaryNode := nodes[0]
	replicaNodes := nodes[1:]

	log.Printf("DELETE key=%s primary=%s replicas=%v (user=%d, consistency=%s)\n",
		key, primaryNode, replicaNodes, userID, consistency)

	// Delete from primary node
	reqURL := fmt.Sprintf("%s/store/%s", primaryNode, key)
	req, err := http.NewRequestWithContext(r.Context(), "DELETE", reqURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	// Send request to primary DHT node
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to primary node: %v\n", err)
		respondError(w, http.StatusServiceUnavailable, "Primary node unavailable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		responseBody, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		w.Write(responseBody)
		return
	}

	// Trigger replication if there are replica nodes
	if len(replicaNodes) > 0 {
		replReq := models.ReplicationRequest{
			Key:          key,
			Operation:    "DELETE",
			Consistency:  consistency,
			PrimaryNode:  primaryNode,
			ReplicaNodes: replicaNodes,
			UserID:       userID,
		}

		h.triggerReplication(&replReq, consistency)
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"key":          key,
		"primary_node": primaryNode,
		"replicas":     len(replicaNodes),
	})
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

// triggerReplication sends replication request to replicator service
func (h *Handler) triggerReplication(replReq *models.ReplicationRequest, consistency string) {
	replicatorURL := fmt.Sprintf("http://localhost:%s/replicate", h.config.ReplicatorPort)

	jsonData, err := json.Marshal(replReq)
	if err != nil {
		log.Printf("Failed to marshal replication request: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", replicatorURL, bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Failed to create replication request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// For eventual consistency, fire and forget
	if consistency == "eventual" {
		go func() {
			resp, err := h.httpClient.Do(req)
			if err != nil {
				log.Printf("Failed to trigger replication: %v\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				log.Printf("Replication request failed with status %d\n", resp.StatusCode)
			}
		}()
	} else {
		// For strong consistency, wait for replication
		resp, err := h.httpClient.Do(req)
		if err != nil {
			log.Printf("Failed to trigger replication: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Strong replication failed with status %d\n", resp.StatusCode)
		}
	}
}
