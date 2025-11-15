package hashring

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"sync"
)

// HashRing implements consistent hashing for node selection
type HashRing struct {
	nodes        []string          // Physical nodes
	virtualNodes map[uint32]string // Virtual nodes (hash -> physical node)
	sortedHashes []uint32          // Sorted hash values
	replicas     int               // Number of virtual nodes per physical node
	mu           sync.RWMutex
}

// NewHashRing creates a new hash ring with the given nodes
func NewHashRing(nodes []string) *HashRing {
	ring := &HashRing{
		nodes:        nodes,
		virtualNodes: make(map[uint32]string),
		replicas:     150, // 150 virtual nodes per physical node for better distribution
	}

	ring.addNodes(nodes)
	return ring
}

// addNodes adds nodes to the hash ring
func (hr *HashRing) addNodes(nodes []string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	for _, node := range nodes {
		// Create virtual nodes for better key distribution
		for i := 0; i < hr.replicas; i++ {
			virtualKey := node + ":" + string(rune(i))
			hash := hr.hash(virtualKey)
			hr.virtualNodes[hash] = node
			hr.sortedHashes = append(hr.sortedHashes, hash)
		}
	}

	// Sort hashes for binary search
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

// GetNode returns the node responsible for the given key
func (hr *HashRing) GetNode(key string) string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return ""
	}

	// Hash the key
	keyHash := hr.hash(key)

	// Binary search to find the first node with hash >= keyHash
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= keyHash
	})

	// Wrap around if necessary
	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	return hr.virtualNodes[hr.sortedHashes[idx]]
}

// GetAllNodes returns all physical nodes in the ring
func (hr *HashRing) GetAllNodes() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	nodes := make([]string, len(hr.nodes))
	copy(nodes, hr.nodes)
	return nodes
}

// AddNode adds a new node to the ring
func (hr *HashRing) AddNode(node string) {
	hr.addNodes([]string{node})
}

// RemoveNode removes a node from the ring
func (hr *HashRing) RemoveNode(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Remove virtual nodes
	newSortedHashes := make([]uint32, 0)
	for _, hash := range hr.sortedHashes {
		if hr.virtualNodes[hash] != node {
			newSortedHashes = append(newSortedHashes, hash)
		} else {
			delete(hr.virtualNodes, hash)
		}
	}
	hr.sortedHashes = newSortedHashes

	// Remove from physical nodes
	newNodes := make([]string, 0)
	for _, n := range hr.nodes {
		if n != node {
			newNodes = append(newNodes, n)
		}
	}
	hr.nodes = newNodes
}

// hash computes SHA-256 hash and returns first 4 bytes as uint32
func (hr *HashRing) hash(key string) uint32 {
	h := sha256.New()
	h.Write([]byte(key))
	hashBytes := h.Sum(nil)
	return binary.BigEndian.Uint32(hashBytes[:4])
}
