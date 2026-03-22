package stream

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"streamer/config"
)

var ErrAlreadyRunning = errors.New("stream is already running")
var ErrStreamNotRunning = errors.New("stream is not running")
var ErrNoSongsFound = errors.New("no mp3 files found in directory")
var ErrNoMediaInput = errors.New("background video path or audio directory must exist")

const DefaultProfileID = "default"

type Service struct {
	cfg     *config.AppConfig
	mu      sync.Mutex
	streams map[string]*StreamState
}

func NewService(cfg *config.AppConfig) *Service {
	return &Service{streams: map[string]*StreamState{}, cfg: cfg}
}

func (s *Service) Start(req StartRequest) error {
	profileID := normalizeProfileID(req.ProfileID)
	state := s.getOrCreateState(profileID)

	playlistOrder := normalizePlaylistOrder(req.PlaylistOrder)
	state.mu.Lock()
	songs := append([]Song(nil), state.songs...)
	existingOrder := state.playlistOrder
	existingAudioDir := state.audioDir
	alreadyRunning := state.isRunning
	state.mu.Unlock()

	if alreadyRunning {
		return ErrAlreadyRunning
	}

	hasVideoInput := fileExists(req.VideoPath)
	hasAudioInput := dirExists(req.AudioDir)
	if !hasVideoInput && !hasAudioInput {
		return ErrNoMediaInput
	}
	if !hasAudioInput {
		songs = nil
	}

	if hasAudioInput && (strings.TrimSpace(req.AudioDir) != strings.TrimSpace(existingAudioDir) || playlistOrder != existingOrder || len(songs) == 0) {
		s.logf("Update playlist for %s: %s (previous: %s, songs: %d)\n", profileID, playlistOrder, existingOrder, len(songs))
		var err error
		songs, err = prepareAudioList(req.AudioDir, playlistOrder)
		if err != nil {
			return err
		}
		if len(songs) == 0 {
			return ErrNoSongsFound
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	textX := strings.TrimSpace(req.TextX)
	if textX == "" {
		textX = "30"
	}

	textY := strings.TrimSpace(req.TextY)
	if textY == "" {
		textY = "h-th-30"
	}

	ffmpegArgs := defaultString(req.FFmpegArgs, defaultFFmpegArgs())
	videoAudioEnabled := hasVideoInput && req.EnableVideoAudio && videoHasAudio(req.VideoPath)
	videoAudioVolume := normalizeVideoAudioVolume(req.VideoAudioVolume)
	streamURLTemplate := defaultString(req.StreamURLTemplate, "rtmp://10.16.0.165:1935/live/%s")
	streamEndMode := normalizeStreamEndMode(req.StreamEndMode)
	endAfter := parseEndAfterMinutes(req.EndAfterMinutes)

	state.mu.Lock()
	state.isRunning = true
	state.playlistOrder = playlistOrder
	if hasAudioInput {
		state.audioDir = req.AudioDir
	} else {
		state.audioDir = ""
	}
	state.songs = songs
	state.currentSong = ""
	state.nextSong = ""
	state.nowPlayingLabel = normalizeNowPlayingLabel(req.NowPlayingLabel)
	state.nextSongLabel = normalizeNextSongLabel(req.NextSongLabel)
	state.streamEndMode = streamEndMode
	state.endAfter = endAfter
	state.startedAt = time.Now()
	state.cancel = cancel
	state.mu.Unlock()

	go s.runStream(ctx, profileID, req.StreamKey, streamURLTemplate, req.VideoPath, hasVideoInput, hasAudioInput, videoAudioEnabled, videoAudioVolume, req.FontPath, textX, textY, ffmpegArgs)
	return nil
}

func (s *Service) UpdatePlaylist(profileID string, orderOverride, audioDirOverride, streamEndModeOverride, endAfterMinutesOverride *string) ([]PlaylistItem, error) {
	profileID = normalizeProfileID(profileID)
	state := s.getOrCreateState(profileID)

	state.mu.Lock()
	if orderOverride != nil {
		state.playlistOrder = normalizePlaylistOrder(*orderOverride)
	}
	if audioDirOverride != nil {
		audioDir := strings.TrimSpace(*audioDirOverride)
		if audioDir != "" {
			state.audioDir = audioDir
		}
	}
	if streamEndModeOverride != nil {
		state.streamEndMode = normalizeStreamEndMode(*streamEndModeOverride)
	}
	if endAfterMinutesOverride != nil {
		state.endAfter = parseEndAfterMinutes(*endAfterMinutesOverride)
	}
	audioDir := state.audioDir
	playlistOrder := state.playlistOrder
	state.mu.Unlock()

	songs, err := prepareAudioList(audioDir, playlistOrder)
	if err != nil {
		return nil, err
	}
	if len(songs) == 0 {
		return nil, ErrNoSongsFound
	}

	state.mu.Lock()
	state.songs = songs
	state.mu.Unlock()

	playlist := make([]PlaylistItem, 0, len(songs))
	var startOffset time.Duration
	for _, song := range songs {
		startText := formatClock(startOffset)
		durationText := formatClock(song.Duration)
		playlist = append(playlist, PlaylistItem{
			Name:     song.Name,
			Start:    startText,
			Duration: durationText,
			Display:  fmt.Sprintf("[%s] %s", startText, song.Name),
		})
		startOffset += song.Duration
	}

	return playlist, nil
}

func (s *Service) Stop(profileID string) {
	state := s.getState(normalizeProfileID(profileID))
	if state == nil {
		return
	}

	state.mu.Lock()
	if state.isRunning && state.cancel != nil {
		state.cancel()
	}
	state.mu.Unlock()
}

func (s *Service) Status(profileID string) Status {
	profileID = normalizeProfileID(profileID)
	state := s.getState(profileID)
	if state == nil {
		return Status{ProfileID: profileID}
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	playlist := make([]PlaylistItem, 0, len(state.songs))
	var startOffset time.Duration
	songIndex := 0
	for i, song := range state.songs {
		startText := formatClock(startOffset)
		durationText := formatClock(song.Duration)
		playlist = append(playlist, PlaylistItem{
			Name:     song.Name,
			Start:    startText,
			Duration: durationText,
			Display:  fmt.Sprintf("[%s] %s", startText, song.Name),
		})
		if song.Name == state.currentSong {
			songIndex = i + 1
		}
		startOffset += song.Duration
	}

	return Status{
		ProfileID:   profileID,
		IsRunning:   state.isRunning,
		CurrentSong: state.currentSong,
		Songs:       playlist,
		StartedAt:   state.startedAt,
		SongIndex:   songIndex,
		SongTotal:   len(state.songs),
	}
}

func (s *Service) Statuses() []Status {
	s.mu.Lock()
	ids := make([]string, 0, len(s.streams))
	for profileID := range s.streams {
		ids = append(ids, profileID)
	}
	s.mu.Unlock()

	sort.Strings(ids)
	statuses := make([]Status, 0, len(ids))
	for _, profileID := range ids {
		status := s.Status(profileID)
		if status.IsRunning {
			statuses = append(statuses, status)
		}
	}

	return statuses
}

func (s *Service) StreamAudio(profileID string, w *bufio.Writer) error {
	profileID = normalizeProfileID(profileID)
	nowFile, nextFile := overlayFilePaths(profileID)

	for {
		state := s.getState(profileID)
		if state == nil {
			return nil
		}

		state.mu.Lock()
		songs := append([]Song(nil), state.songs...)
		isRunning := state.isRunning
		nowPlayingLabel := state.nowPlayingLabel
		nextSongLabel := state.nextSongLabel
		streamEndMode := state.streamEndMode
		endAfter := state.endAfter
		startedAt := state.startedAt
		state.mu.Unlock()

		if !isRunning {
			return nil
		}

		if streamEndMode == StreamEndDuration && endAfter > 0 && time.Since(startedAt) >= endAfter {
			s.logf("Configured stream duration reached for %s. Will stop the stream after the current song finishes.\n", profileID)
			s.Stop(profileID)
			return nil
		}

		if len(songs) == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for idx, song := range songs {
			if streamEndMode == StreamEndDuration && endAfter > 0 && time.Since(startedAt) >= endAfter {
				s.logf("Configured stream duration reached for %s. Will stop the stream after the current song finishes.\n", profileID)
				s.Stop(profileID)
				return nil
			}

			nextSongName := songs[(idx+1)%len(songs)].Name

			state.mu.Lock()
			if !state.isRunning {
				state.mu.Unlock()
				return nil
			}
			state.currentSong = song.Name
			state.nextSong = nextSongName
			state.mu.Unlock()

			start := time.Now()

			nowText := formatNowPlayingText(nowPlayingLabel, song.Name)
			if err := os.WriteFile(nowFile, []byte(nowText), 0644); err != nil {
				s.logf("Failed to update now playing file for %s: %v\n", profileID, err)
			}

			nextText := formatNextSongText(nextSongLabel, nextSongName)
			if err := os.WriteFile(nextFile, []byte(nextText), 0644); err != nil {
				s.logf("Failed to update next song file for %s: %v\n", profileID, err)
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

			elapsed := time.Since(start)
			if elapsed < song.Duration {
				remaining := song.Duration - elapsed
				time.Sleep(remaining)
			}

			if err := w.Flush(); err != nil {
				return err
			}
		}

		if streamEndMode == StreamEndAllSongs {
			s.logf("All songs have been played once for %s. Will stop the stream after the current song finishes.\n", profileID)
			s.Stop(profileID)
			return nil
		}
	}
}

func (s *Service) runStream(ctx context.Context, profileID, streamKey, streamURLTemplate, videoFile string, hasVideoInput, hasAudioInput, videoAudioEnabled bool, videoAudioVolume, fontFile, textX, textY, ffmpegArgsText string) {
	state := s.getOrCreateState(profileID)
	nowFile, nextFile := overlayFilePaths(profileID)

	defer func() {
		state.mu.Lock()
		state.isRunning = false
		state.currentSong = ""
		state.nextSong = ""
		state.streamEndMode = ""
		state.endAfter = 0
		state.startedAt = time.Time{}
		state.nowPlayingLabel = ""
		state.nextSongLabel = ""
		state.audioDir = ""
		state.cancel = nil
		state.mu.Unlock()
		os.Remove(nowFile)
		os.Remove(nextFile)
		s.logf("Stream %s fully stopped and cleaned up.\n", profileID)
	}()

	audioInputURL := fmt.Sprintf("http://127.0.0.1:%s/internal/audio/%s", s.cfg.ServerPort, url.PathEscape(profileID))
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-nostats",
	}
	filterParts := make([]string, 0, 4)
	videoMap := ""
	audioMap := ""
	if hasVideoInput {
		args = append(args, "-re", "-stream_loop", "-1", "-fflags", "+genpts", "-i", videoFile)
	}
	if hasAudioInput {
		args = append(args, "-i", audioInputURL)
	}

	if nowPlayingLabel, nextSongLabel := s.getOverlayLabels(state); hasVideoInput && nowPlayingLabel != "" && fileExists(fontFile) {
		nowInitialText := formatNowPlayingText(nowPlayingLabel, "Starting...")
		nextInitialText := formatNextSongText(nextSongLabel, "")

		if err := os.WriteFile(nowFile, []byte(nowInitialText), 0644); err != nil {
			s.logf("Failed to initialize now playing file for %s: %v\n", profileID, err)
		}
		if err := os.WriteFile(nextFile, []byte(nextInitialText), 0644); err != nil {
			s.logf("Failed to initialize next song file for %s: %v\n", profileID, err)
		}

		safeFontPath := filepath.ToSlash(fontFile)
		safeFontPath = strings.Replace(safeFontPath, ":", "\\:", 1)
		nowTextY := textY
		nextTextY := fmt.Sprintf("(%s)+55", textY)

		drawNowPlaying := fmt.Sprintf("drawtext=fontfile='%s':textfile='%s':reload=1:fontcolor=white:fontsize=40:box=1:boxcolor=black@0.6:boxborderw=10:x=%s:y=%s", safeFontPath, nowFile, textX, nowTextY)

		combinedFilter := ""
		if normalizeNextSongLabel(nextSongLabel) == "" {
			combinedFilter = fmt.Sprintf("[0:v]%s[v]", drawNowPlaying)
		} else {
			drawNextPlaying := fmt.Sprintf("drawtext=fontfile='%s':textfile='%s':reload=1:fontcolor=white@0.7:fontsize=30:box=1:boxcolor=black@0.6:boxborderw=10:x=%s:y=%s", safeFontPath, nextFile, textX, nextTextY)
			combinedFilter = fmt.Sprintf("[0:v]%s,%s[v]", drawNowPlaying, drawNextPlaying)
		}
		filterParts = append(filterParts, combinedFilter)
		videoMap = "[v]"
	} else if hasVideoInput {
		videoMap = "0:v"
	}

	if videoAudioEnabled {
		filterParts = append(filterParts, fmt.Sprintf("[0:a]volume=%s[videoaudio]", videoAudioVolume))
		if hasAudioInput {
			filterParts = append(filterParts, fmt.Sprintf("[%d:a][videoaudio]amix=inputs=2:duration=longest:dropout_transition=2[a]", audioInputIndex(hasVideoInput)))
			audioMap = "[a]"
		} else {
			audioMap = "[videoaudio]"
		}
	} else if hasAudioInput {
		audioMap = fmt.Sprintf("%d:a", audioInputIndex(hasVideoInput))
	}

	if len(filterParts) > 0 {
		args = append(args, "-filter_complex", strings.Join(filterParts, ";"))
	}
	if videoMap != "" {
		args = append(args, "-map", videoMap)
	}
	if audioMap != "" {
		args = append(args, "-map", audioMap)
	}

	args = append(args, filterManagedFFmpegArgs(parseFFmpegArgs(ffmpegArgsText), videoFile, audioInputURL, hasVideoInput, hasAudioInput)...)
	args = append(args, fmt.Sprintf(streamURLTemplate, streamKey))

	for {
		s.logf("Starting FFmpeg encoder for profile %s...\n", profileID)
		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if ctx.Err() != nil {
			s.logf("Stream %s has been safely stopped from the web interface.\n", profileID)
			break
		}

		s.logf("Warning for %s (%v)!\n", profileID, err)
		s.logln("Attempting to reconnect automatically in 5 seconds...")

		state.mu.Lock()
		streamEndMode := state.streamEndMode
		state.mu.Unlock()
		if streamEndMode == StreamEndForever {
			time.Sleep(5 * time.Second)
		} else {
			s.logln("Stream end mode is not set to 'forever', so will not attempt to restart the stream.")
			break
		}
	}
}

func prepareAudioList(dir string, playlistOrder string) ([]Song, error) {
	var songs []Song
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
			absPath, _ := filepath.Abs(filepath.Join(dir, entry.Name()))
			cleanName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			cleanName = strings.TrimSpace(strings.TrimLeftFunc(cleanName, func(r rune) bool {
				return (r >= '0' && r <= '9') || r == '.' || r == '-' || r == ' '
			}))
			duration, err := getDuration(absPath)
			if err != nil {
				continue
			}
			songs = append(songs, Song{Path: absPath, Name: cleanName, Duration: duration})
		}
	}

	sortSongsByOrder(songs, playlistOrder)
	return songs, nil
}

func sortSongsByOrder(songs []Song, playlistOrder string) {
	order := normalizePlaylistOrder(playlistOrder)

	switch order {
	case PlaylistOrderAZ:
		sort.Slice(songs, func(i, j int) bool {
			return strings.ToLower(songs[i].Name) < strings.ToLower(songs[j].Name)
		})
	case PlaylistOrderZA:
		sort.Slice(songs, func(i, j int) bool {
			return strings.ToLower(songs[i].Name) > strings.ToLower(songs[j].Name)
		})
	case PlaylistOrderShuffle:
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		rng.Shuffle(len(songs), func(i, j int) {
			songs[i], songs[j] = songs[j], songs[i]
		})
	case PlaylistOrderNormal:
	}
}

func normalizePlaylistOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case PlaylistOrderAZ:
		return PlaylistOrderAZ
	case PlaylistOrderZA:
		return PlaylistOrderZA
	case PlaylistOrderShuffle:
		return PlaylistOrderShuffle
	default:
		return PlaylistOrderNormal
	}
}

func formatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int64(d.Seconds())
	minutes := totalSec / 60
	seconds := totalSec % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
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

func skipID3Tags(file *os.File) {
	header := make([]byte, 10)
	_, err := file.Read(header)
	if err != nil {
		file.Seek(0, io.SeekStart)
		return
	}

	if string(header[:3]) == "ID3" {
		size := (int(header[6]) << 21) | (int(header[7]) << 14) | (int(header[8]) << 7) | int(header[9])
		file.Seek(int64(size+10), io.SeekStart)
	} else {
		file.Seek(0, io.SeekStart)
	}
}

func (s *Service) getOverlayLabels(state *StreamState) (string, string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	return normalizeNowPlayingLabel(state.nowPlayingLabel), normalizeNextSongLabel(state.nextSongLabel)
}

func normalizeNowPlayingLabel(label string) string {
	return strings.TrimSpace(label)
}

func formatNowPlayingText(label, song string) string {
	cleanLabel := normalizeNowPlayingLabel(label)
	cleanSong := strings.TrimSpace(song)
	return fmt.Sprintf("%s %s", cleanLabel, cleanSong)
}

func normalizeNextSongLabel(label string) string {
	return strings.TrimSpace(label)
}

func formatNextSongText(label, song string) string {
	cleanLabel := normalizeNextSongLabel(label)
	if cleanLabel == "" {
		return ""
	}

	cleanSong := strings.TrimSpace(song)
	if cleanSong == "" {
		return ""
	}

	return fmt.Sprintf("%s %s", cleanLabel, cleanSong)
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

func defaultString(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func defaultFFmpegArgs() string {
	return strings.Join([]string{
		"-c:v",
		"libx264",
		"-preset",
		"ultrafast",
		"-b:v",
		"6000k",
		"-maxrate",
		"6000k",
		"-bufsize",
		"12000k",
		"-pix_fmt",
		"yuv420p",
		"-g",
		"50",
		"-c:a",
		"aac",
		"-b:a",
		"128k",
		"-ar",
		"44100",
		"-f",
		"flv",
		"-flvflags",
		"no_duration_filesize",
		"-rtmp_buffer",
		"1000",
		"-rtmp_live",
		"live",
		"-reconnect",
		"1",
		"-reconnect_at_eof",
		"1",
		"-reconnect_streamed",
		"1",
		"-reconnect_delay_max",
		"5",
	}, "\n")
}

func parseFFmpegArgs(value string) []string {
	lines := strings.Split(value, "\n")
	args := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		args = append(args, trimmed)
	}
	return args
}

func filterManagedFFmpegArgs(args []string, videoFile, audioInputURL string, hasVideoInput, hasAudioInput bool) []string {
	filtered := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "-i" && i+1 < len(args) {
			next := args[i+1]
			if hasVideoInput && samePath(next, videoFile) {
				i++
				continue
			}
			if hasAudioInput && isManagedAudioInput(next, audioInputURL) {
				i++
				continue
			}
		}
		filtered = append(filtered, args[i])
	}
	return filtered
}

