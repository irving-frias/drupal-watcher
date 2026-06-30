//go:build !cgo

package ui

type SoundPlayer struct{}

func NewSoundPlayer() *SoundPlayer {
	return &SoundPlayer{}
}

func (sp *SoundPlayer) PlayLevel(PowerLevel, PowerLevel) {}
func (sp *SoundPlayer) PlayComboUp(int)                 {}
func (sp *SoundPlayer) Close()                          {}
