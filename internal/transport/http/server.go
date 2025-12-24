package http

import (
	"fmt"
	"time"

	"github.com/AuraReaper/strangedb/internal/telemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	app     *fiber.App
	handler *Handler
	port    int
}

func NewServer(handler *Handler, port int) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "StrangeDB",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))
	app.Use(metricsMiddleware())

	app.Get("/health", handler.Health)

	api := app.Group("/api/v1")
	api.Post("/kv", handler.SetKey)
	api.Get("/kv/:key", handler.GetKey)
	api.Delete("/kv/:key", handler.DeleteKey)
	api.Get("/status", handler.Status)
	api.Get("/cluster/status", handler.ClusterStatus)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/cluster/status", handler.ClusterStatus)
	app.Get("/cluster/ring", handler.RingStatus)

	api.Get("/keys", handler.ListKeys)

	return &Server{
		app:     app,
		handler: handler,
		port:    port,
	}
}

func (s *Server) Start() error {
	return s.app.Listen(fmt.Sprintf(":%d", s.port))
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func metricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}

		telemetry.RecordRequest(c.Method(), status, duration)

		return err
	}
}
