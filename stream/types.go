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
	StreamEndForever     = "forever"
	StreamEndDuration    = "duration"
	StreamEndAllSongs    = "all_songs"
)

type StreamState struct {
	mu                 sync.Mutex
	isRunning          bool
	playlistOrder      string
	streamEndMode      string
	isEnding           bool
	endAfter           time.Duration
	startedAt          time.Time
	currentSong        string
	nextSong           string
	enablePlayingLabel bool
	nowPlayingLabel    string
	enableNextLabel    bool
	nextSongLabel      string
	songs              []Song
	audioDir           string
	cancel             context.CancelFunc
}

type PlaylistItem struct {
	Name     string `json:"name"`
	Start    string `json:"start"`
	Duration string `json:"duration"`
	Display  string `json:"display"`
}

type Status struct {
	ProfileID   string         `json:"profileId"`
	IsRunning   bool           `json:"isRunning"`
	CurrentSong string         `json:"currentSong"`
	Songs       []PlaylistItem `json:"songs"`
	StartedAt   time.Time      `json:"startedAt"`
	SongIndex   int            `json:"songIndex"`
	SongTotal   int            `json:"songTotal"`
}

type StartRequest struct {
	ProfileID          string `json:"profileId"`
	StreamKey          string `json:"streamKey"`
	StreamURLTemplate  string `json:"streamUrlTemplate"`
	VideoPath          string `json:"videoPath"`
	EnableVideoAudio   bool   `json:"enableVideoAudio"`
	VideoAudioVolume   string `json:"videoAudioVolume"`
	AudioDir           string `json:"audioDir"`
	FFmpegArgs         string `json:"ffmpegArgs"`
	PlaylistOrder      string `json:"playlistOrder"`
	StreamEndMode      string `json:"streamEndMode"`
	EndAfterMinutes    string `json:"endAfterMinutes"`
	FontPath           string `json:"fontPath"`
	TextX              string `json:"textX"`
	TextY              string `json:"textY"`
	EnablePlayingLabel bool   `json:"enablePlayingLabel"`
	NowPlayingLabel    string `json:"nowPlayingLabel"`
	EnableNextLabel    bool   `json:"enableNextLabel"`
	NextSongLabel      string `json:"nextSongLabel"`
}

type Song struct {
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
}
