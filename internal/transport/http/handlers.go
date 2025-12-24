package http

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/AuraReaper/strangedb/internal/coordinator"
	"github.com/AuraReaper/strangedb/internal/gossip"
	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/ring"
	"github.com/AuraReaper/strangedb/internal/storage"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	coordinator *coordinator.Coordinator
	clock       *hlc.Clock
	nodeID      string
	startTime   time.Time
	gossiper    *gossip.Gossiper
	ring        *ring.ConsistentHashRing
}

func NewHandler(coord *coordinator.Coordinator, clock *hlc.Clock, nodeID string,
	gossiper *gossip.Gossiper, ring *ring.ConsistentHashRing) *Handler {
	return &Handler{
		coordinator: coord,
		clock:       clock,
		nodeID:      nodeID,
		startTime:   time.Now(),
		gossiper:    gossiper,
		ring:        ring,
	}
}

type SetKeyRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetKeyResponse struct {
	Success   bool          `json:"success"`
	Key       string        `json:"key"`
	Timestamp hlc.Timestamp `json:"timestamp"`
}

func (h *Handler) SetKey(c *fiber.Ctx) error {
	var req SetKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key is required")
	}

	ctx := context.Background()
	record, err := h.coordinator.Set(ctx, req.Key, []byte(req.Value))
	if err != nil {
		if err == coordinator.ErrQuorumNotReached {
			return fiber.NewError(fiber.StatusServiceUnavailable, "quorum not reached")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(SetKeyResponse{
		Success:   true,
		Key:       req.Key,
		Timestamp: record.Timestamp,
	})
}

type GetKeyResponse struct {
	Key       string        `json:"key"`
	Value     string        `json:"value"`
	Timestamp hlc.Timestamp `json:"timestamp"`
	Node      string        `json:"node"`
}

func (h *Handler) GetKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key is required")
	}

	ctx := context.Background()
	record, err := h.coordinator.Get(ctx, key)
	if err == storage.ErrKeyNotFound || err == storage.ErrKeyDeleted {
		return fiber.NewError(fiber.StatusNotFound, "key not found")
	}
	if err == coordinator.ErrQuorumNotReached {
		return fiber.NewError(fiber.StatusServiceUnavailable, "quorum not reached")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(GetKeyResponse{
		Key:       record.Key,
		Value:     string(record.Value),
		Timestamp: record.Timestamp,
		Node:      h.nodeID,
	})
}

type DeleteKeyResponse struct {
	Success   bool   `json:"success"`
	Key       string `json:"key"`
	Tombstone bool   `json:"tombstone_created"`
}

func (h *Handler) DeleteKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key is required")
	}

	ctx := context.Background()
	if err := h.coordinator.Delete(ctx, key); err != nil {
		if err == coordinator.ErrQuorumNotReached {
			return fiber.NewError(fiber.StatusServiceUnavailable, "quorum not reached")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(DeleteKeyResponse{
		Success:   true,
		Key:       key,
		Tombstone: true,
	})
}

type HealthResponse struct {
	Status        string `json:"status"`
	Node          string `json:"node"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(HealthResponse{
		Status:        "healthy",
		Node:          h.nodeID,
		UptimeSeconds: int64(time.Since(h.startTime).Seconds()),
	})
}

type StatusResponse struct {
	NodeID  string `json:"node_id"`
	Status  string `json:"status"`
	Storage string `json:"storage"`
}

func (h *Handler) Status(c *fiber.Ctx) error {
	return c.JSON(StatusResponse{
		NodeID:  h.nodeID,
		Status:  "running",
		Storage: "ok",
	})
}

type ClusterStatusResponse struct {
	NodeID  string       `json:"node_id"`
	Members []MemberInfo `json:"members"`
	Total   int          `json:"total"`
}

type MemberInfo struct {
	NodeID string `json:"node_id"`
	Addr   string `json:"addr"`
	Status string `json:"status"`
}

func (h *Handler) ClusterStatus(c *fiber.Ctx) error {
	var members []MemberInfo

	if h.gossiper != nil {
		for _, addr := range h.gossiper.GetMembers() {
			members = append(members, MemberInfo{
				NodeID: addr,
				Addr:   addr,
				Status: "alive",
			})
		}
	}

	return c.JSON(ClusterStatusResponse{
		NodeID:  h.nodeID,
		Members: members,
		Total:   len(members),
	})
}

func (h *Handler) RingStatus(c *fiber.Ctx) error {
	nodes := h.ring.GetNodes()
	return c.JSON(fiber.Map{
		"nodes":       nodes,
		"total_nodes": len(nodes),
	})
}

type ListKeysResponse struct {
	Keys  []KeyInfo `json:"keys"`
	Total int       `json:"total"`
}

type KeyInfo struct {
	Key       string        `json:"key"`
	Value     string        `json:"value"`
	Timestamp hlc.Timestamp `json:"timestamp"`
}

// returns all keys, optionally filtered by prefix and sorted
func (h *Handler) ListKeys(c *fiber.Ctx) error {
	prefix := c.Query("prefix", "")
	limitStr := c.Query("limit", "100")
	sortOrder := c.Query("sort", "asc") // asc or desc

	limit := 100
	if l, err := c.ParamsInt("limit"); err == nil && l > 0 {
		limit = l
	}
	// Try query param
	if limitStr != "" {
		if l := c.QueryInt("limit", 100); l > 0 {
			limit = l
		}
	}

	// Get records from local storage
	records, err := h.coordinator.Storage().List(prefix, 0) // 0 = no limit, we sort then apply limit
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Sort by key
	if sortOrder == "desc" {
		sort.Slice(records, func(i, j int) bool {
			return strings.Compare(records[i].Key, records[j].Key) > 0
		})
	} else {
		sort.Slice(records, func(i, j int) bool {
			return strings.Compare(records[i].Key, records[j].Key) < 0
		})
	}

	// Apply limit
	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}

	// Convert to response
	keys := make([]KeyInfo, len(records))
	for i, r := range records {
		keys[i] = KeyInfo{
			Key:       r.Key,
			Value:     string(r.Value),
			Timestamp: r.Timestamp,
		}
	}

	return c.JSON(ListKeysResponse{
		Keys:  keys,
		Total: len(keys),
	})
}
