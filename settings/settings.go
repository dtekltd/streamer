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
	ID                 string `yaml:"id"                json:"id"`
	Name               string `yaml:"name"              json:"name"`
	StreamKey          string `yaml:"stream_key"        json:"stream_key"`
	StreamURLTemplate  string `yaml:"stream_url_template" json:"stream_url_template"`
	AudioDir           string `yaml:"audio_dir"         json:"audio_dir"`
	EnableVideoAudio   bool   `yaml:"enable_video_audio" json:"enable_video_audio"`
	VideoAudioVolume   string `yaml:"video_audio_volume" json:"video_audio_volume"`
	FFmpegArgs         string `yaml:"ffmpeg_args"       json:"ffmpeg_args"`
	PlaylistOrder      string `yaml:"playlist_order"    json:"playlist_order"`
	StreamEndMode      string `yaml:"stream_end_mode"   json:"stream_end_mode"`
	EndAfterMinutes    string `yaml:"end_after_minutes" json:"end_after_minutes"`
	VideoPath          string `yaml:"video_path"        json:"video_path"`
	FontPath           string `yaml:"font_path"         json:"font_path"`
	TextX              string `yaml:"text_x"            json:"text_x"`
	TextY              string `yaml:"text_y"            json:"text_y"`
	EnablePlayingLabel bool   `yaml:"enable_playing_label" json:"enable_playing_label"`
	NowPlayingLabel    string `yaml:"now_playing_label" json:"now_playing_label"`
	EnableNextLabel    bool   `yaml:"enable_next_label" json:"enable_next_label"`
	NextSongLabel      string `yaml:"next_song_label"   json:"next_song_label"`
}

// DashboardSettings holds all user-configurable dashboard fields.
// It is persisted to settings.yaml in the working directory.
type DashboardSettings struct {
	Saved             bool            `yaml:"saved"             json:"saved"`
	SelectedProfile   string          `yaml:"selected_profile"  json:"selected_profile"`
	Profiles          []StreamProfile `yaml:"profiles" json:"profiles"`
	StreamKey         string          `yaml:"stream_key"        json:"stream_key"`
	StreamURLTemplate string          `yaml:"stream_url_template" json:"stream_url_template"`
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

	legacyTemplate := strings.TrimSpace(s.StreamURLTemplate)
	if legacyTemplate == "" {
		legacyTemplate = "rtmp://10.16.0.165:1935/live/%s"
	}

	if len(s.Profiles) == 0 {
		p := buildDefaultProfile()
		if strings.TrimSpace(s.StreamKey) != "" {
			p.StreamKey = s.StreamKey
		}
		p.StreamURLTemplate = legacyTemplate
		s.Profiles = []StreamProfile{p}
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
				ID:                 id,
				Name:               name,
				StreamKey:          p.StreamKey,
				StreamURLTemplate:  defaultString(p.StreamURLTemplate, legacyTemplate),
				AudioDir:           p.AudioDir,
				EnableVideoAudio:   p.EnableVideoAudio,
				VideoAudioVolume:   defaultString(p.VideoAudioVolume, "1.0"),
				FFmpegArgs:         defaultString(p.FFmpegArgs, defaultFFmpegArgs()),
				PlaylistOrder:      p.PlaylistOrder,
				StreamEndMode:      p.StreamEndMode,
				EndAfterMinutes:    p.EndAfterMinutes,
				VideoPath:          p.VideoPath,
				FontPath:           p.FontPath,
				TextX:              p.TextX,
				TextY:              p.TextY,
				EnablePlayingLabel: p.EnablePlayingLabel,
				NowPlayingLabel:    p.NowPlayingLabel,
				EnableNextLabel:    p.EnableNextLabel,
				NextSongLabel:      p.NextSongLabel,
			})
		}

		if !hasDefault {
			p := buildDefaultProfile()
			if strings.TrimSpace(s.StreamKey) != "" {
				p.StreamKey = s.StreamKey
			}
			p.StreamURLTemplate = legacyTemplate
			normalized = append([]StreamProfile{p}, normalized...)
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

	if strings.TrimSpace(s.StreamURLTemplate) == "" {
		s.StreamURLTemplate = legacyTemplate
	}
}

func buildDefaultProfile() StreamProfile {
	return StreamProfile{
		ID:                 DefaultProfileID,
		Name:               "Default",
		StreamKey:          "",
		StreamURLTemplate:  "rtmp://10.16.0.165:1935/live/%s",
		AudioDir:           "",
		EnableVideoAudio:   false,
		VideoAudioVolume:   "1.0",
		FFmpegArgs:         defaultFFmpegArgs(),
		PlaylistOrder:      "normal",
		StreamEndMode:      "forever",
		EndAfterMinutes:    "60",
		VideoPath:          "",
		FontPath:           "",
		TextX:              "30",
		TextY:              "h-th-30",
		EnablePlayingLabel: true,
		NowPlayingLabel:    "Now Playing:",
		EnableNextLabel:    true,
		NextSongLabel:      "Next song:",
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
