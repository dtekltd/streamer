package api

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apphtml "streamer/html"
	"streamer/models"
	"streamer/utils"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	state *models.StreamState
}

func NewHandler() *Handler {
	return &Handler{state: &models.StreamState{}}
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
	h.state.Mu.Lock()
	defer h.state.Mu.Unlock()

	return c.JSON(h.state)
}

func (h *Handler) handleStartStream(c *fiber.Ctx) error {
	h.state.Mu.Lock()
	if h.state.IsRunning {
		h.state.Mu.Unlock()
		return c.Status(fiber.StatusConflict).SendString("Stream is already running")
	}
	h.state.Mu.Unlock()

	var req models.StartStreamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if req.StreamKey == "" || req.VideoPath == "" || req.AudioDir == "" || req.FontPath == "" {
		return c.Status(fiber.StatusBadRequest).SendString("All fields are required")
	}

	textX := strings.TrimSpace(req.TextX)
	if textX == "" {
		textX = "30"
	}

	textY := strings.TrimSpace(req.TextY)
	if textY == "" {
		textY = "h-th-30"
	}

	go h.runStream(req.StreamKey, req.VideoPath, req.AudioDir, req.FontPath, textX, textY)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream starting"})
}

func (h *Handler) handleStopStream(c *fiber.Ctx) error {
	h.state.Mu.Lock()
	if h.state.IsRunning && h.state.Cancel != nil {
		fmt.Println("Stopping stream via web interface...")
		h.state.Cancel()
	}
	h.state.Mu.Unlock()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Stream stopping"})
}

func (h *Handler) runStream(streamKey, videoFile, audioDir, fontFile, textX, textY string) {
	h.state.Mu.Lock()
	h.state.IsRunning = true
	ctx, cancel := context.WithCancel(context.Background())
	h.state.Cancel = cancel
	h.state.Mu.Unlock()

	defer func() {
		h.state.Mu.Lock()
		h.state.IsRunning = false
		h.state.CurrentSong = ""
		h.state.Cancel = nil
		h.state.Mu.Unlock()
		fmt.Println("Stream fully stopped and cleaned up.")
	}()

	songs, err := prepareAudioList(audioDir)
	if err != nil || len(songs) == 0 {
		fmt.Printf("Error preparing audio: %v\n", err)
		return
	}

	concatFilePath, err := buildConcatFile(songs)
	if err != nil {
		fmt.Printf("Error building concat file: %v\n", err)
		return
	}
	defer os.Remove(concatFilePath)
	defer os.Remove(models.NowPlayingFile)

	go h.manageNowPlayingText(ctx, songs)

	// Switch to YouTube RTMP when ready:
	// rtmpURL := fmt.Sprintf("rtmp://a.rtmp.youtube.com/live2/%s", streamKey)
	rtmpURL := fmt.Sprintf("rtmp://10.16.0.165:1935/live/%s", streamKey)
	safeFontPath := filepath.ToSlash(fontFile)                  // Convert Windows backslashes to forward slashes
	safeFontPath = strings.Replace(safeFontPath, ":", "\\:", 1) // Escape the Windows drive colon (e.g., turns "E:/" into "E\:/")
	drawtextFilter := fmt.Sprintf("drawtext=fontfile='%s':textfile='%s':reload=1:fontcolor=white:fontsize=40:box=1:boxcolor=black@0.6:boxborderw=10:x=%s:y=%s", safeFontPath, models.NowPlayingFile, textX, textY)

	args := []string{
		"-re",
		"-stream_loop", "-1",
		"-i", videoFile,
		"-stream_loop", "-1",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFilePath,
		"-filter_complex", fmt.Sprintf("[0:v]%s[v]", drawtextFilter),
		"-map", "[v]",
		"-map", "1:a",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-b:v", "3000k",
		"-maxrate", "3000k",
		"-bufsize", "6000k",
		"-pix_fmt", "yuv420p",
		"-g", "50",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-f", "flv",
		rtmpURL,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("FFmpeg exited: %v\n", err)
	}
}

func (h *Handler) manageNowPlayingText(ctx context.Context, songs []models.Song) {
	for {
		for _, song := range songs {
			select {
			case <-ctx.Done():
				return
			default:
				text := fmt.Sprintf("Now Playing: %s", song.Name)
				if err := os.WriteFile(models.NowPlayingFile, []byte(text), 0644); err != nil {
					fmt.Printf("Failed to update now playing file: %v\n", err)
				}

				h.state.Mu.Lock()
				h.state.CurrentSong = song.Name
				h.state.Mu.Unlock()

				time.Sleep(song.Duration)
			}
		}
	}
}

func prepareAudioList(dir string) ([]models.Song, error) {
	var songs []models.Song
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
			absPath, _ := filepath.Abs(filepath.Join(dir, entry.Name()))
			duration, err := getDuration(absPath)
			if err != nil {
				utils.Dump("Error getting duration for file:", entry.Name(), "Error:", err.Error())
				continue
			}
			cleanName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			songs = append(songs, models.Song{Path: absPath, Name: cleanName, Duration: duration})
		}
	}
	return songs, nil
}

func buildConcatFile(songs []models.Song) (string, error) {
	content := ""
	for _, song := range songs {
		content += fmt.Sprintf("file '%s'\n", song.Path)
	}
	tmpFile, err := os.CreateTemp("", "ffmpeg_audio_*.txt")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString(content)
	return tmpFile.Name(), err
}

func getDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	secondsStr := strings.TrimSpace(string(out))
	seconds, err := strconv.ParseFloat(secondsStr, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds * float64(time.Second)), nil
}
