package stream

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"streamer/utils"
)

var ErrAlreadyRunning = errors.New("stream is already running")

type Service struct {
	state *StreamState
}

func NewService() *Service {
	return &Service{state: &StreamState{}}
}

func (s *Service) Start(req StartRequest) error {
	s.state.mu.Lock()
	if s.state.isRunning {
		s.state.mu.Unlock()
		return ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.state.isRunning = true
	s.state.cancel = cancel
	s.state.mu.Unlock()

	textX := strings.TrimSpace(req.TextX)
	if textX == "" {
		textX = "30"
	}

	textY := strings.TrimSpace(req.TextY)
	if textY == "" {
		textY = "h-th-30"
	}

	go s.runStream(ctx, req.StreamKey, req.VideoPath, req.AudioDir, req.FontPath, textX, textY)
	return nil
}

func (s *Service) Stop() {
	s.state.mu.Lock()
	if s.state.isRunning && s.state.cancel != nil {
		fmt.Println("Stopping stream via web interface...")
		s.state.cancel()
	}
	s.state.mu.Unlock()
}

func (s *Service) Status() Status {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	return Status{
		IsRunning:   s.state.isRunning,
		CurrentSong: s.state.currentSong,
	}
}

func (s *Service) runStream(ctx context.Context, streamKey, videoFile, audioDir, fontFile, textX, textY string) {
	defer func() {
		s.state.mu.Lock()
		s.state.isRunning = false
		s.state.currentSong = ""
		s.state.cancel = nil
		s.state.mu.Unlock()
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
	defer os.Remove(nowPlayingFile)

	go s.manageNowPlayingText(ctx, songs)

	// Switch to YouTube RTMP when ready:
	// rtmpURL := fmt.Sprintf("rtmp://a.rtmp.youtube.com/live2/%s", streamKey)
	rtmpURL := fmt.Sprintf("rtmp://10.16.0.165:1935/live/%s", streamKey)
	safeFontPath := filepath.ToSlash(fontFile)                  // Convert Windows backslashes to forward slashes
	safeFontPath = strings.Replace(safeFontPath, ":", "\\:", 1) // Escape the Windows drive colon (e.g., turns "E:/" into "E\:/")
	drawtextFilter := fmt.Sprintf("drawtext=fontfile='%s':textfile='%s':reload=1:fontcolor=white:fontsize=40:box=1:boxcolor=black@0.6:boxborderw=10:x=%s:y=%s", safeFontPath, nowPlayingFile, textX, textY)

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

func (s *Service) manageNowPlayingText(ctx context.Context, songs []Song) {
	for {
		for _, song := range songs {
			select {
			case <-ctx.Done():
				return
			default:
				text := fmt.Sprintf("Now Playing: %s", song.Name)
				if err := os.WriteFile(nowPlayingFile, []byte(text), 0644); err != nil {
					fmt.Printf("Failed to update now playing file: %v\n", err)
				}

				s.state.mu.Lock()
				s.state.currentSong = song.Name
				s.state.mu.Unlock()

				time.Sleep(song.Duration)
			}
		}
	}
}

func prepareAudioList(dir string) ([]Song, error) {
	var songs []Song
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
			// Remove prefixes like "01." or "01 - "
			cleanName = strings.TrimSpace(strings.TrimLeftFunc(cleanName, func(r rune) bool {
				return (r >= '0' && r <= '9') || r == '.' || r == '-' || r == ' '
			}))
			songs = append(songs, Song{Path: absPath, Name: cleanName, Duration: duration})
		}
	}
	return songs, nil
}

func buildConcatFile(songs []Song) (string, error) {
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
