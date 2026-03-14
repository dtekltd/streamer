package stream

import (
	"context"
	"sync"
	"time"
)

const nowPlayingFile = "now_playing.txt"
const nextPlayingFile = "next_playing.txt"

type StreamState struct {
	mu              sync.Mutex
	isRunning       bool
	songsPaused     bool
	shufflePlaylist bool
	currentSong     string
	nextSong        string
	nowPlayingLabel string
	nextSongLabel   string
	songs           []Song
	audioDir        string
	cancel          context.CancelFunc
}

type Status struct {
	IsRunning   bool     `json:"isRunning"`
	SongsPaused bool     `json:"songsPaused"`
	CurrentSong string   `json:"currentSong"`
	Songs       []string `json:"songs"`
}

type StartRequest struct {
	StreamKey       string `json:"streamKey"`
	VideoPath       string `json:"videoPath"`
	AudioDir        string `json:"audioDir"`
	ShufflePlaylist bool   `json:"shufflePlaylist"`
	FontPath        string `json:"fontPath"`
	TextX           string `json:"textX"`
	TextY           string `json:"textY"`
	NowPlayingLabel string `json:"nowPlayingLabel"`
	NextSongLabel   string `json:"nextSongLabel"`
}

type Song struct {
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
}
