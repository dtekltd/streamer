package stream

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"streamer/config"
)

var ErrAlreadyRunning = errors.New("stream is already running")
var ErrStreamNotRunning = errors.New("stream is not running")
var ErrNoSongsFound = errors.New("no mp3 files found in directory")

type Service struct {
	state *StreamState
	cfg   config.AppConfig
}

func NewService(cfg config.AppConfig) *Service {
	return &Service{state: &StreamState{}, cfg: cfg}
}

func (s *Service) Start(req StartRequest) error {
	songs, err := prepareAudioList(req.AudioDir)
	if err != nil {
		return err
	}
	if len(songs) == 0 {
		return ErrNoSongsFound
	}

	s.state.mu.Lock()
	if s.state.isRunning {
		s.state.mu.Unlock()
		return ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.state.isRunning = true
	s.state.audioDir = req.AudioDir
	s.state.songs = songs
	s.state.currentSong = ""
	s.state.nextSong = ""
	s.state.nowPlayingLabel = normalizeNowPlayingLabel(req.NowPlayingLabel)
	s.state.nextSongLabel = normalizeNextSongLabel(req.NextSongLabel)
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

	go s.runStream(ctx, req.StreamKey, req.VideoPath, req.FontPath, textX, textY)
	return nil
}

func (s *Service) UpdatePlaylist() (int, error) {
	s.state.mu.Lock()
	if !s.state.isRunning {
		s.state.mu.Unlock()
		return 0, ErrStreamNotRunning
	}
	audioDir := s.state.audioDir
	s.state.mu.Unlock()

	songs, err := prepareAudioList(audioDir)
	if err != nil {
		return 0, err
	}
	if len(songs) == 0 {
		return 0, ErrNoSongsFound
	}

	s.state.mu.Lock()
	if !s.state.isRunning {
		s.state.mu.Unlock()
		return 0, ErrStreamNotRunning
	}
	s.state.songs = songs
	s.state.mu.Unlock()

	return len(songs), nil
}

func (s *Service) Stop() {
	s.state.mu.Lock()
	if s.state.isRunning && s.state.cancel != nil {
		s.logln("Stopping stream via web interface...")
		s.state.cancel()
	}
	s.state.mu.Unlock()
}

func (s *Service) Status() Status {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	songs := make([]string, 0, len(s.state.songs))
	for _, song := range s.state.songs {
		songs = append(songs, song.Name)
	}

	return Status{
		IsRunning:   s.state.isRunning,
		CurrentSong: s.state.currentSong,
		Songs:       songs,
	}
}

func (s *Service) StreamAudio(w *bufio.Writer) error {
	for {
		s.state.mu.Lock()
		songs := append([]Song(nil), s.state.songs...)
		nowPlayingLabel := s.state.nowPlayingLabel
		nextSongLabel := s.state.nextSongLabel
		isRunning := s.state.isRunning
		s.state.mu.Unlock()

		if !isRunning {
			return nil
		}

		if len(songs) == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for idx, song := range songs {
			nextSongName := songs[(idx+1)%len(songs)].Name

			s.state.mu.Lock()
			if !s.state.isRunning {
				s.state.mu.Unlock()
				return nil
			}
			s.state.currentSong = song.Name
			s.state.nextSong = nextSongName
			s.state.mu.Unlock()

			// Start the stopwatch!
			start := time.Now()

			text := formatOverlayText(nowPlayingLabel, song.Name, nextSongLabel, nextSongName)
			if err := os.WriteFile(nowPlayingFile, []byte(text), 0644); err != nil {
				s.logf("Failed to update now playing file: %v\n", err)
			}

			file, err := os.Open(song.Path)
			if err != nil {
				s.logf("Failed to open song %s: %v\n", song.Path, err)
				continue
			}

			skipID3Tags(file)
			_, copyErr := io.Copy(w, file)
			file.Close()
			if copyErr != nil {
				return copyErr
			}

			// Look at the stopwatch. Wait out the remaining length of the song,
			// minus a 3-second padding to keep FFmpeg's buffer fed!
			elapsed := time.Since(start)
			bufferPadding := 3 * time.Second

			if elapsed+bufferPadding < song.Duration {
				time.Sleep(song.Duration - elapsed - bufferPadding)
			}

			if err := w.Flush(); err != nil {
				return err
			}
		}
	}
}

func (s *Service) runStream(ctx context.Context, streamKey, videoFile, fontFile, textX, textY string) {
	defer func() {
		s.state.mu.Lock()
		s.state.isRunning = false
		s.state.currentSong = ""
		s.state.nextSong = ""
		s.state.nowPlayingLabel = ""
		s.state.nextSongLabel = ""
		s.state.songs = nil
		s.state.audioDir = ""
		s.state.cancel = nil
		s.state.mu.Unlock()
		os.Remove(nowPlayingFile)
		s.logln("Stream fully stopped and cleaned up.")
	}()

	nowPlayingLabel, nextSongLabel := s.getOverlayLabels()
	initialText := formatOverlayText(nowPlayingLabel, "Starting...", nextSongLabel, "")

	// Ensure drawtext can read a non-empty file before FFmpeg starts.
	if err := os.WriteFile(nowPlayingFile, []byte(initialText), 0644); err != nil {
		s.logf("Failed to initialize now playing file: %v\n", err)
	}

	rtmpURL := fmt.Sprintf(s.cfg.StreamURLTemplate, streamKey)
	internalAudioURL := fmt.Sprintf("http://127.0.0.1:%s/internal/audio", s.cfg.ServerPort)
	safeFontPath := filepath.ToSlash(fontFile)                  // Convert Windows backslashes to forward slashes
	safeFontPath = strings.Replace(safeFontPath, ":", "\\:", 1) // Escape the Windows drive colon (e.g., turns "E:/" into "E\:/")
	drawtextFilter := fmt.Sprintf("drawtext=fontfile='%s':textfile='%s':reload=1:fontcolor=white:fontsize=40:box=1:boxcolor=black@0.6:boxborderw=10:x=%s:y=%s", safeFontPath, nowPlayingFile, textX, textY)

	args := []string{
		"-re",
		"-stream_loop", "-1",
		"-i", videoFile,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", internalAudioURL,
		"-filter_complex", fmt.Sprintf("[0:v]%s[v]", drawtextFilter),
		"-map", "[v]",
		"-map", "1:a",
		"-c:v", "libx264",
		// "-preset", "veryfast",
		// // "-b:v", "3000k",
		// // "-maxrate", "3000k",
		// // "-bufsize", "6000k",
		// "-b:v", "13500k", // Set target bitrate to 13.5 Mbps
		// "-maxrate", "13500k", // Cap the maximum bitrate at 13.5 Mbps
		// "-bufsize", "27000k", // Set buffer size to double the maxrate (Standard practice)
		"-preset", "ultrafast", // <-- Changed to ultrafast to save CPU power
		"-b:v", "6000k", // <-- Lowered to 6 Mbps (Standard 1080p)
		"-maxrate", "6000k",
		"-bufsize", "12000k",
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

	s.logf("FFmpeg output target: %s\n", rtmpURL)
	s.logf("FFmpeg internal audio source: %s\n", internalAudioURL)

	if err := cmd.Run(); err != nil {
		s.logf("FFmpeg exited: %v\n", err)
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
			cleanName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			// Remove prefixes like "01." or "01 - "
			cleanName = strings.TrimSpace(strings.TrimLeftFunc(cleanName, func(r rune) bool {
				return (r >= '0' && r <= '9') || r == '.' || r == '-' || r == ' '
			}))
			// Calculate the exact duration of the MP3
			duration, err := getDuration(absPath)
			if err != nil {
				continue
			}
			songs = append(songs, Song{Path: absPath, Name: cleanName, Duration: duration})
		}
	}

	// Do not sort songs alphabetically to preserve the order in which they are read from the directory
	// sort.Slice(songs, func(i, j int) bool {
	// 	return strings.ToLower(songs[i].Name) < strings.ToLower(songs[j].Name)
	// })

	return songs, nil
}

// getDuration uses ffprobe to get the exact duration of an audio file
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

// skipID3Tags checks if an MP3 file has metadata at the beginning and fast-forwards past it.
func skipID3Tags(file *os.File) {
	header := make([]byte, 10)
	_, err := file.Read(header)
	if err != nil {
		file.Seek(0, io.SeekStart)
		return
	}

	// ID3v2 tags always start with the letters "ID3"
	if string(header[:3]) == "ID3" {
		// ID3 uses a special 32-bit "syncsafe" integer to declare its size.
		// We calculate the exact size of the metadata block here:
		size := (int(header[6]) << 21) | (int(header[7]) << 14) | (int(header[8]) << 7) | int(header[9])

		// Fast-forward the file exactly past the metadata (size + 10 bytes for the header itself)
		file.Seek(int64(size+10), io.SeekStart)
	} else {
		// No ID3 tag found, rewind back to the very beginning of the audio
		file.Seek(0, io.SeekStart)
	}
}

func (s *Service) getOverlayLabels() (string, string) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	return normalizeNowPlayingLabel(s.state.nowPlayingLabel), normalizeNextSongLabel(s.state.nextSongLabel)
}

func normalizeNowPlayingLabel(label string) string {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return "Now Playing:"
	}
	return trimmed
}

func formatNowPlayingText(label, song string) string {
	cleanLabel := normalizeNowPlayingLabel(label)
	cleanSong := strings.TrimSpace(song)
	return fmt.Sprintf("%s %s", cleanLabel, cleanSong)
}

func normalizeNextSongLabel(label string) string {
	return strings.TrimSpace(label)
}

func formatOverlayText(nowLabel, currentSong, nextLabel, nextSong string) string {
	nowText := formatNowPlayingText(nowLabel, currentSong)
	if normalizeNextSongLabel(nextLabel) == "" {
		return nowText
	}

	cleanNextSong := strings.TrimSpace(nextSong)
	if cleanNextSong == "" {
		return nowText
	}

	return fmt.Sprintf("%s | %s %s", nowText, normalizeNextSongLabel(nextLabel), cleanNextSong)
}

func (s *Service) logf(format string, args ...any) {
	if s.cfg.EnableLogging {
		fmt.Printf(format, args...)
	}
}

func (s *Service) logln(msg string) {
	if s.cfg.EnableLogging {
		fmt.Println(msg)
	}
}