func isManagedAudioInput(value, audioInputURL string) bool {
	trimmed := strings.TrimSpace(value)
	candidates := []string{
		audioInputURL,
		strings.Replace(audioInputURL, "127.0.0.1", "localhost", 1),
		"http://127.0.0.1:8080/internal/audio",
		"http://localhost:8080/internal/audio",
		fmt.Sprintf("http://127.0.0.1:8080/internal/audio/%s", url.PathEscape(DefaultProfileID)),
		fmt.Sprintf("http://localhost:8080/internal/audio/%s", url.PathEscape(DefaultProfileID)),
	}
	for _, candidate := range candidates {
		if strings.EqualFold(trimmed, candidate) {
			return true
		}
	}
	return false
}

func samePath(left, right string) bool {
	if strings.TrimSpace(left) == "" || strings.TrimSpace(right) == "" {
		return false
	}
	leftClean := filepath.Clean(strings.TrimSpace(left))
	rightClean := filepath.Clean(strings.TrimSpace(right))
	return strings.EqualFold(leftClean, rightClean)
}

func fileExists(path string) bool {
	info, err := os.Stat(strings.TrimSpace(path))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(strings.TrimSpace(path))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func videoHasAudio(path string) bool {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=index", "-of", "csv=p=0", trimmed)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func audioInputIndex(hasVideoInput bool) int {
	if hasVideoInput {
		return 1
	}
	return 0
}

func normalizeStreamEndMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case StreamEndDuration:
		return StreamEndDuration
	case StreamEndAllSongs:
		return StreamEndAllSongs
	default:
		return StreamEndForever
	}
}

func parseEndAfterMinutes(value string) time.Duration {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 60 * time.Minute
	}

	minutes, err := strconv.Atoi(trimmed)
	if err != nil || minutes <= 0 {
		return 60 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}

func normalizeVideoAudioVolume(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "1.0"
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
		return "1.0"
	}
	return trimmed
}

func normalizeProfileID(profileID string) string {
	clean := strings.TrimSpace(profileID)
	if clean == "" {
		return DefaultProfileID
	}
	return clean
}

func overlayFilePaths(profileID string) (string, string) {
	safeID := sanitizeFileToken(profileID)
	return "now_playing_" + safeID + ".txt", "next_playing_" + safeID + ".txt"
}

func sanitizeFileToken(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultProfileID
	}

	var b strings.Builder
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	result := b.String()
	if result == "" {
		return DefaultProfileID
	}
	return result
}

func (s *Service) getOrCreateState(profileID string) *StreamState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if st, ok := s.streams[profileID]; ok {
		return st
	}

	st := &StreamState{
		playlistOrder: PlaylistOrderNormal,
		streamEndMode: StreamEndForever,
	}
	s.streams[profileID] = st
	return st
}

func (s *Service) getState(profileID string) *StreamState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.streams[profileID]
}
