package models

import (
	"context"
	"sync"
	"time"
)

const NowPlayingFile = "now_playing.txt"

type StreamState struct {
	Mu          sync.Mutex         `json:"-"`
	IsRunning   bool               `json:"isRunning"`
	CurrentSong string             `json:"currentSong"`
	Cancel      context.CancelFunc `json:"-"`
}

type StartStreamRequest struct {
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
