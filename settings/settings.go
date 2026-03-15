package settings

import (
	"os"

	"gopkg.in/yaml.v3"
)

const filePath = "settings.yaml"

// DashboardSettings holds all user-configurable dashboard fields.
// It is persisted to settings.yaml in the working directory.
type DashboardSettings struct {
	Saved           bool   `yaml:"saved"             json:"saved"`
	StreamKey       string `yaml:"stream_key"        json:"stream_key"`
	VideoPath       string `yaml:"video_path"        json:"video_path"`
	AudioDir        string `yaml:"audio_dir"         json:"audio_dir"`
	PlaylistOrder   string `yaml:"playlist_order"    json:"playlist_order"`
	ShufflePlaylist bool   `yaml:"shuffle_playlist,omitempty"  json:"shuffle_playlist,omitempty"`
	FontPath        string `yaml:"font_path"         json:"font_path"`
	TextX           string `yaml:"text_x"            json:"text_x"`
	TextY           string `yaml:"text_y"            json:"text_y"`
	NowPlayingLabel string `yaml:"now_playing_label" json:"now_playing_label"`
	NextSongLabel   string `yaml:"next_song_label"   json:"next_song_label"`
}

// Load reads settings.yaml. Returns zero-value settings (Saved=false) if the
// file does not exist yet, so callers can detect "no saved settings".
func Load() (DashboardSettings, error) {
	var s DashboardSettings
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := yaml.Unmarshal(data, &s); err != nil {
		return s, err
	}
	return s, nil
}

// Save writes s to settings.yaml, setting the Saved flag so future loads can
// distinguish "explicitly saved" from "never saved".
func Save(s DashboardSettings) error {
	s.Saved = true
	data, err := yaml.Marshal(&s)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
