package main

import (
	"sync"
	"time"
)

// TokenBucket implements a simple token bucket rate limiter
type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket
// maxTokens: maximum number of tokens (burst capacity)
// refillRate: tokens added per second
func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// AllowRequest checks if a request can proceed (consumes 1 token)
func (tb *TokenBucket) AllowRequest() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate

	// Cap at max tokens
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefill = now

	// Check if we have enough tokens
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// RateLimiterStore manages token buckets for each user
type RateLimiterStore struct {
	buckets map[int64]*TokenBucket
	mu      sync.RWMutex
}

// NewRateLimiterStore creates a new rate limiter store
func NewRateLimiterStore() *RateLimiterStore {
	store := &RateLimiterStore{
		buckets: make(map[int64]*TokenBucket),
	}

	// Start cleanup goroutine to remove old buckets
	go store.cleanup()

	return store
}

// AllowRequest checks if a request from userID should be allowed
func (rls *RateLimiterStore) AllowRequest(userID int64) bool {
	rls.mu.RLock()
	bucket, exists := rls.buckets[userID]
	rls.mu.RUnlock()

	if !exists {
		// Create new bucket for this user
		// 100 requests per minute = 100/60 = 1.67 requests per second
		// Burst capacity: 10 requests
		bucket = NewTokenBucket(10, 100.0/60.0)

		rls.mu.Lock()
		rls.buckets[userID] = bucket
		rls.mu.Unlock()
	}

	return bucket.AllowRequest()
}

// cleanup removes inactive buckets periodically
func (rls *RateLimiterStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rls.mu.Lock()
		// In a production system, you'd track last access time
		// and remove buckets that haven't been used recently
		// For now, we'll just keep all buckets
		rls.mu.Unlock()
	}
}
