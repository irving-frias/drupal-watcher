//go:build cgo

package ui

import (
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/generators"
	"github.com/gopxl/beep/v2/speaker"
)

type SoundPlayer struct {
	initOnce sync.Once
	failed   bool
	sr       beep.SampleRate
}

func NewSoundPlayer() *SoundPlayer {
	return &SoundPlayer{sr: beep.SampleRate(44100)}
}

func (sp *SoundPlayer) ensure() {
	sp.initOnce.Do(func() {
		err := speaker.Init(sp.sr, sp.sr.N(time.Second/10))
		if err != nil {
			sp.failed = true
		}
	})
}

func (sp *SoundPlayer) Tones(freqs []float64, noteDur time.Duration) {
	sp.ensure()
	if sp.failed {
		return
	}

	parts := make([]beep.Streamer, 0, len(freqs)*2)
	for _, f := range freqs {
		tone, err := generators.SineTone(sp.sr, f)
		if err != nil {
			return
		}
		parts = append(parts, beep.Take(sp.sr.N(noteDur), tone))
	}

	speaker.Play(beep.Seq(parts...))
}

func (sp *SoundPlayer) PlayLevel(level PowerLevel, prevLevel PowerLevel) {
	if level == prevLevel {
		return
	}

	switch level {
	case LevelWarm:
		sp.Tones([]float64{523.25}, 100*time.Millisecond)
	case LevelHot:
		sp.Tones([]float64{523.25, 659.25}, 80*time.Millisecond)
	case LevelPower:
		sp.Tones([]float64{523.25, 659.25, 783.99, 1046.50}, 60*time.Millisecond)
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
	sp.Tones([]float64{base}, 60*time.Millisecond)
}

func (sp *SoundPlayer) Close() {
	speaker.Clear()
	speaker.Close()
}
