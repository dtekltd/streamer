package stream

import (
	"context"
	"sync"
	"time"
)

const nowPlayingFile = "now_playing.txt"
const nextPlayingFile = "next_playing.txt"

const (
	PlaylistOrderNormal  = "normal"
	PlaylistOrderAZ      = "a-z"
	PlaylistOrderZA      = "z-a"
	PlaylistOrderShuffle = "shuffle"
)

type StreamState struct {
	mu              sync.Mutex
	isRunning       bool
	playlistOrder   string
	currentSong     string
	nextSong        string
	nowPlayingLabel string
	nextSongLabel   string
	songs           []Song
	audioDir        string
	cancel          context.CancelFunc
}

type PlaylistItem struct {
	Name     string `json:"name"`
	Start    string `json:"start"`
	Duration string `json:"duration"`
	Display  string `json:"display"`
}

type Status struct {
	IsRunning   bool           `json:"isRunning"`
	CurrentSong string         `json:"currentSong"`
	Songs       []PlaylistItem `json:"songs"`
}

type StartRequest struct {
	StreamKey       string `json:"streamKey"`
	VideoPath       string `json:"videoPath"`
	AudioDir        string `json:"audioDir"`
	PlaylistOrder   string `json:"playlistOrder"`
	FontPath        string `json:"fontPath"`
	VideoCodec      string `json:"videoCodec"`
	VideoPreset     string `json:"videoPreset"`
	VideoBitrate    string `json:"videoBitrate"`
	VideoMaxRate    string `json:"videoMaxRate"`
	VideoBufSize    string `json:"videoBufSize"`
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
