package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dht/internal/storage"
)

type DHTNode struct {
	storage *storage.Storage
	wal     *storage.WAL
	port    string
	nodeID  string
}

func main() {
	// Get configuration from environment
	port := os.Getenv("DHTNODE_PORT")
	if port == "" {
		port = "8082"
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node-1"
	}

	// Initialize storage
	store := storage.NewStorage()

	// Initialize WAL
	walPath := fmt.Sprintf("data/%s-wal.log", nodeID)
	os.MkdirAll("data", 0755)

	wal, err := storage.NewWAL(walPath)
	if err != nil {
		log.Fatalf("Failed to initialize WAL: %v\n", err)
	}
	defer wal.Close()

	// Restore from WAL
	if err := wal.Restore(store); err != nil {
		log.Printf("Warning: Failed to restore from WAL: %v\n", err)
	}

	node := &DHTNode{
		storage: store,
		wal:     wal,
		port:    port,
		nodeID:  nodeID,
	}

	// Setup HTTP server (we'll use HTTP instead of gRPC for simplicity)
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /store/{key}", node.handlePut)
	mux.HandleFunc("GET /store/{key}", node.handleGet)
	mux.HandleFunc("DELETE /store/{key}", node.handleDelete)
	mux.HandleFunc("GET /metrics", node.handleMetrics)
	mux.HandleFunc("GET /health", node.handleHealth)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      LoggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("DHT Node %s starting on port %s\n", nodeID, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited gracefully")
}

// handlePut handles PUT requests
func (n *DHTNode) handlePut(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Read value from body
	value, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read body")
		return
	}
	defer r.Body.Close()

	// Get TTL from query parameter (optional)
	ttl := time.Duration(0)
	if ttlStr := r.URL.Query().Get("ttl"); ttlStr != "" {
		ttlDuration, err := time.ParseDuration(ttlStr)
		if err == nil {
			ttl = ttlDuration
		}
	}

	// Write to WAL first (write-ahead logging)
	if err := n.wal.Append("SET", key, value, ttl); err != nil {
		log.Printf("WAL append failed: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to write to WAL")
		return
	}

	// Then write to storage
	if err := n.storage.Set(key, value, ttl); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to store value")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"key":     key,
		"node":    n.nodeID,
	})
}

// handleGet handles GET requests
func (n *DHTNode) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	value, err := n.storage.Get(key)
	if err != nil {
		respondError(w, http.StatusNotFound, "Key not found")
		return
	}

	// Return the raw value with appropriate content type
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Node-ID", n.nodeID)
	w.WriteHeader(http.StatusOK)
	w.Write(value)
}

// handleDelete handles DELETE requests
func (n *DHTNode) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Write to WAL first
	if err := n.wal.Append("DELETE", key, nil, 0); err != nil {
		log.Printf("WAL append failed: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to write to WAL")
		return
	}

	// Then delete from storage
	if err := n.storage.Delete(key); err != nil {
		respondError(w, http.StatusNotFound, "Key not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"key":     key,
		"node":    n.nodeID,
	})
}

// handleMetrics returns node metrics
func (n *DHTNode) handleMetrics(w http.ResponseWriter, r *http.Request) {
	walSize, _ := n.wal.Size()

	metrics := map[string]interface{}{
		"node_id":   n.nodeID,
		"key_count": n.storage.KeyCount(),
		"wal_size":  walSize,
		"timestamp": time.Now().Unix(),
	}

	respondJSON(w, http.StatusOK, metrics)
}

// handleHealth returns health status
func (n *DHTNode) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"node_id": n.nodeID,
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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
