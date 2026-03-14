package html

import (
	stdhtml "html"
	"strings"
)

const Dashboard = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stream Controller</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #eef1f5; color: #333; margin: 0; padding: 24px; }
        .layout { max-width: 1200px; margin: 0 auto; display: grid; grid-template-columns: 360px 1fr; gap: 20px; }
        .panel { background: white; border-radius: 12px; box-shadow: 0 6px 20px rgba(0,0,0,0.08); border: 1px solid #e8ebef; }
        .sidebar { padding: 22px; }
        .content { padding: 24px; }
        h1 { margin: 0 0 14px; color: #222; font-size: 28px; }
        .subtitle { margin: 0 0 10px; color: #5e6672; font-size: 14px; }
        label { font-weight: bold; display: block; margin-top: 13px; margin-bottom: 5px; }
        input[type="text"] { width: 100%; padding: 10px; border: 1px solid #cfd6df; border-radius: 6px; box-sizing: border-box; }
        .checkbox-row { display: flex; align-items: center; gap: 8px; margin-top: 14px; }
        .checkbox-row label { margin: 0; font-weight: 600; cursor: pointer; }
        .btn { padding: 12px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 15px; font-weight: bold; color: white; transition: 0.2s; }
        .btn:disabled { opacity: 0.7; cursor: not-allowed; }
        .btn-row { margin-top: 16px; display: flex; gap: 10px; flex-wrap: wrap; }
        .btn-start { background-color: #28a745; min-width: 160px; }
        .btn-start:hover { background-color: #218838; }
        .btn-update { background-color: #007bff; display: none; }
        .btn-update:hover { background-color: #0069d9; }
        .btn-pause-songs { background-color: #fd7e14; display: none; }
        .btn-pause-songs:hover { background-color: #e86c0b; }
        .btn-restart-songs { background-color: #20c997; display: none; }
        .btn-restart-songs:hover { background-color: #1bab86; }
        .btn-stop { background-color: #dc3545; display: none; }
        .btn-stop:hover { background-color: #c82333; }
        .status-box { margin-top: 20px; padding: 20px; border-radius: 8px; background-color: #edf1f6; }
        .status-box h3 { margin: 0 0 8px; }
        .status-box p { margin: 0; }
        .indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; background-color: gray; margin-right: 8px; }
        .running .indicator { background-color: #28a745; box-shadow: 0 0 8px #28a745; }
        .playlist-box { margin-top: 20px; padding: 16px; border-radius: 8px; background: #f8f9fa; border: 1px solid #e0e3e7; min-height: 140px; }
        .playlist-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
        .playlist-box h4 { margin: 0; color: #222; }
        .badge { background: #1f6feb; color: #fff; border-radius: 999px; padding: 4px 10px; font-size: 12px; font-weight: 700; }
        .playlist-list { margin: 0; padding-left: 25px; max-height: 400px; overflow: auto; }
        .playlist-list li { margin: 5px 0; color: #444; }
        .playlist-list li.playing { color: #28a745; font-weight: bold; }
        .saved-indicator { font-size: 12px; color: #28a745; opacity: 0; transition: opacity 0.4s; margin-left: 8px; vertical-align: middle; }
        .sidebar-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 8px; }
        .sidebar-header h1 { flex: 1; }
        .sidebar-toggle { background: none; border: 1px solid #cfd6df; border-radius: 6px; cursor: pointer; padding: 6px 9px; color: #5e6672; font-size: 16px; line-height: 1; flex-shrink: 0; margin-top: 4px; transition: background 0.15s, color 0.15s; }
        .sidebar-toggle:hover { background: #eef1f5; color: #222; }
        .sidebar-toggle .chevron { display: inline-block; transition: transform 0.25s; }
        .sidebar.collapsed .chevron { transform: rotate(-90deg); }
        .sidebar-body { overflow: hidden; transition: max-height 0.3s ease, opacity 0.25s ease; max-height: 2000px; opacity: 1; }
        .sidebar.collapsed .sidebar-body { max-height: 0; opacity: 0; }

        @media (max-width: 900px) {
            body { padding: 14px; }
            .layout { grid-template-columns: 1fr; gap: 14px; }
            .sidebar, .content { padding: 18px; }
            h1 { font-size: 24px; }
            .btn-row { flex-direction: column; }
            .btn-start, .btn-update, .btn-pause-songs, .btn-restart-songs, .btn-stop { width: 100%; }
        }
    </style>
</head>
<body>

<div class="layout">
    <aside class="panel sidebar" id="sidebar">
        <div class="sidebar-header">
            <h1>🔴 Stream Controller</h1>
            <button class="sidebar-toggle" id="sidebarToggle" onclick="toggleSidebar()" title="Toggle settings panel">
                <span class="chevron">&#9660;</span>
            </button>
        </div>
        <p class="subtitle">Settings panel. <span id="settingsSavedIndicator" class="saved-indicator">✓ Saved</span></p>

        <div class="sidebar-body" id="sidebarBody">
        <div id="inputsPanel">
        <label for="streamKey">YouTube Stream Key</label>
        <input type="text" id="streamKey" value="__DEFAULT_STREAM_KEY__" placeholder="xxxx-xxxx-xxxx-xxxx">

        <label for="videoPath">Background Video Path</label>
        <input type="text" id="videoPath" value="__DEFAULT_VIDEO_PATH__" placeholder="C:\\videos\\loop.mp4 or ./loop.mp4">

        <label for="audioDir">Audio Directory</label>
        <input type="text" id="audioDir" value="__DEFAULT_AUDIO_DIR__" placeholder="C:\\music or ./music">

        <div class="checkbox-row">
            <input type="checkbox" id="shufflePlaylist">
            <label for="shufflePlaylist">Shuffle Playlist</label>
        </div>

        <label for="fontPath">Font File Path</label>
        <input type="text" id="fontPath" value="__DEFAULT_FONT_PATH__" placeholder="font.ttf">

        <label for="textX">Now Playing Text X</label>
        <input type="text" id="textX" value="30" placeholder="30 or w-tw-30">

        <label for="textY">Now Playing Text Y</label>
        <input type="text" id="textY" value="h-th-30" placeholder="h-th-30 or 50">

        <label for="nowPlayingLabel">Now Playing Label</label>
        <input type="text" id="nowPlayingLabel" value="Now Playing:" placeholder="Now Playing:">

        <label for="nextSongLabel">Next Song Label</label>
        <input type="text" id="nextSongLabel" value="Next song:" placeholder="Leave empty to hide">
        </div>
        </div><!-- /sidebar-body -->
    </aside>

    <main class="panel content">
        <h2 style="margin:0 0 6px;">Streaming Console</h2>
        <p class="subtitle">Start and monitor stream from here.</p>

    <div class="btn-row">
        <button class="btn btn-start" id="startBtn" onclick="startStream()">Start Stream</button>
        <button class="btn btn-update" id="updateBtn" onclick="updatePlaylist()">Scan Folder & Update Playlist</button>
        <button class="btn btn-pause-songs" id="pauseSongsBtn" onclick="stopSongs()">Stop Songs</button>
        <button class="btn btn-restart-songs" id="restartSongsBtn" onclick="restartSongs()">Restart Songs</button>
        <button class="btn btn-stop" id="stopBtn" onclick="stopStream()">Stop Stream</button>
    </div>

    <div class="status-box" id="statusBox">
        <h3><span class="indicator" id="indicator"></span> <span id="statusText">Offline</span></h3>
        <p><strong>Now Playing:</strong> <span id="nowPlaying">-</span></p>
    </div>

    <div class="playlist-box" id="playlistBox" style="display:none;">
        <div class="playlist-header">
            <h4>Playlist</h4>
            <span class="badge" id="songCountBadge">Total songs: 0</span>
        </div>
        <ol class="playlist-list" id="songList"></ol>
    </div>
</main>
</div>

<script>
    let saveSettingsTimer = null;

    document.addEventListener('DOMContentLoaded', async function() {
        initSidebar();
        await loadSettings();
        document.getElementById('inputsPanel').querySelectorAll('input').forEach(function(el) {
            el.addEventListener('input', scheduleSettingsSave);
            el.addEventListener('change', scheduleSettingsSave);
        });
    });

    function initSidebar() {
        // On mobile default to collapsed; on desktop default to expanded.
        const isMobile = window.innerWidth <= 900;
        const stored = localStorage.getItem('sidebarCollapsed');
        const shouldCollapse = stored !== null ? stored === 'true' : isMobile;
        if (shouldCollapse) {
            document.getElementById('sidebar').classList.add('collapsed');
        }
    }

    function toggleSidebar() {
        const sidebar = document.getElementById('sidebar');
        const collapsed = sidebar.classList.toggle('collapsed');
        localStorage.setItem('sidebarCollapsed', collapsed);
    }

    async function loadSettings() {
        try {
            const res = await fetch('/api/settings');
            if (!res.ok) return;
            const s = await res.json();
            if (!s.saved) return; // no settings saved yet — keep HTML defaults
            document.getElementById('streamKey').value = s.stream_key || '';
            document.getElementById('videoPath').value = s.video_path || '';
            document.getElementById('audioDir').value  = s.audio_dir  || '';
            document.getElementById('shufflePlaylist').checked = !!s.shuffle_playlist;
            document.getElementById('fontPath').value  = s.font_path  || '';
            document.getElementById('textX').value     = s.text_x     || '30';
            document.getElementById('textY').value     = s.text_y     || 'h-th-30';
            document.getElementById('nowPlayingLabel').value = s.now_playing_label !== undefined ? s.now_playing_label : 'Now Playing:';
            document.getElementById('nextSongLabel').value   = s.next_song_label   !== undefined ? s.next_song_label   : 'Next song:';
        } catch (e) {
            console.error('Failed to load settings', e);
        }
    }

    function scheduleSettingsSave() {
        clearTimeout(saveSettingsTimer);
        saveSettingsTimer = setTimeout(saveSettings, 800);
    }

    async function saveSettings() {
        const payload = {
            stream_key:        document.getElementById('streamKey').value,
            video_path:        document.getElementById('videoPath').value,
            audio_dir:         document.getElementById('audioDir').value,
            shuffle_playlist:  document.getElementById('shufflePlaylist').checked,
            font_path:         document.getElementById('fontPath').value,
            text_x:            document.getElementById('textX').value,
            text_y:            document.getElementById('textY').value,
            now_playing_label: document.getElementById('nowPlayingLabel').value,
            next_song_label:   document.getElementById('nextSongLabel').value
        };
        try {
            const res = await fetch('/api/settings', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (res.ok) showSavedIndicator();
        } catch (e) {
            console.error('Failed to save settings', e);
        }
    }

    function showSavedIndicator() {
        const ind = document.getElementById('settingsSavedIndicator');
        if (!ind) return;
        ind.style.opacity = '1';
        clearTimeout(ind._hideTimer);
        ind._hideTimer = setTimeout(function() { ind.style.opacity = '0'; }, 1500);
    }

    setInterval(fetchStatus, 2000);
    fetchStatus();

    async function fetchStatus() {
        try {
            const res = await fetch('/api/status');
            const data = await res.json();

            const statusBox = document.getElementById('statusBox');
            const statusText = document.getElementById('statusText');
            const nowPlaying = document.getElementById('nowPlaying');
            const stopBtn = document.getElementById('stopBtn');
            const updateBtn = document.getElementById('updateBtn');
            const pauseSongsBtn = document.getElementById('pauseSongsBtn');
            const restartSongsBtn = document.getElementById('restartSongsBtn');
            const startBtn = document.getElementById('startBtn');
            const inputsPanel = document.getElementById('inputsPanel');
            const playlistBox = document.getElementById('playlistBox');

            renderSongList(data.songs || [], data.currentSong || "");

            if (data.isRunning) {
                statusBox.classList.add('running');
                statusText.innerText = "Live / Streaming";
                nowPlaying.innerText = data.currentSong || "Loading...";
                startBtn.style.display = "none";
                updateBtn.style.display = "block";
                stopBtn.style.display = "block";
                if (data.songsPaused) {
                    pauseSongsBtn.style.display = "none";
                    restartSongsBtn.style.display = "block";
                } else {
                    pauseSongsBtn.style.display = "block";
                    restartSongsBtn.style.display = "none";
                }
                playlistBox.style.display = "block";
                setInputsDisabled(inputsPanel, true);
            } else {
                statusBox.classList.remove('running');
                statusText.innerText = "Offline";
                nowPlaying.innerText = "-";
                startBtn.style.display = "inline-block";
                updateBtn.style.display = "none";
                stopBtn.style.display = "none";
                pauseSongsBtn.style.display = "none";
                restartSongsBtn.style.display = "none";
                playlistBox.style.display = "none";
                setInputsDisabled(inputsPanel, false);
            }
        } catch (e) {
            console.error("Failed to fetch status", e);
        }
    }

    function setInputsDisabled(container, disabled) {
        const fields = container.querySelectorAll('input');
        for (const field of fields) {
            field.disabled = disabled;
        }
    }

    function renderSongList(songs, currentSong) {
        const songList = document.getElementById('songList');
        const songCountBadge = document.getElementById('songCountBadge');
        songList.innerHTML = "";
        songCountBadge.innerText = "Total songs: " + songs.length;

        if (!songs.length) {
            const item = document.createElement('li');
            item.innerText = "No songs loaded";
            songList.appendChild(item);
            return;
        }

        for (const song of songs) {
            const item = document.createElement('li');
            item.innerText = song;
            if (song === currentSong) {
                item.classList.add('playing');
            }
            songList.appendChild(item);
        }
    }

    async function startStream() {
        const payload = {
            streamKey: document.getElementById('streamKey').value,
            videoPath: document.getElementById('videoPath').value,
            audioDir: document.getElementById('audioDir').value,
            shufflePlaylist: document.getElementById('shufflePlaylist').checked,
            fontPath: document.getElementById('fontPath').value,
            textX: document.getElementById('textX').value,
            textY: document.getElementById('textY').value,
            nowPlayingLabel: document.getElementById('nowPlayingLabel').value,
            nextSongLabel: document.getElementById('nextSongLabel').value
        };

        if (!payload.streamKey || !payload.videoPath || !payload.audioDir || !payload.fontPath) {
            alert("Please fill in all fields.");
            return;
        }

        document.getElementById('startBtn').innerText = "Starting...";

        await fetch('/api/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        setTimeout(fetchStatus, 500);
        document.getElementById('startBtn').innerText = "Start Stream";
    }

    async function stopStream() {
        if (confirm("Are you sure you want to stop the stream?")) {
            await fetch('/api/stop', { method: 'POST' });
            setTimeout(fetchStatus, 500);
        }
    }

    async function stopSongs() {
        const btn = document.getElementById('pauseSongsBtn');
        const originalText = btn.innerText;
        btn.disabled = true;
        btn.innerText = "Pausing...";

        try {
            const res = await fetch('/api/stop-songs', { method: 'POST' });
            if (!res.ok) {
                const message = await res.text();
                throw new Error(message || 'Failed to stop songs');
            }
            await fetchStatus();
        } catch (e) {
            alert('Failed to stop songs: ' + e.message);
        } finally {
            btn.disabled = false;
            btn.innerText = originalText;
        }
    }

    async function restartSongs() {
        const btn = document.getElementById('restartSongsBtn');
        const originalText = btn.innerText;
        btn.disabled = true;
        btn.innerText = "Restarting...";

        try {
            const res = await fetch('/api/restart-songs', { method: 'POST' });
            if (!res.ok) {
                const message = await res.text();
                throw new Error(message || 'Failed to restart songs');
            }
            await fetchStatus();
        } catch (e) {
            alert('Failed to restart songs: ' + e.message);
        } finally {
            btn.disabled = false;
            btn.innerText = originalText;
        }
    }

    async function updatePlaylist() {
        const btn = document.getElementById('updateBtn');
        const originalText = btn.innerText;
        btn.disabled = true;
        btn.innerText = "Scanning...";

        try {
            const res = await fetch('/api/update-playlist', { method: 'POST' });
            if (!res.ok) {
                const message = await res.text();
                throw new Error(message || "Failed to update playlist");
            }

            const data = await res.json();
            alert("Playlist updated. " + data.songs + " songs loaded.");
            await fetchStatus();
        } catch (e) {
            alert("Failed to update playlist: " + e.message);
        } finally {
            btn.disabled = false;
            btn.innerText = originalText;
        }
    }
</script>

</body>
</html>
`

func RenderDashboard(serverMode string) string {
	isDev := strings.EqualFold(strings.TrimSpace(serverMode), "dev")

	defaultStreamKey := ""
	defaultVideoPath := ""
	defaultAudioDir := ""
	defaultFontPath := ""

	if isDev {
		defaultStreamKey = "test"
		defaultVideoPath = `E:\Live-Stream\test\video\TOP20-Video.mp4`
		defaultAudioDir = `E:\Live-Stream\test\audio`
		defaultFontPath = `E:\Live-Stream\resources\TiltNeon-Regular-VariableFont.ttf`
	}

	replacer := strings.NewReplacer(
		"__DEFAULT_STREAM_KEY__", stdhtml.EscapeString(defaultStreamKey),
		"__DEFAULT_VIDEO_PATH__", stdhtml.EscapeString(defaultVideoPath),
		"__DEFAULT_AUDIO_DIR__", stdhtml.EscapeString(defaultAudioDir),
		"__DEFAULT_FONT_PATH__", stdhtml.EscapeString(defaultFontPath),
	)

	return replacer.Replace(Dashboard)
}
