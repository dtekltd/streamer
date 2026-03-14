package api

import (
	"errors"

	apphtml "streamer/html"
	"streamer/stream"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	streamService *stream.Service
}

func NewHandler() *Handler {
	return &Handler{streamService: stream.NewService()}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/", h.serveDashboard)

	api := app.Group("/api")
	api.Post("/start", h.handleStartStream)
	api.Post("/stop", h.handleStopStream)
	api.Get("/status", h.handleStatus)
}

func (h *Handler) serveDashboard(c *fiber.Ctx) error {
	c.Type("html", "utf-8")
	return c.SendString(apphtml.Dashboard)
}

func (h *Handler) handleStatus(c *fiber.Ctx) error {
	return c.JSON(h.streamService.Status())
}

func (h *Handler) handleStartStream(c *fiber.Ctx) error {
	var req stream.StartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if req.StreamKey == "" || req.VideoPath == "" || req.AudioDir == "" || req.FontPath == "" {
		return c.Status(fiber.StatusBadRequest).SendString("All fields are required")
	}

	if err := h.streamService.Start(req); err != nil {
		if errors.Is(err, stream.ErrAlreadyRunning) {
			return c.Status(fiber.StatusConflict).SendString("Stream is already running")
		}
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream starting"})
}

func (h *Handler) handleStopStream(c *fiber.Ctx) error {
	h.streamService.Stop()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream stopping"})
}
