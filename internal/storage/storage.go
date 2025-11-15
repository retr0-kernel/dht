package storage

import (
	"fmt"
	"sync"
	"time"
)

// Entry represents a key-value entry with metadata
type Entry struct {
	Key       string
	Value     []byte
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Storage provides in-memory key-value storage with TTL support
type Storage struct {
	data map[string]*Entry
	mu   sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	s := &Storage{
		data: make(map[string]*Entry),
	}

	// Start cleanup goroutine for expired entries
	go s.cleanupExpired()

	return s
}

// Set stores a key-value pair with optional TTL
func (s *Storage) Set(key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry := &Entry{
		Key:       key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set expiration if TTL provided
	if ttl > 0 {
		expiresAt := now.Add(ttl)
		entry.ExpiresAt = &expiresAt
	}

	s.data[key] = entry
	return nil
}

// Get retrieves a value by key
func (s *Storage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	// Check if expired
	if entry.ExpiresAt != nil && entry.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("key expired")
	}

	return entry.Value, nil
}

// Delete removes a key
func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		return fmt.Errorf("key not found")
	}

	delete(s.data, key)
	return nil
}

// Exists checks if a key exists
func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return false
	}

	// Check if expired
	if entry.ExpiresAt != nil && entry.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// KeyCount returns the number of keys in storage
func (s *Storage) KeyCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, entry := range s.data {
		// Only count non-expired entries
		if entry.ExpiresAt == nil || entry.ExpiresAt.After(now) {
			count++
		}
	}

	return count
}

// GetAll returns all non-expired entries (for WAL restore)
func (s *Storage) GetAll() map[string]*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*Entry)
	now := time.Now()

	for key, entry := range s.data {
		// Only include non-expired entries
		if entry.ExpiresAt == nil || entry.ExpiresAt.After(now) {
			result[key] = entry
		}
	}

	return result
}

// cleanupExpired removes expired entries periodically
func (s *Storage) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.data {
			if entry.ExpiresAt != nil && entry.ExpiresAt.Before(now) {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}
