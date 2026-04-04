package api

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"streamer/config"
	apphtml "streamer/html"
	"streamer/settings"
	"streamer/stream"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	streamService *stream.Service
	cfg           *config.AppConfig
	sessionMu     sync.Mutex
	sessions      map[string]time.Time
}

const sessionTTL = 24 * time.Hour

const authCookieName = "streamer_auth"

var spaDistPath = filepath.Join("frontend", "dist", "spa")

func NewHandler(cfg *config.AppConfig) *Handler {
	return &Handler{streamService: stream.NewService(cfg), cfg: cfg, sessions: map[string]time.Time{}}
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
		ProfileID:          profile.ID,
		StreamKey:          profile.StreamKey,
		StreamURLTemplate:  profile.StreamURLTemplate,
		VideoPath:          profile.VideoPath,
		EnableVideoAudio:   profile.EnableVideoAudio,
		VideoAudioVolume:   profile.VideoAudioVolume,
		AudioDir:           profile.AudioDir,
		FFmpegArgs:         profile.FFmpegArgs,
		PlaylistOrder:      profile.PlaylistOrder,
		StreamEndMode:      profile.StreamEndMode,
		EndAfterMinutes:    profile.EndAfterMinutes,
		FontPath:           profile.FontPath,
		TextX:              profile.TextX,
		TextY:              profile.TextY,
		EnablePlayingLabel: profile.EnablePlayingLabel,
		NowPlayingLabel:    profile.NowPlayingLabel,
		EnableNextLabel:    profile.EnableNextLabel,
		NextSongLabel:      profile.NextSongLabel,
	}
	if req.StreamKey == "" {
		req.StreamKey = saved.StreamKey
	}
	if req.StreamURLTemplate == "" {
		req.StreamURLTemplate = saved.StreamURLTemplate
	}

	if req.StreamKey == "" {
		return errors.New("saved settings are incomplete: stream key is required")
	}

	return h.streamService.Start(req)
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Static("/", spaDistPath)
	app.Get("/", h.serveDashboard)
	app.Get("/profiles", h.serveDashboard)
	app.Get("/streaming", h.serveDashboard)
	app.Get("/internal/audio/:profileId", h.handleInternalAudio)

	api := app.Group("/api")
	api.Post("/auth/login", h.handleLogin)
	api.Post("/auth/logout", h.requireAuth, h.handleLogout)
	api.Get("/auth/me", h.requireAuth, h.handleAuthMe)

	api.Use(h.requireAuth)
	api.Post("/start", h.handleStartStream)
	api.Post("/stop", h.handleStopStream)
	api.Post("/update-playlist", h.handleUpdatePlaylist)
	api.Get("/status", h.handleStatus)
	api.Get("/settings", h.handleGetSettings)
	api.Post("/settings", h.handleSaveSettings)

	app.Get("/*", h.serveDashboard)
}

func (h *Handler) serveDashboard(c *fiber.Ctx) error {
	if strings.HasPrefix(c.Path(), "/api/") || strings.HasPrefix(c.Path(), "/internal/") {
		return c.SendStatus(fiber.StatusNotFound)
	}

	indexPath := filepath.Join(spaDistPath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		return c.SendFile(indexPath)
	}

	c.Type("html", "utf-8")
	return c.SendString(apphtml.Dashboard)
}

func (h *Handler) requireAuth(c *fiber.Ctx) error {
	token := extractAuthToken(c)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	h.sessionMu.Lock()
	expiresAt, ok := h.sessions[token]
	if !ok || time.Now().After(expiresAt) {
		delete(h.sessions, token)
		h.sessionMu.Unlock()
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	h.sessions[token] = time.Now().Add(sessionTTL)
	h.sessionMu.Unlock()

	return c.Next()
}

func (h *Handler) handleLogin(c *fiber.Ctx) error {
	var req struct {
		PIN string `json:"pin"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	provided := strings.TrimSpace(req.PIN)
	expected := strings.TrimSpace(h.cfg.LoginPIN)
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid pin"})
	}

	token, err := randomToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	h.sessionMu.Lock()
	h.sessions[token] = time.Now().Add(sessionTTL)
	h.sessionMu.Unlock()

	c.Cookie(&fiber.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{"token": token})
}

func (h *Handler) handleLogout(c *fiber.Ctx) error {
	token := extractAuthToken(c)
	if token != "" {
		h.sessionMu.Lock()
		delete(h.sessions, token)
		h.sessionMu.Unlock()
	}

	c.Cookie(&fiber.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(-time.Hour),
	})

	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *Handler) handleAuthMe(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"authenticated": true})
}

func extractAuthToken(c *fiber.Ctx) string {
	if value := strings.TrimSpace(c.Get("X-Auth-Token")); value != "" {
		return value
	}

	authorization := strings.TrimSpace(c.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		return strings.TrimSpace(authorization[7:])
	}

	return strings.TrimSpace(c.Cookies(authCookieName))
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) handleStatus(c *fiber.Ctx) error {
	profileID := c.Query("profileId")
	return c.JSON(fiber.Map{
		"current": h.streamService.Status(profileID),
		"streams": h.streamService.Statuses(),
	})
}

func (h *Handler) handleStartStream(c *fiber.Ctx) error {
	var req stream.StartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if req.StreamKey == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Stream key is required")
	}

	if err := h.streamService.Start(req); err != nil {
		if errors.Is(err, stream.ErrAlreadyRunning) {
			return c.Status(fiber.StatusConflict).SendString("Stream is already running")
		}
		if errors.Is(err, stream.ErrNoMediaInput) || errors.Is(err, stream.ErrNoSongsFound) {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream starting"})
}

func (h *Handler) handleStopStream(c *fiber.Ctx) error {
	var req struct {
		ProfileID string `json:"profileId"`
	}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
	}

	if h.cfg.EnableLogging {
		fmt.Println("Stopping stream via web interface...")
	}
	h.streamService.Stop(req.ProfileID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream stopping"})
}

func (h *Handler) handleUpdatePlaylist(c *fiber.Ctx) error {
	var req struct {
		ProfileID       string  `json:"profileId"`
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

	playlist, err := h.streamService.UpdatePlaylist(req.ProfileID, req.PlaylistOrder, req.AudioDir, req.StreamEndMode, req.EndAfterMinutes)
	if err != nil {
		if errors.Is(err, stream.ErrStreamNotRunning) {
			return c.Status(fiber.StatusBadRequest).SendString("Stream is not running")
		}
		if errors.Is(err, stream.ErrNoSongsFound) {
			return c.Status(fiber.StatusBadRequest).SendString("No MP3 files found in the audio directory")
		}
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// return c.JSON(fiber.Map{"message": "Playlist updated", "songs": count})
	return c.JSON(fiber.Map{
		"songs":    len(playlist),
		"playlist": playlist,
	})
}

func (h *Handler) handleInternalAudio(c *fiber.Ctx) error {
	profileID := c.Params("profileId")
	c.Set("Content-Type", "audio/mpeg")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_ = h.streamService.StreamAudio(profileID, w)
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
