package stream

import (
	"context"
	"sync"
	"time"
)

const nowPlayingFile = "now_playing.txt"

type StreamState struct {
	mu          sync.Mutex
	isRunning   bool
	currentSong string
	cancel      context.CancelFunc
}

type Status struct {
	IsRunning   bool   `json:"isRunning"`
	CurrentSong string `json:"currentSong"`
}

type StartRequest struct {
	StreamKey string `json:"streamKey"`
	VideoPath string `json:"videoPath"`
	AudioDir  string `json:"audioDir"`
	FontPath  string `json:"fontPath"`
	TextX     string `json:"textX"`
	TextY     string `json:"textY"`
}

type Song struct {
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
}
