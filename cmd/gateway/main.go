package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dht/internal/config"
	"dht/internal/hashring"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize hash ring with DHT nodes
	// In production, this would come from service discovery for now we use localhosts
	nodes := []string{
		"http://localhost:8082", // dhtnode-1
		"http://localhost:8083", // dhtnode-2
		"http://localhost:8084", // dhtnode-3
	}

	ring := hashring.NewHashRing(nodes)
	log.Printf("Hash ring initialized with %d nodes\n", len(nodes))

	// Initialize rate limiter store
	rateLimiterStore := NewRateLimiterStore()

	// Initialize handlers
	handler := NewHandler(cfg, ring, rateLimiterStore)

	// Setup router
	mux := http.NewServeMux()

	// KV routes
	mux.HandleFunc("PUT /v1/kv/{key}", handler.PutKey)
	mux.HandleFunc("GET /v1/kv/{key}", handler.GetKey)
	mux.HandleFunc("DELETE /v1/kv/{key}", handler.DeleteKey)
	mux.HandleFunc("GET /v1/kv", handler.ListKeys)

	// Health check
	mux.HandleFunc("GET /health", handler.Health)

	// Wrap with middleware (order matters: logging -> CORS -> auth -> rate limit -> handler)
	wrappedMux := LoggingMiddleware(
		CORSMiddleware(
			AuthMiddleware(cfg, rateLimiterStore)(mux),
		),
	)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.GatewayPort),
		Handler:      wrappedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Gateway service starting on port %s\n", cfg.GatewayPort)
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
