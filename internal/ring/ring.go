package ring

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
	"sync"
)

type ConsistentHashRing struct {
	mu           sync.RWMutex
	ring         map[uint64]string // hash -> nodeUrl
	sortedHashes []uint64
	nodes        map[string]bool
	vnodes       int
}

func New(vnodes int) *ConsistentHashRing {
	return &ConsistentHashRing{
		ring:   make(map[uint64]string),
		nodes:  make(map[string]bool),
		vnodes: vnodes,
	}
}

func (r *ConsistentHashRing) hash(key string) uint64 {
	h := md5.Sum([]byte(key))
	return binary.BigEndian.Uint64(h[:8])
}

func (r *ConsistentHashRing) AddNode(nodeURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nodes[nodeURL] {
		return
	}

	r.nodes[nodeURL] = true

	for i := 0; i < r.vnodes; i++ {
		virtualKey := fmt.Sprintf("%s:%d", nodeURL, i)
		hash := r.hash(virtualKey)
		r.ring[hash] = nodeURL
		r.sortedHashes = append(r.sortedHashes, hash)
	}

	slices.Sort(r.sortedHashes)
}

func (r *ConsistentHashRing) RemoveNode(nodeURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.nodes[nodeURL] {
		return
	}

	delete(r.nodes, nodeURL)

	newRing := make(map[uint64]string)
	var newHashes []uint64

	for hash, url := range r.ring {
		if url != nodeURL {
			newRing[hash] = url
			newHashes = append(newHashes, hash)
		}
	}

	r.ring = newRing
	r.sortedHashes = newHashes

	slices.Sort(r.sortedHashes)
}

func (r *ConsistentHashRing) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.ring) == 0 {
		return ""
	}

	hash := r.hash(key)

	// binary search for fisrt node >= hash
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})

	if idx == len(r.sortedHashes) {
		idx = 0
	}

	return r.ring[r.sortedHashes[idx]]
}

// return all physical nodes
func (r *ConsistentHashRing) GetNodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]string, 0, len(r.nodes))

	for node := range r.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

func (r *ConsistentHashRing) GetReplicas(key string, n int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.nodes) == 0 {
		return nil
	}

	if n > len(r.nodes) {
		n = len(r.nodes)
	}

	hash := r.hash(key)

	// find start pos
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})

	if idx == len(r.sortedHashes) {
		idx = 0
	}

	replicas := make([]string, 0, n)
	seen := make(map[string]bool)

	for len(replicas) < n {
		nodeURL := r.ring[r.sortedHashes[idx]]
		if !seen[nodeURL] {
			seen[nodeURL] = true
			replicas = append(replicas, nodeURL)
		}

		idx++
		if idx == len(r.sortedHashes) {
			idx = 0
		}
	}

	return replicas
}

func (r *ConsistentHashRing) Stats() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	distribution := make(map[string]int)
	for _, nodeURL := range r.ring {
		distribution[nodeURL]++
	}

	return map[string]any{
		"total_nodes":     len(r.nodes),
		"virtual_nodes":   len(r.ring),
		"vnodes_per_node": r.vnodes,
		"distribution":    distribution,
	}
}
