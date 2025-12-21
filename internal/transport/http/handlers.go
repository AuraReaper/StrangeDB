package http

import (
	"encoding/base64"
	"time"
	"unicode/utf8"

	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/storage"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	storage   storage.Storage
	clock     *hlc.Clock
	nodeID    string
	startTime time.Time
}

func NewHandler(storage storage.Storage, clock *hlc.Clock, nodeID string) *Handler {
	return &Handler{
		storage:   storage,
		clock:     clock,
		nodeID:    nodeID,
		startTime: time.Now(),
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
			return fiber.NewError(fiber.StatusBadRequest, "invalid nase64 value")
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

	ts := h.clock.Now()
	record := &storage.Record{
		Key:       req.Key,
		Value:     value,
		Timestamp: ts,
		Tombstone: false,
	}

	if err := h.storage.Set(record); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(SetKeyResponse{
		Success:   true,
		Key:       req.Key,
		Timestamp: ts,
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

	record, err := h.storage.Get(key)
	if err == storage.ErrKeyNotFound {
		return fiber.NewError(fiber.StatusNotFound, "key not found")
	}
	if err == storage.ErrKeyDeleted {
		return fiber.NewError(fiber.StatusNotFound, "key not found")
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

	ts := h.clock.Now()
	if err := h.storage.Delete(key, ts); err != nil {
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
