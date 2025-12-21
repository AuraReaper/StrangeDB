package gossip

import (
	"math/rand/v2"
	"sync"
	"time"
)

type Gossiper struct {
	mu         sync.RWMutex
	membership *Membership
	nodeURL    string
	peers      []string
	interval   time.Duration
	timeout    time.Duration
	stopCh     chan struct{}

	onMembershipChange func([]string)
}

func New(nodeURL string, seeds []string, intreval time.Duration) *Gossiper {
	g := &Gossiper{
		membership: NewMembership(nodeURL),
		nodeURL:    nodeURL,
		peers:      seeds,
		interval:   intreval,
		timeout:    5 * time.Second,
		stopCh:     make(chan struct{}),
	}

	for _, seed := range seeds {
		if seed != nodeURL {
			g.membership.UpdateMember(seed, 0)
		}
	}

	return g
}

func (g *Gossiper) SetMembershipChangeCallback(fn func([]string)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onMembershipChange = fn
}

func (g *Gossiper) Start() {
	go g.gossipLoop()
	go g.failureDetectionLoop()
}

func (g *Gossiper) Stop() {
	close(g.stopCh)
}

func (g *Gossiper) GetMembers() []string {
	return g.membership.GetMembers()
}

func (g *Gossiper) gossipLoop() {
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.gossipRound()
		case <-g.stopCh:
			return
		}
	}
}

func (g *Gossiper) gossipRound() {
	g.membership.IncrementHeartbeat()

	peers := g.membership.GetMembers()
	if len(peers) < 1 {
		return
	}

	var candidates []string
	for _, peer := range peers {
		if peer != g.nodeURL {
			candidates = append(candidates, peer)
		}
	}

	if len(candidates) == 0 {
		return
	}

	target := candidates[rand.IntN(len(candidates))]
	g.gossipWith(target)
}

func (g *Gossiper) gossipWith(targetURL string) {
	g.membership.UpdateMember(targetURL, time.Now().Unix())
}

func (g *Gossiper) failureDetectionLoop() {
	ticker := time.NewTicker(g.interval * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.checkFailures()
		case <-g.stopCh:
			return
		}
	}
}

func (g *Gossiper) checkFailures() {
	now := time.Now()
	suspectThreshold := g.interval * 5
	deadThreshold := g.interval * 10

	allMembers := g.membership.GetAllMembers()
	membershipChanged := false

	for url, member := range allMembers {
		if url == g.nodeURL {
			continue
		}

		age := now.Sub(member.LastUpdated)

		switch member.State {
		case Alive:
			if age > suspectThreshold {
				g.membership.MarkSuspect(url)
				membershipChanged = true
			}
		case Suspect:
			if age > deadThreshold {
				g.membership.MarkDead(url)
				membershipChanged = true
			}
		}
	}

	if membershipChanged {
		g.mu.RLock()
		callback := g.onMembershipChange
		defer g.mu.RUnlock()

		if callback != nil {
			callback(g.membership.GetMembers())
		}
	}
}

func (g *Gossiper) HandleGossip(digest map[string]int64) map[string]int64 {
	for url, heartbeat := range digest {
		g.membership.UpdateMember(url, heartbeat)
	}

	return g.membership.GetDigest()
}
