package api

import (
	"bufio"
	"errors"

	"streamer/config"
	apphtml "streamer/html"
	"streamer/stream"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	streamService *stream.Service
	serverMode    string
}

func NewHandler(cfg config.AppConfig) *Handler {
	return &Handler{streamService: stream.NewService(cfg), serverMode: cfg.ServerMode}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/", h.serveDashboard)
	app.Get("/internal/audio", h.handleInternalAudio)

	api := app.Group("/api")
	api.Post("/start", h.handleStartStream)
	api.Post("/stop", h.handleStopStream)
	api.Post("/update-playlist", h.handleUpdatePlaylist)
	api.Get("/status", h.handleStatus)
}

func (h *Handler) serveDashboard(c *fiber.Ctx) error {
	c.Type("html", "utf-8")
	return c.SendString(apphtml.RenderDashboard(h.serverMode))
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

func (h *Handler) handleUpdatePlaylist(c *fiber.Ctx) error {
	count, err := h.streamService.UpdatePlaylist()
	if err != nil {
		if errors.Is(err, stream.ErrStreamNotRunning) {
			return c.Status(fiber.StatusBadRequest).SendString("Stream is not running")
		}
		if errors.Is(err, stream.ErrNoSongsFound) {
			return c.Status(fiber.StatusBadRequest).SendString("No MP3 files found in the audio directory")
		}
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fiber.Map{"message": "Playlist updated", "songs": count})
}

func (h *Handler) handleInternalAudio(c *fiber.Ctx) error {
	c.Set("Content-Type", "audio/mpeg")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_ = h.streamService.StreamAudio(w)
	})
	return nil
}
