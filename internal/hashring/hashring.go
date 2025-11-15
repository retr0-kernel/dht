package hashring

import (
	"hash/fnv"
	"sort"
	"sync"
)

// HashRing implements consistent hashing for node selection
type HashRing struct {
	nodes           []string          // Physical nodes
	virtualNodes    map[uint64]string // Virtual nodes (hash -> physical node)
	sortedHashes    []uint64          // Sorted hash values
	virtualReplicas int               // Number of virtual nodes per physical node
	replicationN    int               // Number of replicas for each key
	mu              sync.RWMutex
}

// NewHashRing creates a new hash ring with the given nodes
// virtualReplicas: number of virtual nodes per physical node (for distribution)
// replicationN: number of physical replicas for each key
func NewHashRing(nodes []string) *HashRing {
	ring := &HashRing{
		nodes:           nodes,
		virtualNodes:    make(map[uint64]string),
		virtualReplicas: 150, // 150 virtual nodes per physical node
		replicationN:    3,   // Store each key on 3 nodes
	}

	ring.addNodes(nodes)
	return ring
}

// addNodes adds nodes to the hash ring
func (hr *HashRing) addNodes(nodes []string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	for _, node := range nodes {
		// virtual nodes for better key distribution
		for i := 0; i < hr.virtualReplicas; i++ {
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

// GetNode returns the primary node responsible for the given key
func (hr *HashRing) GetNode(key string) string {
	nodes := hr.LocateKey(key, 1)
	if len(nodes) == 0 {
		return ""
	}
	return nodes[0]
}

// LocateKey returns primary and replica nodes for a key
// Returns up to n unique physical nodes
func (hr *HashRing) LocateKey(key string, n int) []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return nil
	}

	if n <= 0 {
		n = hr.replicationN
	}

	// Hash the key
	keyHash := hr.hash(key)

	// Binary search to find the first node with hash >= keyHash
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= keyHash
	})

	// Collect unique physical nodes
	seen := make(map[string]bool)
	result := make([]string, 0, n)

	for len(result) < n && len(result) < len(hr.nodes) {
		if idx >= len(hr.sortedHashes) {
			idx = 0
		}

		node := hr.virtualNodes[hr.sortedHashes[idx]]
		if !seen[node] {
			seen[node] = true
			result = append(result, node)
		}

		idx++
	}

	return result
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
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Check if node already exists
	for _, n := range hr.nodes {
		if n == node {
			return
		}
	}

	// Add to physical nodes
	hr.nodes = append(hr.nodes, node)

	// Create virtual nodes
	for i := 0; i < hr.virtualReplicas; i++ {
		virtualKey := node + ":" + string(rune(i))
		hash := hr.hash(virtualKey)
		hr.virtualNodes[hash] = node
		hr.sortedHashes = append(hr.sortedHashes, hash)
	}

	// Re-sort hashes
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

// RemoveNode removes a node from the ring
func (hr *HashRing) RemoveNode(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Remove virtual nodes
	newSortedHashes := make([]uint64, 0)
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

// hash computes FNV-1a hash (fast and good distribution)
func (hr *HashRing) hash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

// NodeCount returns the number of physical nodes
func (hr *HashRing) NodeCount() int {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return len(hr.nodes)
}
