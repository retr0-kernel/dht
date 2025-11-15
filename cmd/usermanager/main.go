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

	"dht/internal/auth"
	"dht/internal/config"
	"dht/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection pool
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Database connection established")

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret, cfg.JWTExpiration)

	// Initialize user service
	userService := models.NewUserService(dbPool, authService)

	// Initialize API key service
	apiKeyService := models.NewAPIKeyService(dbPool)

	// Initialize handlers
	handler := NewHandler(userService, apiKeyService, authService, dbPool)

	// Setup router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", handler.Signup)
	mux.HandleFunc("POST /login", handler.Login)
	mux.HandleFunc("POST /apikeys", handler.CreateAPIKey)
	mux.HandleFunc("GET /apikeys", handler.ListAPIKeys)
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /validate-key", handler.ValidateAPIKey)
	mux.HandleFunc("GET /usage", handler.ListUsageRecords)
	mux.HandleFunc("GET /usage/stats", handler.GetUsageStats)

	// Wrap with middleware
	wrappedMux := LoggingMiddleware(CORSMiddleware(mux))

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.UserManagerPort),
		Handler:      wrappedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("User Manager service starting on port %s\n", cfg.UserManagerPort)
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
