package storage

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"
)

// WALEntry represents a write-ahead log entry
type WALEntry struct {
	Operation string // "SET" or "DELETE"
	Key       string
	Value     []byte
	TTL       time.Duration
	Timestamp time.Time
}

// WAL implements write-ahead logging
type WAL struct {
	file     *os.File
	encoder  *gob.Encoder
	filepath string
	mu       sync.Mutex
}

// NewWAL creates or opens a WAL file
func NewWAL(filepath string) (*WAL, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %w", err)
	}

	return &WAL{
		file:     file,
		encoder:  gob.NewEncoder(file),
		filepath: filepath,
	}, nil
}

// Append writes an entry to the WAL
func (w *WAL) Append(operation, key string, value []byte, ttl time.Duration) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry := WALEntry{
		Operation: operation,
		Key:       key,
		Value:     value,
		TTL:       ttl,
		Timestamp: time.Now(),
	}

	if err := w.encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode WAL entry: %w", err)
	}

	// Sync to disk for durability
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL: %w", err)
	}

	return nil
}

// Restore reads the WAL and applies entries to storage
func (w *WAL) Restore(storage *Storage) error {
	// Open file for reading
	file, err := os.Open(w.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// WAL doesn't exist yet, that's okay
			return nil
		}
		return fmt.Errorf("failed to open WAL for restore: %w", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(bufio.NewReader(file))
	entriesRestored := 0
	now := time.Now()

	for {
		var entry WALEntry
		err := decoder.Decode(&entry)
		if err != nil {
			// EOF is expected
			if err.Error() == "EOF" {
				break
			}
			// Skip corrupted entries and continue
			continue
		}

		// Check if entry is expired
		if entry.TTL > 0 {
			expiresAt := entry.Timestamp.Add(entry.TTL)
			if expiresAt.Before(now) {
				// Skip expired entry
				continue
			}
		}

		// Apply operation
		switch entry.Operation {
		case "SET":
			storage.Set(entry.Key, entry.Value, entry.TTL)
			entriesRestored++
		case "DELETE":
			storage.Delete(entry.Key)
		}
	}

	fmt.Printf("WAL: Restored %d entries from %s\n", entriesRestored, w.filepath)
	return nil
}

// Size returns the size of the WAL file in bytes
func (w *WAL) Size() (int64, error) {
	info, err := os.Stat(w.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return info.Size(), nil
}

// Close closes the WAL file
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

// Truncate creates a new WAL file (after compaction/snapshot)
func (w *WAL) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close current file
	w.file.Close()

	// Remove old file
	if err := os.Remove(w.filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old WAL: %w", err)
	}

	// Create new file
	file, err := os.OpenFile(w.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new WAL: %w", err)
	}

	w.file = file
	w.encoder = gob.NewEncoder(file)

	return nil
}
