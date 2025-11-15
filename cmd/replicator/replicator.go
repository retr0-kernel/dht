package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"dht/internal/config"
	"dht/internal/models"
)

// ReplicationTask represents a replication task in the queue
type ReplicationTask struct {
	Request     *models.ReplicationRequest
	Retries     int
	MaxRetries  int
	EnqueuedAt  time.Time
	LastAttempt time.Time
}

// Replicator handles data replication across nodes
type Replicator struct {
	config     *config.Config
	httpClient *http.Client

	// Async replication queue
	eventualQueue chan *ReplicationTask
	retryQueue    chan *ReplicationTask

	// Metrics
	metrics struct {
		totalReplications  atomic.Int64
		successfulReplicas atomic.Int64
		failedReplicas     atomic.Int64
		ackTimes           []float64
		ackTimesMu         sync.Mutex
		maxLag             atomic.Int64 // in milliseconds
		retriesInProgress  atomic.Int32
	}

	// Control
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewReplicator creates a new replicator instance
func NewReplicator(cfg *config.Config) *Replicator {
	return &Replicator{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		eventualQueue: make(chan *ReplicationTask, 1000),
		retryQueue:    make(chan *ReplicationTask, 500),
		stopCh:        make(chan struct{}),
	}
}

// Start starts the background workers
func (r *Replicator) Start() {
	// Start eventual consistency workers
	for i := 0; i < 5; i++ {
		r.wg.Add(1)
		go r.eventualWorker()
	}

	// Start retry worker
	r.wg.Add(1)
	go r.retryWorker()

	log.Println("Replicator workers started")
}

// Stop stops all workers
func (r *Replicator) Stop() {
	close(r.stopCh)
	r.wg.Wait()
	log.Println("Replicator workers stopped")
}

// HandleReplicate handles replication requests
func (r *Replicator) HandleReplicate(w http.ResponseWriter, req *http.Request) {
	var replReq models.ReplicationRequest
	if err := json.NewDecoder(req.Body).Decode(&replReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if replReq.Key == "" || replReq.Operation == "" {
		respondError(w, http.StatusBadRequest, "Key and operation are required")
		return
	}

	if replReq.Consistency == "" {
		replReq.Consistency = "eventual"
	}

	r.metrics.totalReplications.Add(1)

	// Handle based on consistency level
	switch replReq.Consistency {
	case "eventual":
		r.handleEventualReplication(&replReq, w)
	case "strong":
		r.handleStrongReplication(&replReq, w)
	default:
		respondError(w, http.StatusBadRequest, "Invalid consistency level")
	}
}

// handleEventualReplication handles eventual consistency replication
func (r *Replicator) handleEventualReplication(replReq *models.ReplicationRequest, w http.ResponseWriter) {
	// Queue the replication task
	task := &ReplicationTask{
		Request:    replReq,
		MaxRetries: 3,
		EnqueuedAt: time.Now(),
	}

	select {
	case r.eventualQueue <- task:
		// Successfully queued
		respondJSON(w, http.StatusAccepted, models.ReplicationResponse{
			Success: true,
			NodeID:  "replicator",
		})
	default:
		// Queue is full
		respondError(w, http.StatusServiceUnavailable, "Replication queue is full")
	}
}

// handleStrongReplication handles strong consistency replication
func (r *Replicator) handleStrongReplication(replReq *models.ReplicationRequest, w http.ResponseWriter) {
	startTime := time.Now()

	// Calculate majority
	totalNodes := len(replReq.ReplicaNodes)
	majorityRequired := (totalNodes / 2) + 1

	// Replicate to all replica nodes concurrently
	results := make(chan bool, totalNodes)
	var ackedNodes []string
	var failedNodes []string
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, node := range replReq.ReplicaNodes {
		go func(nodeURL string) {
			success := r.replicateToNode(ctx, nodeURL, replReq)
			results <- success

			mu.Lock()
			if success {
				ackedNodes = append(ackedNodes, nodeURL)
			} else {
				failedNodes = append(failedNodes, nodeURL)
			}
			mu.Unlock()
		}(node)
	}

	// Wait for majority or timeout
	ackedCount := 0
	for i := 0; i < totalNodes; i++ {
		select {
		case success := <-results:
			if success {
				ackedCount++
				if ackedCount >= majorityRequired {
					// Majority achieved
					ackTime := time.Since(startTime).Milliseconds()
					r.recordAckTime(float64(ackTime))

					respondJSON(w, http.StatusOK, models.ReplicationResponse{
						Success:    true,
						NodeID:     "replicator",
						AckedNodes: ackedNodes,
					})
					return
				}
			}
		case <-ctx.Done():
			// Timeout
			respondError(w, http.StatusRequestTimeout, "Replication timeout - majority not reached")
			return
		}
	}

	// All nodes responded but majority not achieved
	if ackedCount < majorityRequired {
		respondError(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to achieve majority: %d/%d nodes acked", ackedCount, majorityRequired))
	}
}

// eventualWorker processes eventual consistency replication tasks
func (r *Replicator) eventualWorker() {
	defer r.wg.Done()

	for {
		select {
		case task := <-r.eventualQueue:
			r.processEventualTask(task)
		case <-r.stopCh:
			return
		}
	}
}

// processEventualTask processes a single eventual replication task
func (r *Replicator) processEventualTask(task *ReplicationTask) {
	startTime := time.Now()
	task.LastAttempt = startTime

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	successCount := 0
	for _, node := range task.Request.ReplicaNodes {
		if r.replicateToNode(ctx, node, task.Request) {
			successCount++
			r.metrics.successfulReplicas.Add(1)
		} else {
			r.metrics.failedReplicas.Add(1)
		}
	}

	// Calculate replication lag
	lag := time.Since(task.EnqueuedAt).Milliseconds()
	currentMaxLag := r.metrics.maxLag.Load()
	if lag > currentMaxLag {
		r.metrics.maxLag.Store(lag)
	}

	// If not all replicas succeeded and retries remaining, queue for retry
	if successCount < len(task.Request.ReplicaNodes) && task.Retries < task.MaxRetries {
		task.Retries++
		r.metrics.retriesInProgress.Add(1)

		// Add delay before retry
		go func() {
			time.Sleep(time.Duration(task.Retries) * 2 * time.Second)
			select {
			case r.retryQueue <- task:
			case <-r.stopCh:
			}
		}()
	} else if successCount > 0 {
		// At least one replica succeeded
		ackTime := time.Since(startTime).Milliseconds()
		r.recordAckTime(float64(ackTime))
	}
}

// retryWorker processes retry tasks
func (r *Replicator) retryWorker() {
	defer r.wg.Done()

	for {
		select {
		case task := <-r.retryQueue:
			r.metrics.retriesInProgress.Add(-1)
			log.Printf("Retrying replication for key=%s (attempt %d/%d)\n",
				task.Request.Key, task.Retries, task.MaxRetries)
			r.processEventualTask(task)
		case <-r.stopCh:
			return
		}
	}
}

// replicateToNode replicates data to a specific node
func (r *Replicator) replicateToNode(ctx context.Context, nodeURL string, replReq *models.ReplicationRequest) bool {
	var reqURL string
	var method string
	var body io.Reader

	switch replReq.Operation {
	case "SET":
		reqURL = fmt.Sprintf("%s/store/%s", nodeURL, replReq.Key)
		method = "PUT"
		body = bytes.NewReader(replReq.Value)
	case "DELETE":
		reqURL = fmt.Sprintf("%s/store/%s", nodeURL, replReq.Key)
		method = "DELETE"
		body = nil
	default:
		log.Printf("Unknown operation: %s\n", replReq.Operation)
		return false
	}

	// Add TTL if provided
	if replReq.TTL > 0 {
		reqURL = fmt.Sprintf("%s?ttl=%s", reqURL, replReq.TTL.String())
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		log.Printf("Failed to create request to %s: %v\n", nodeURL, err)
		return false
	}

	if replReq.Operation == "SET" {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	req.Header.Set("X-Replication", "true")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to replicate to %s: %v\n", nodeURL, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}

	log.Printf("Replication to %s failed with status %d\n", nodeURL, resp.StatusCode)
	return false
}

// recordAckTime records an acknowledgment time for metrics
func (r *Replicator) recordAckTime(ackTimeMs float64) {
	r.metrics.ackTimesMu.Lock()
	defer r.metrics.ackTimesMu.Unlock()

	r.metrics.ackTimes = append(r.metrics.ackTimes, ackTimeMs)

	// Keep only last 1000 samples
	if len(r.metrics.ackTimes) > 1000 {
		r.metrics.ackTimes = r.metrics.ackTimes[len(r.metrics.ackTimes)-1000:]
	}
}

// HandleMetrics returns replication metrics
func (r *Replicator) HandleMetrics(w http.ResponseWriter, req *http.Request) {
	r.metrics.ackTimesMu.Lock()
	avgAckTime := 0.0
	if len(r.metrics.ackTimes) > 0 {
		sum := 0.0
		for _, t := range r.metrics.ackTimes {
			sum += t
		}
		avgAckTime = sum / float64(len(r.metrics.ackTimes))
	}
	r.metrics.ackTimesMu.Unlock()

	metrics := models.ReplicationMetrics{
		TotalReplications:  r.metrics.totalReplications.Load(),
		SuccessfulReplicas: r.metrics.successfulReplicas.Load(),
		FailedReplicas:     r.metrics.failedReplicas.Load(),
		QueueSize:          len(r.eventualQueue),
		AverageAckTime:     avgAckTime,
		MaxReplicationLag:  float64(r.metrics.maxLag.Load()),
		RetriesInProgress:  int(r.metrics.retriesInProgress.Load()),
	}

	respondJSON(w, http.StatusOK, metrics)
}

// HandleHealth returns health status
func (r *Replicator) HandleHealth(w http.ResponseWriter, req *http.Request) {
	queueSize := len(r.eventualQueue)
	status := "healthy"

	// Mark as unhealthy if queue is too full
	if queueSize > 900 {
		status = "degraded"
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     status,
		"service":    "replicator",
		"queue_size": queueSize,
	})
}
