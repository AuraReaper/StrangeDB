package hlc

import (
	"sync"
	"time"
)

type Timestamp struct {
	WallTime int64  `json:"wall_time"` // time in nanosec
	Logical  uint32 `json:"logical"`   // counter
	NodeID   string `json:"node_id"`   // node genrating the Timestamp
}

// hybrid logical clock
type Clock struct {
	mu     sync.Mutex
	nodeID string
	last   Timestamp
}

func NewClock(nodeID string) *Clock {
	return &Clock{
		nodeID: nodeID,
		last: Timestamp{
			WallTime: 0,
			Logical:  0,
			NodeID:   nodeID,
		},
	}
}

// genrates a new timestamp for local event
func (c *Clock) Now() Timestamp {
	c.mu.Lock()
	defer c.mu.Unlock()

	physicalTime := time.Now().UnixNano()

	if physicalTime > c.last.WallTime {
		c.last = Timestamp{
			WallTime: physicalTime,
			Logical:  0,
			NodeID:   c.nodeID,
		}
	} else {
		c.last = Timestamp{
			WallTime: c.last.WallTime,
			Logical:  c.last.Logical + 1,
			NodeID:   c.nodeID,
		}
	}

	return c.last
}

// upadtes clock based on given timestamp
func (c *Clock) Update(received Timestamp) Timestamp {
	c.mu.Lock()
	defer c.mu.Unlock()

	physicalTime := time.Now().UnixNano()

	if physicalTime > c.last.WallTime && physicalTime > received.WallTime {
		c.last = Timestamp{
			WallTime: physicalTime,
			Logical:  0,
			NodeID:   c.nodeID,
		}
	} else if c.last.WallTime > physicalTime && c.last.WallTime > received.WallTime {
		c.last = Timestamp{
			WallTime: c.last.WallTime,
			Logical:  c.last.Logical + 1,
			NodeID:   c.nodeID,
		}
	} else if received.WallTime > physicalTime && received.WallTime > c.last.WallTime {
		c.last = Timestamp{
			WallTime: received.WallTime,
			Logical:  received.Logical + 1,
			NodeID:   c.nodeID,
		}
	} else if c.last.WallTime == received.WallTime {
		maxLogical := max(received.Logical, c.last.Logical)
		c.last = Timestamp{
			WallTime: c.last.WallTime,
			Logical:  maxLogical + 1,
			NodeID:   c.nodeID,
		}
	}

	return c.last
}

// compare 2 timestamps
func Compare(a, b Timestamp) int {
	if a.WallTime < b.WallTime {
		return -1
	}

	if a.WallTime > b.WallTime {
		return 1
	}

	if a.Logical < b.Logical {
		return -1
	}

	if a.Logical > b.Logical {
		return 1
	}

	if a.NodeID < b.NodeID {
		return -1
	}

	if a.NodeID > b.NodeID {
		return 1
	}

	return 0
}

func IsAfter(a, b Timestamp) bool {
	return Compare(a, b) > 0
}

func IsBefore(a, b Timestamp) bool {
	return Compare(a, b) < 0
}
