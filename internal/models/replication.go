package models

import "time"

// ReplicationRequest represents a replication request
type ReplicationRequest struct {
	Key          string        `json:"key"`
	Value        []byte        `json:"value"`
	Operation    string        `json:"operation"` // "SET" or "DELETE"
	TTL          time.Duration `json:"ttl"`
	Consistency  string        `json:"consistency"` // "strong" or "eventual"
	PrimaryNode  string        `json:"primary_node"`
	ReplicaNodes []string      `json:"replica_nodes"`
	UserID       int64         `json:"user_id"`
}

// ReplicationResponse represents a replication response
type ReplicationResponse struct {
	Success     bool     `json:"success"`
	NodeID      string   `json:"node_id"`
	AckedNodes  []string `json:"acked_nodes,omitempty"`
	FailedNodes []string `json:"failed_nodes,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// ReplicationMetrics represents replication metrics
type ReplicationMetrics struct {
	TotalReplications  int64   `json:"total_replications"`
	SuccessfulReplicas int64   `json:"successful_replicas"`
	FailedReplicas     int64   `json:"failed_replicas"`
	QueueSize          int     `json:"queue_size"`
	AverageAckTime     float64 `json:"average_ack_time_ms"`
	MaxReplicationLag  float64 `json:"max_replication_lag_ms"`
	RetriesInProgress  int     `json:"retries_in_progress"`
}
