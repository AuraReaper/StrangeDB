package hlc

import "testing"

func TestNow(t *testing.T) {
	clock := NewClock("node-1")

	ts1 := clock.Now()
	ts2 := clock.Now()

	if !IsAfter(ts2, ts1) {
		t.Errorf("ts2 should be after ts1")
	}
}

func TestUpdate(t *testing.T) {
	clock1 := NewClock("node-1")
	clock2 := NewClock("node-2")

	ts1 := clock1.Now()
	ts2 := clock2.Update(ts1)

	if !IsAfter(ts2, ts1) {
		t.Errorf("ts2 should be after ts1 after update")
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b     Timestamp
		expected int
	}{
		{
			Timestamp{WallTime: 100, Logical: 0, NodeID: "a"},
			Timestamp{WallTime: 200, Logical: 0, NodeID: "a"},
			-1,
		},
		{
			Timestamp{WallTime: 100, Logical: 5, NodeID: "a"},
			Timestamp{WallTime: 100, Logical: 3, NodeID: "a"},
			1,
		},
		{
			Timestamp{WallTime: 100, Logical: 0, NodeID: "a"},
			Timestamp{WallTime: 200, Logical: 0, NodeID: "b"},
			-1,
		},
	}

	for _, tt := range tests {
		result := Compare(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Compare(%v, %v) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMontonicity(t *testing.T) {
	clock := NewClock("node-1")
	var last Timestamp

	for i := 0; i < 1000; i++ {
		ts := clock.Now()
		if i > 0 && !IsAfter(ts, last) {
			t.Errorf("timestamp %d not after previous", i)
		}
		last = ts
	}
}
