package settings

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const filePath = "settings.yaml"

const DefaultProfileID = "default"

type StreamProfile struct {
	ID              string `yaml:"id"                json:"id"`
	Name            string `yaml:"name"              json:"name"`
	AudioDir        string `yaml:"audio_dir"         json:"audio_dir"`
	PlaylistOrder   string `yaml:"playlist_order"    json:"playlist_order"`
	StreamEndMode   string `yaml:"stream_end_mode"   json:"stream_end_mode"`
	EndAfterMinutes string `yaml:"end_after_minutes" json:"end_after_minutes"`
	VideoPath       string `yaml:"video_path"        json:"video_path"`
	FontPath        string `yaml:"font_path"         json:"font_path"`
	TextX           string `yaml:"text_x"            json:"text_x"`
	TextY           string `yaml:"text_y"            json:"text_y"`
	NowPlayingLabel string `yaml:"now_playing_label" json:"now_playing_label"`
	NextSongLabel   string `yaml:"next_song_label"   json:"next_song_label"`
}

// DashboardSettings holds all user-configurable dashboard fields.
// It is persisted to settings.yaml in the working directory.
type DashboardSettings struct {
	Saved           bool            `yaml:"saved"             json:"saved"`
	SelectedProfile string          `yaml:"selected_profile"  json:"selected_profile"`
	Profiles        []StreamProfile `yaml:"profiles" json:"profiles"`
	StreamKey       string          `yaml:"stream_key"        json:"stream_key"`
	VideoCodec      string          `yaml:"video_codec"       json:"video_codec"`
	VideoPreset     string          `yaml:"video_preset"      json:"video_preset"`
	VideoBitrate    string          `yaml:"video_bitrate"     json:"video_bitrate"`
	VideoMaxRate    string          `yaml:"video_maxrate"     json:"video_maxrate"`
	VideoBufSize    string          `yaml:"video_bufsize"     json:"video_bufsize"`
}

// Load reads settings.yaml. Returns zero-value settings (Saved=false) if the
// file does not exist yet, so callers can detect "no saved settings".
func Load() (*DashboardSettings, error) {
	var s DashboardSettings
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		normalizeProfiles(&s)
		return &s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	normalizeProfiles(&s)
	return &s, nil
}

// Save writes s to settings.yaml, setting the Saved flag so future loads can
// distinguish "explicitly saved" from "never saved".
func Save(s *DashboardSettings) error {
	s.Saved = true
	normalizeProfiles(s)
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func GetActiveProfile(s *DashboardSettings) (*StreamProfile, error) {
	if s == nil {
		return nil, errors.New("settings is nil")
	}
	for _, p := range s.Profiles {
		if p.ID == s.SelectedProfile {
			return &p, nil
		}
	}

	// return default profile if selected not found
	for _, p := range s.Profiles {
		if p.ID == DefaultProfileID {
			return &p, nil
		}
	}

	return nil, errors.New("active profile not found")
}

func normalizeProfiles(s *DashboardSettings) {
	if s == nil {
		return
	}

	if len(s.Profiles) == 0 {
		s.Profiles = []StreamProfile{buildDefaultProfile()}
	} else {
		normalized := make([]StreamProfile, 0, len(s.Profiles)+1)
		seen := map[string]bool{}
		hasDefault := false

		for _, p := range s.Profiles {
			id := strings.TrimSpace(p.ID)
			if id == "" {
				continue
			}
			if seen[id] {
				continue
			}
			seen[id] = true

			name := strings.TrimSpace(p.Name)
			if id == DefaultProfileID {
				hasDefault = true
				if name == "" {
					name = "Default"
				}
			}
			if name == "" {
				name = "Profile"
			}

			normalized = append(normalized, StreamProfile{
				ID:              id,
				Name:            name,
				AudioDir:        p.AudioDir,
				PlaylistOrder:   p.PlaylistOrder,
				StreamEndMode:   p.StreamEndMode,
				EndAfterMinutes: p.EndAfterMinutes,
				VideoPath:       p.VideoPath,
				FontPath:        p.FontPath,
				TextX:           p.TextX,
				TextY:           p.TextY,
				NowPlayingLabel: p.NowPlayingLabel,
				NextSongLabel:   p.NextSongLabel,
			})
		}

		if !hasDefault {
			normalized = append([]StreamProfile{buildDefaultProfile()}, normalized...)
		}

		s.Profiles = normalized
	}

	if strings.TrimSpace(s.SelectedProfile) == "" {
		s.SelectedProfile = DefaultProfileID
	}

	selectedFound := false
	for _, p := range s.Profiles {
		if p.ID == s.SelectedProfile {
			selectedFound = true
			break
		}
	}
	if !selectedFound {
		s.SelectedProfile = DefaultProfileID
	}
}

func buildDefaultProfile() StreamProfile {
	return StreamProfile{
		ID:              DefaultProfileID,
		Name:            "Default",
		AudioDir:        "",
		PlaylistOrder:   "normal",
		StreamEndMode:   "forever",
		EndAfterMinutes: "60",
		VideoPath:       "",
		FontPath:        "",
		TextX:           "30",
		TextY:           "h-th-30",
		NowPlayingLabel: "Now Playing:",
		NextSongLabel:   "Next song:",
	}
}
