package gossip

import (
	"sync"
	"time"
)

type NodeState int

const (
	Alive NodeState = iota
	Suspect
	Dead
)

type Member struct {
	NodeURL     string
	State       NodeState
	Heartbeat   int64
	LastUpdated time.Time
}

type Membership struct {
	mu      sync.RWMutex
	members map[string]*Member
	nodeURL string
}

func NewMembership(nodeURL string) *Membership {
	m := &Membership{
		members: make(map[string]*Member),
		nodeURL: nodeURL,
	}

	m.members[nodeURL] = &Member{
		NodeURL:     nodeURL,
		State:       Alive,
		Heartbeat:   0,
		LastUpdated: time.Now(),
	}

	return m
}

// return all alive members
func (m *Membership) GetMembers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var members []string
	for url, member := range m.members {
		if member.State == Alive {
			members = append(members, url)
		}
	}

	return members
}

func (m *Membership) GetAllMembers() map[string]*Member {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Member)
	for k, v := range m.members {
		result[k] = v
	}

	return result
}

func (m *Membership) UpdateMember(nodeURL string, heartbeat int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if member, ok := m.members[nodeURL]; ok {
		if heartbeat > member.Heartbeat {
			member.Heartbeat = heartbeat
			member.State = Alive
			member.LastUpdated = time.Now()
		}
	} else {
		m.members[nodeURL] = &Member{
			NodeURL:     nodeURL,
			State:       Alive,
			Heartbeat:   heartbeat,
			LastUpdated: time.Now(),
		}
	}
}

func (m *Membership) MarkSuspect(nodeURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if member, ok := m.members[nodeURL]; ok {
		member.State = Suspect
		member.LastUpdated = time.Now()
	}
}

func (m *Membership) MarkDead(nodeURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if member, ok := m.members[nodeURL]; ok {
		member.State = Dead
		member.LastUpdated = time.Now()
	}
}

func (m *Membership) IncrementHeartbeat() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if self, ok := m.members[m.nodeURL]; ok {
		self.Heartbeat++
		self.LastUpdated = time.Now()
		return self.Heartbeat
	}

	return 0
}

func (m *Membership) GetDigest() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	digest := make(map[string]int64)
	for url, member := range m.members {
		digest[url] = member.Heartbeat
	}

	return digest
}
