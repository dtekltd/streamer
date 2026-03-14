package html

const Dashboard = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stream Controller</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #f4f4f9; color: #333; max-width: 600px; margin: 40px auto; padding: 20px; }
        .card { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 4px 8px rgba(0,0,0,0.1); }
        h1 { margin-top: 0; color: #222; }
        label { font-weight: bold; display: block; margin-top: 15px; margin-bottom: 5px; }
        input[type="text"] { width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 5px; box-sizing: border-box; }
        .btn { padding: 12px 20px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; font-weight: bold; width: 100%; margin-top: 20px; color: white; transition: 0.2s; }
        .btn-start { background-color: #28a745; }
        .btn-start:hover { background-color: #218838; }
        .btn-stop { background-color: #dc3545; display: none; }
        .btn-stop:hover { background-color: #c82333; }
        .status-box { margin-top: 30px; padding: 20px; border-radius: 8px; background-color: #e9ecef; text-align: center; }
        .indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; background-color: gray; margin-right: 8px; }
        .running .indicator { background-color: #28a745; box-shadow: 0 0 8px #28a745; }
    </style>
</head>
<body>

<div class="card">
    <h1>🔴 Stream Controller</h1>

    <div id="controls">
        <label for="streamKey">YouTube Stream Key</label>
        <input type="text" id="streamKey" value="test" placeholder="xxxx-xxxx-xxxx-xxxx">

        <label for="videoPath">Background Video Path</label>
        <input type="text" id="videoPath" value="E:\\Live-Stream\\test\\video\\TOP20-Video.mp4" placeholder="C:\\videos\\loop.mp4 or ./loop.mp4">

        <label for="audioDir">Audio Directory</label>
        <input type="text" id="audioDir" value="E:\\Live-Stream\\test\\audio" placeholder="C:\\music or ./music">

        <label for="fontPath">Font File Path</label>
        <input type="text" id="fontPath" value="E:\\Live-Stream\\resources\\TiltNeon-Regular-VariableFont.ttf" placeholder="font.ttf">

        <label for="textX">Now Playing Text X</label>
        <input type="text" id="textX" value="30" placeholder="30 or w-tw-30">

        <label for="textY">Now Playing Text Y</label>
        <input type="text" id="textY" value="h-th-30" placeholder="h-th-30 or 50">

        <button class="btn btn-start" id="startBtn" onclick="startStream()">Start Stream</button>
    </div>

    <button class="btn btn-stop" id="stopBtn" onclick="stopStream()">Stop Stream</button>

    <div class="status-box" id="statusBox">
        <h3><span class="indicator" id="indicator"></span> <span id="statusText">Offline</span></h3>
        <p><strong>Now Playing:</strong> <span id="nowPlaying">-</span></p>
    </div>
</div>

<script>
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
            const controls = document.getElementById('controls');

            if (data.isRunning) {
                statusBox.classList.add('running');
                statusText.innerText = "Live / Streaming";
                nowPlaying.innerText = data.currentSong || "Loading...";
                controls.style.display = "none";
                stopBtn.style.display = "block";
            } else {
                statusBox.classList.remove('running');
                statusText.innerText = "Offline";
                nowPlaying.innerText = "-";
                controls.style.display = "block";
                stopBtn.style.display = "none";
            }
        } catch (e) {
            console.error("Failed to fetch status", e);
        }
    }

    async function startStream() {
        const payload = {
            streamKey: document.getElementById('streamKey').value,
            videoPath: document.getElementById('videoPath').value,
            audioDir: document.getElementById('audioDir').value,
            fontPath: document.getElementById('fontPath').value,
            textX: document.getElementById('textX').value,
            textY: document.getElementById('textY').value
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
</script>

</body>
</html>
`
