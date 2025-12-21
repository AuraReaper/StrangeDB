package http

import (
	"context"
	"encoding/base64"
	"time"
	"unicode/utf8"

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
	Key      string `json:"key"`
	Value    string `json:"value"`
	Encoding string `json:"encoding,omitempty"`
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

	var value []byte
	var err error

	if req.Encoding == "base64" {
		value, err = base64.StdEncoding.DecodeString(req.Value)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid base64 value")
		}
	} else {
		value = []byte(req.Value)
	}

	if req.Key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key is required")
	}

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "value must be base64 encoded")
	}

	ctx := context.Background()
	record, err := h.coordinator.Set(ctx, req.Key, value)
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
	ValueRaw  string        `json:"value_base64"`
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

	valueStr := string(record.Value)
	if !utf8.ValidString(valueStr) {
		valueStr = "[binary data]"
	}

	return c.JSON(GetKeyResponse{
		Key:       record.Key,
		Value:     valueStr,
		ValueRaw:  base64.StdEncoding.EncodeToString(record.Value),
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
