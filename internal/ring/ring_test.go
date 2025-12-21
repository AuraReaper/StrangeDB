package ring

import (
	"fmt"
	"testing"
)

func TestAddNode(t *testing.T) {
	ring := New(150)

	ring.AddNode("http://node1:9000")
	ring.AddNode("http://node2:9000")
	ring.AddNode("http://node3:9000")

	nodes := ring.GetNodes()
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}

	stats := ring.Stats()
	if stats["virtual_nodes"].(int) != 450 {
		t.Errorf("Expected 450 virtual nodes, got %d", stats["virtual_nodes"])
	}
}

func TestGetNode(t *testing.T) {
	ring := New(150)

	ring.AddNode("http://node1:9000")
	ring.AddNode("http://node2:9000")
	ring.AddNode("http://node3:9000")

	// Same key should always return same node
	node1 := ring.GetNode("user:123")
	node2 := ring.GetNode("user:123")

	if node1 != node2 {
		t.Errorf("Same key returned different nodes: %s vs %s", node1, node2)
	}
}

func TestGetReplicas(t *testing.T) {
	ring := New(150)

	ring.AddNode("http://node1:9000")
	ring.AddNode("http://node2:9000")
	ring.AddNode("http://node3:9000")

	replicas := ring.GetReplicas("user:123", 3)

	if len(replicas) != 3 {
		t.Errorf("Expected 3 replicas, got %d", len(replicas))
	}

	// All replicas should be distinct
	seen := make(map[string]bool)
	for _, r := range replicas {
		if seen[r] {
			t.Errorf("Duplicate replica: %s", r)
		}
		seen[r] = true
	}
}

func TestDistribution(t *testing.T) {
	ring := New(150)

	ring.AddNode("http://node1:9000")
	ring.AddNode("http://node2:9000")
	ring.AddNode("http://node3:9000")

	// Generate many keys and count distribution
	counts := make(map[string]int)
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key:%d", i)
		node := ring.GetNode(key)
		counts[node]++
	}

	// Each node should have roughly 1/3 of keys
	for node, count := range counts {
		ratio := float64(count) / 10000.0
		// Allow 10% deviation
		if ratio < 0.25 || ratio > 0.42 {
			t.Errorf("Node %s has uneven distribution: %.2f%%", node, ratio*100)
		}
	}
}

func TestRemoveNode(t *testing.T) {
	ring := New(150)

	ring.AddNode("http://node1:9000")
	ring.AddNode("http://node2:9000")
	ring.AddNode("http://node3:9000")

	// Get node for a key
	keyBefore := ring.GetNode("test-key")

	// Remove that node
	ring.RemoveNode(keyBefore)

	// Key should now map to different node
	keyAfter := ring.GetNode("test-key")

	if keyAfter == keyBefore {
		t.Error("Key should map to different node after removal")
	}

	nodes := ring.GetNodes()
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes after removal, got %d", len(nodes))
	}
}
