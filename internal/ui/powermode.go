package ui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type PowerLevel int

const (
	LevelNormal PowerLevel = iota
	LevelWarm
	LevelHot
	LevelPower
)

const (
	comboTimeout  = 3 * time.Second
	comboWindow   = 2 * time.Second
	energyDecay   = 0.96
	maxParticles  = 25
	pulseDuration = 6
)

var particleChars = []string{"✦", "✧", "⚡", "★", "♦", "‧", "*", "·"}

type Particle struct {
	Char    string
	X, Y    float64
	VX, VY  float64
	Life    int
	MaxLife int
}

type PowerMode struct {
	active      bool
	combo       int
	maxCombo    int
	energy      float64
	level       PowerLevel
	lastHit     time.Time
	particles   []Particle
	pulseFrames int
	sound       *SoundPlayer
}

func NewPowerMode() *PowerMode {
	return &PowerMode{
		active:    true,
		particles: make([]Particle, 0, maxParticles),
		sound:     NewSoundPlayer(),
	}
}

func (pm *PowerMode) Toggle() {
	pm.active = !pm.active
	if !pm.active {
		pm.reset()
	}
}

func (pm *PowerMode) IsActive() bool {
	return pm.active
}

func (pm *PowerMode) Level() PowerLevel {
	return pm.level
}

func (pm *PowerMode) Combo() int {
	return pm.combo
}

func (pm *PowerMode) MaxCombo() int {
	return pm.maxCombo
}

func (pm *PowerMode) Energy() float64 {
	return pm.energy
}

func (pm *PowerMode) PulseFrames() int {
	return pm.pulseFrames
}

func (pm *PowerMode) reset() {
	pm.combo = 0
	pm.energy = 0
	pm.level = LevelNormal
	pm.particles = pm.particles[:0]
	pm.pulseFrames = 0
}

func (pm *PowerMode) Punch() {
	if !pm.active {
		return
	}

	now := time.Now()
	if pm.lastHit.IsZero() {
		pm.lastHit = now
		pm.combo = 1
		return
	}

	elapsed := now.Sub(pm.lastHit)
	pm.lastHit = now

	if elapsed <= comboWindow {
		pm.combo++
	} else {
		pm.combo = 1
	}

	if pm.combo > pm.maxCombo {
		pm.maxCombo = pm.combo
	}

	pm.energy = math.Min(1.0, pm.energy+float64(pm.combo)/50.0)

	prevLevel := pm.level
	pm.updateLevel()

	pm.sound.PlayLevel(pm.level, prevLevel)
	if pm.combo > 1 && pm.level == prevLevel {
		pm.sound.PlayComboUp(pm.combo)
	}

	if pm.level > prevLevel && pm.level >= LevelHot {
		pm.pulseFrames = pulseDuration
	}

	pm.spawnParticles()
}

func (pm *PowerMode) Tick() {
	if !pm.active {
		return
	}

	if pm.combo > 0 && time.Since(pm.lastHit) > comboTimeout {
		pm.combo--
		if pm.combo < 0 {
			pm.combo = 0
		}
		pm.energy *= energyDecay
		if pm.energy < 0.01 {
			pm.energy = 0
		}
		pm.updateLevel()
	}

	pm.tickParticles()

	if pm.pulseFrames > 0 {
		pm.pulseFrames--
	}
}

func (pm *PowerMode) updateLevel() {
	switch {
	case pm.combo >= 11:
		pm.level = LevelPower
	case pm.combo >= 6:
		pm.level = LevelHot
	case pm.combo >= 3:
		pm.level = LevelWarm
	default:
		pm.level = LevelNormal
	}
}

func (pm *PowerMode) spawnParticles() {
	n := 0
	switch pm.level {
	case LevelWarm:
		n = 2
	case LevelHot:
		n = 6
	case LevelPower:
		n = 12
	default:
		return
	}

	for i := 0; i < n && len(pm.particles) < maxParticles; i++ {
		pm.particles = append(pm.particles, Particle{
			Char:    particleChars[rand.Intn(len(particleChars))],
			X:       0.5,
			Y:       0.5,
			VX:      (rand.Float64() - 0.5) * 0.08,
			VY:      -(rand.Float64() * 0.06),
			Life:    10 + rand.Intn(15),
			MaxLife: 25,
		})
	}
}

func (pm *PowerMode) tickParticles() {
	alive := pm.particles[:0]
	for _, p := range pm.particles {
		p.Life--
		if p.Life <= 0 {
			continue
		}
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.002
		alive = append(alive, p)
	}
	pm.particles = alive
}

func (pm *PowerMode) BorderColor() lipgloss.Color {
	switch pm.level {
	case LevelPower:
		return lipgloss.Color("196")
	case LevelHot:
		return lipgloss.Color("208")
	case LevelWarm:
		return lipgloss.Color("214")
	default:
		return lipgloss.Color("62")
	}
}

func (pm *PowerMode) RenderCombo() string {
	if !pm.active || pm.combo < 3 {
		return ""
	}

	color := comboNormal
	switch pm.level {
	case LevelWarm:
		color = comboWarm
	case LevelHot:
		color = comboHot
	case LevelPower:
		color = comboPower
	}

	bolt := "⚡"
	if pm.level == LevelPower {
		bolt = "🔥"
	}

	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	return style.Render(fmt.Sprintf("%s x%d", bolt, pm.combo))
}

func (pm *PowerMode) RenderEnergyBar(width int) string {
	if !pm.active || pm.combo < 3 {
		return ""
	}

	filled := int(pm.energy * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Builder{}
	for i := 0; i < width; i++ {
		if i < filled {
			if pm.level >= LevelHot {
				bar.WriteString("█")
			} else {
				bar.WriteString("▓")
			}
		} else {
			bar.WriteString("░")
		}
	}

	color := comboNormal
	switch pm.level {
	case LevelWarm:
		color = comboWarm
	case LevelHot:
		color = comboHot
	case LevelPower:
		color = comboPower
	}

	return lipgloss.NewStyle().Foreground(color).Render(bar.String())
}

func (pm *PowerMode) RenderParticles(width, height int) string {
	if !pm.active || len(pm.particles) == 0 {
		return ""
	}

	grid := make([][]string, height)
	for y := range grid {
		grid[y] = make([]string, width)
		for x := range grid[y] {
			grid[y][x] = " "
		}
	}

	for _, p := range pm.particles {
		lifetime := p.Life
		if lifetime < 0 {
			lifetime = 0
		}
		alpha := float64(lifetime) / float64(p.MaxLife)
		x := int(p.X * float64(width-1))
		y := int(p.Y * float64(height-1))
		if x >= 0 && x < width && y >= 0 && y < height {
			if alpha > 0.5 {
				grid[y][x] = p.Char
			} else {
				grid[y][x] = "·"
			}
		}
	}

	var b strings.Builder
	for _, row := range grid {
		b.WriteString(strings.Join(row, ""))
	}
	return b.String()
}
