package ui

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type SoundPlayer struct {
	tmpDir string
}

func NewSoundPlayer() *SoundPlayer {
	dir, err := os.MkdirTemp("", "dw-sound")
	if err != nil {
		dir = ""
	}
	return &SoundPlayer{tmpDir: dir}
}

func (sp *SoundPlayer) PlayLevel(level PowerLevel, prevLevel PowerLevel) {
	if level == prevLevel {
		return
	}
	switch level {
	case LevelWarm:
		sp.playTone(523.25, 100*time.Millisecond)
	case LevelHot:
		sp.playTone(523.25, 80*time.Millisecond)
		time.Sleep(20 * time.Millisecond)
		sp.playTone(659.25, 80*time.Millisecond)
	case LevelPower:
		sp.playTone(523.25, 60*time.Millisecond)
		time.Sleep(15 * time.Millisecond)
		sp.playTone(659.25, 60*time.Millisecond)
		time.Sleep(15 * time.Millisecond)
		sp.playTone(783.99, 60*time.Millisecond)
		time.Sleep(15 * time.Millisecond)
		sp.playTone(1046.50, 60*time.Millisecond)
	}
}

func (sp *SoundPlayer) PlayComboUp(combo int) {
	if combo <= 0 {
		return
	}
	base := 220.0 + float64(combo)*15.0
	if base > 880 {
		base = 880
	}
	sp.playTone(base, 60*time.Millisecond)
}

func (sp *SoundPlayer) Close() {
	if sp.tmpDir != "" {
		os.RemoveAll(sp.tmpDir)
	}
}

func (sp *SoundPlayer) playTone(freq float64, dur time.Duration) {
	data := generateWAV(freq, dur, 44100)
	sp.playWAV(data)
}

func generateWAV(freq float64, dur time.Duration, sampleRate int) []byte {
	numSamples := int(float64(sampleRate) * dur.Seconds())
	if numSamples <= 0 {
		numSamples = 1
	}

	pcm := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		value := math.Sin(2 * math.Pi * freq * t)

		envelope := 1.0
		fadeLen := numSamples / 10
		if i < fadeLen {
			envelope = float64(i) / float64(fadeLen)
		} else if i > numSamples-fadeLen {
			envelope = float64(numSamples-i) / float64(fadeLen)
		}

		pcm[i] = int16(value * envelope * 0.8 * 32767)
	}

	var buf bytes.Buffer
	dataSize := numSamples * 2
	fileSize := 36 + dataSize

	h := make([]byte, 44)
	copy(h[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(h[4:8], uint32(fileSize))
	copy(h[8:12], []byte("WAVE"))
	copy(h[12:16], []byte("fmt "))
	binary.LittleEndian.PutUint32(h[16:20], 16)
	binary.LittleEndian.PutUint16(h[20:22], 1)
	binary.LittleEndian.PutUint16(h[22:24], 1)
	binary.LittleEndian.PutUint32(h[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(h[28:32], uint32(sampleRate*2))
	binary.LittleEndian.PutUint16(h[32:34], 2)
	binary.LittleEndian.PutUint16(h[34:36], 16)
	copy(h[36:40], []byte("data"))
	binary.LittleEndian.PutUint32(h[40:44], uint32(dataSize))

	buf.Write(h)
	buf.Write(encodePCM16(pcm))
	return buf.Bytes()
}

func encodePCM16(samples []int16) []byte {
	b := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(b[i*2:], uint16(s))
	}
	return b
}

func (sp *SoundPlayer) playWAV(data []byte) {
	if sp.tmpDir == "" {
		return
	}

	switch runtime.GOOS {
	case "darwin":
		f, err := os.CreateTemp(sp.tmpDir, "*.wav")
		if err != nil {
			return
		}
		f.Write(data)
		f.Close()
		exec.Command("afplay", f.Name()).Start()

	case "linux":
		// Docker: PULSE_SERVER env var set by docker-compose
		pulseSrv := os.Getenv("PULSE_SERVER")
		if pulseSrv != "" && hasExec("paplay") {
			cmd := exec.Command("paplay", "--raw", "--rate=44100", "--channels=1", "--format=s16le")
			cmd.Stdin = bytes.NewReader(data[44:])
			cmd.Env = append(os.Environ(), "PULSE_SERVER="+pulseSrv)
			cmd.Start()
		} else if hasExec("paplay") {
			cmd := exec.Command("paplay", "--raw", "--rate=44100", "--channels=1", "--format=s16le")
			cmd.Stdin = bytes.NewReader(data[44:])
			cmd.Start()
		} else if hasExec("pw-play") {
			cmd := exec.Command("pw-play", "--rate=44100", "--channels=1", "--format=s16", "-")
			cmd.Stdin = bytes.NewReader(data[44:])
			cmd.Start()
		} else if hasExec("aplay") {
			cmd := exec.Command("aplay", "-r", "44100", "-c", "1", "-f", "S16_LE")
			cmd.Stdin = bytes.NewReader(data[44:])
			cmd.Start()
		} else if hasExec("powershell.exe") {
			exec.Command("powershell.exe", "-c",
				"[console]::beep(440,200)").Start()
		}

	case "windows":
		if hasExec("powershell.exe") {
			exec.Command("powershell.exe", "-c",
				"[console]::beep(440,200)").Start()
		}
	}
}

func hasExec(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
