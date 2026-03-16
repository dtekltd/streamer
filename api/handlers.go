package api

import (
	"bufio"
	"errors"
	"fmt"

	"streamer/config"
	apphtml "streamer/html"
	"streamer/settings"
	"streamer/stream"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	streamService *stream.Service
	cfg           *config.AppConfig
}

func NewHandler(cfg *config.AppConfig) *Handler {
	return &Handler{streamService: stream.NewService(cfg), cfg: cfg}
}

func (h *Handler) AutoStartFromSavedSettings() error {
	saved, err := settings.Load()
	if err != nil {
		return err
	}
	if !saved.Saved {
		return errors.New("no saved settings found")
	}

	profile, err := settings.GetActiveProfile(saved)
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	req := stream.StartRequest{
		StreamKey:       saved.StreamKey,
		VideoPath:       profile.VideoPath,
		AudioDir:        profile.AudioDir,
		PlaylistOrder:   profile.PlaylistOrder,
		StreamEndMode:   profile.StreamEndMode,
		EndAfterMinutes: profile.EndAfterMinutes,
		FontPath:        profile.FontPath,
		TextX:           profile.TextX,
		TextY:           profile.TextY,
		NowPlayingLabel: profile.NowPlayingLabel,
		NextSongLabel:   profile.NextSongLabel,
		VideoCodec:      saved.VideoCodec,
		VideoPreset:     saved.VideoPreset,
		VideoBitrate:    saved.VideoBitrate,
		VideoMaxRate:    saved.VideoMaxRate,
		VideoBufSize:    saved.VideoBufSize,
	}

	if req.StreamKey == "" || req.VideoPath == "" || req.AudioDir == "" || req.FontPath == "" {
		return errors.New("saved settings are incomplete: stream key, video path, audio directory, and font path are required")
	}

	return h.streamService.Start(req)
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/", h.serveDashboard)
	app.Get("/internal/audio", h.handleInternalAudio)

	api := app.Group("/api")
	api.Post("/start", h.handleStartStream)
	api.Post("/stop", h.handleStopStream)
	api.Post("/preview-playlist", h.handlePreviewPlaylist)
	api.Post("/update-playlist", h.handleUpdatePlaylist)
	api.Get("/status", h.handleStatus)
	api.Get("/settings", h.handleGetSettings)
	api.Post("/settings", h.handleSaveSettings)
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
	if h.cfg.EnableLogging {
		fmt.Println("Stopping stream via web interface...")
	}
	h.streamService.Stop()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream stopping"})
}

func (h *Handler) handleUpdatePlaylist(c *fiber.Ctx) error {
	var req struct {
		PlaylistOrder   *string `json:"playlistOrder"`
		AudioDir        *string `json:"audioDir"`
		StreamEndMode   *string `json:"streamEndMode"`
		EndAfterMinutes *string `json:"endAfterMinutes"`
	}

	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
	}

	count, err := h.streamService.UpdatePlaylist(req.PlaylistOrder, req.AudioDir, req.StreamEndMode, req.EndAfterMinutes)
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

func (h *Handler) handlePreviewPlaylist(c *fiber.Ctx) error {
	var req struct {
		PlaylistOrder string `json:"playlistOrder"`
		AudioDir      string `json:"audioDir"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	playlist, err := h.streamService.PreviewPlaylist(req.AudioDir, req.PlaylistOrder)
	if err != nil {
		if errors.Is(err, stream.ErrNoSongsFound) {
			return c.Status(fiber.StatusBadRequest).SendString("No MP3 files found in the audio directory")
		}
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	return c.JSON(fiber.Map{
		"message":  "Playlist preview ready",
		"songs":    len(playlist),
		"playlist": playlist,
	})
}

func (h *Handler) handleInternalAudio(c *fiber.Ctx) error {
	c.Set("Content-Type", "audio/mpeg")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_ = h.streamService.StreamAudio(w)
	})
	return nil
}

func (h *Handler) handleGetSettings(c *fiber.Ctx) error {
	s, err := settings.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(s)
}

func (h *Handler) handleSaveSettings(c *fiber.Ctx) error {
	var s settings.DashboardSettings
	if err := c.BodyParser(&s); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := settings.Save(&s); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(fiber.Map{"message": "Settings saved"})
}
