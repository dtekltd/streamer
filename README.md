# youtube-streamer

Lightweight web UI and API for running one or more YouTube audio/video live streams from profile-based settings.

## Run

1. Install dependencies and ensure `ffmpeg` and `ffprobe` are available on PATH.
2. Start the app:

```bash
go run ./cmd/streamer
```

With runtime flags:

```bash
go run ./cmd/streamer --mode=dev --port=8080 --logs=true
```

3. Open the dashboard at `http://localhost:8080` (or your configured port).

Auto-start from saved settings:

```bash
go run ./cmd/streamer --start
```

Build binary:

```bash
go build ./cmd/streamer
```

## Multi-stream model

Streams are scoped by `profileId`.

- Each profile can run its own FFmpeg process concurrently.
- Overlay files (`now playing` and `next song`) are profile-specific to avoid collisions.
- Internal audio source is profile-scoped: `/internal/audio/:profileId`.

## API

### GET /api/status?profileId={id}

Returns selected profile status plus all currently running streams.

Response:

```json
{
  "current": {
    "profileId": "default",
    "isRunning": true,
    "currentSong": "Track Name",
    "songs": [
      {
        "name": "Track Name",
        "start": "00:00",
        "duration": "03:45",
        "display": "[00:00] Track Name"
      }
    ],
    "startedAt": "2026-03-21T10:00:00Z",
    "songIndex": 1,
    "songTotal": 10
  },
  "streams": [
    {
      "profileId": "default",
      "isRunning": true,
      "currentSong": "Track Name",
      "songs": [],
      "startedAt": "2026-03-21T10:00:00Z",
      "songIndex": 1,
      "songTotal": 10
    }
  ]
}
```

Notes:

- `current` is the status for requested `profileId` (or `default` when omitted).
- `streams` includes only active/running streams.

### POST /api/start

Start a stream for a profile.

Request body:

```json
{
  "profileId": "default",
  "streamKey": "xxxx-xxxx-xxxx-xxxx",
  "streamUrlTemplate": "rtmp://host:1935/live/%s",
  "videoPath": "C:\\videos\\loop.mp4",
  "audioDir": "C:\\music",
  "playlistOrder": "normal",
  "streamEndMode": "forever",
  "endAfterMinutes": "60",
  "fontPath": "C:\\fonts\\font.ttf",
  "videoCodec": "libx264",
  "videoPreset": "ultrafast",
  "videoBitrate": "6000k",
  "videoMaxRate": "6000k",
  "videoBufSize": "12000k",
  "textX": "30",
  "textY": "h-th-30",
  "nowPlayingLabel": "Now Playing:",
  "nextSongLabel": "Next song:"
}
```

### POST /api/stop

Stop a specific running stream.

Request body:

```json
{
  "profileId": "default"
}
```

### POST /api/update-playlist

Update and preview playlist settings for a profile.

Request body:

```json
{
  "profileId": "default",
  "playlistOrder": "shuffle",
  "audioDir": "C:\\music",
  "streamEndMode": "duration",
  "endAfterMinutes": "120"
}
```

Response:

```json
{
  "songs": 42,
  "playlist": [
    {
      "name": "Track Name",
      "start": "00:00",
      "duration": "03:45",
      "display": "[00:00] Track Name"
    }
  ]
}
```

### GET /internal/audio/:profileId

Internal profile-scoped MP3 stream consumed by FFmpeg input. This is not intended for browser playback.

### Settings endpoints

- `GET /api/settings`
- `POST /api/settings`

Profiles now support per-profile stream key via `stream_key`.

Example profile object:

```json
{
  "id": "default",
  "name": "Default",
  "stream_key": "xxxx-xxxx-xxxx-xxxx",
  "stream_url_template": "rtmp://host:1935/live/%s",
  "audio_dir": "C:\\music",
  "playlist_order": "normal",
  "stream_end_mode": "forever",
  "end_after_minutes": "60",
  "video_path": "C:\\videos\\loop.mp4",
  "font_path": "C:\\fonts\\font.ttf",
  "text_x": "30",
  "text_y": "h-th-30",
  "now_playing_label": "Now Playing:",
  "next_song_label": "Next song:"
}
```

## Notes

- `profileId` defaults to `default` when not provided.
- Legacy top-level `stream_key` remains supported as a fallback for auto-start.
- Legacy top-level `stream_url_template` remains supported as a fallback for auto-start.
