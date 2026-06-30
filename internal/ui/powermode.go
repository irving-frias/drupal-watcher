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
	comboTimeout   = 3 * time.Second
	comboWindow    = 2 * time.Second
	energyDecay    = 0.96
	maxParticles   = 45
	pulseDuration  = 6
)

var sparkChars = []string{"✦", "✧", "⚡", "★", "♦"}
var fireChars  = []string{"🔥", "💥", "⚡"}
var smokeChars = []string{"·", "‧", "∘", "°"}

type ParticleType int

const (
	ParticleSpark ParticleType = iota
	ParticleFire
	ParticleSmoke
)

type Particle struct {
	Char    string
	X, Y    float64
	VX, VY  float64
	Life    int
	MaxLife int
	Typ     ParticleType
}

type PowerMode struct {
	active        bool
	combo         int
	maxCombo      int
	energy        float64
	level         PowerLevel
	lastHit       time.Time
	particles     []Particle
	pulseFrames   int
	overheatGlow  int
}

func NewPowerMode() *PowerMode {
	return &PowerMode{
		active:    true,
		particles: make([]Particle, 0, maxParticles),
	}
}

func (pm *PowerMode) Toggle() {
	pm.active = !pm.active
	if !pm.active {
		pm.reset()
	}
}

func (pm *PowerMode) IsActive() bool    { return pm.active }
func (pm *PowerMode) Level() PowerLevel { return pm.level }
func (pm *PowerMode) Combo() int        { return pm.combo }
func (pm *PowerMode) MaxCombo() int     { return pm.maxCombo }
func (pm *PowerMode) Energy() float64   { return pm.energy }
func (pm *PowerMode) PulseFrames() int  { return pm.pulseFrames }
func (pm *PowerMode) OverheatGlow() int { return pm.overheatGlow }

func (pm *PowerMode) reset() {
	pm.combo = 0
	pm.energy = 0
	pm.level = LevelNormal
	pm.particles = pm.particles[:0]
	pm.pulseFrames = 0
	pm.overheatGlow = 0
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

	if pm.level > prevLevel && pm.level >= LevelHot {
		pm.pulseFrames = pulseDuration
		pm.overheatGlow = 8
		pm.explosion()
	}

	if pm.level == prevLevel && pm.combo > 1 {
		pm.spawnParticles()
	}
}

func (pm *PowerMode) explosion() {
	n := 15
	if pm.level == LevelPower {
		n = 25
	}
	for i := 0; i < n && len(pm.particles) < maxParticles; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := 0.03 + rand.Float64()*0.07
		life := 15 + rand.Intn(15)
		char := sparkChars[rand.Intn(len(sparkChars))]
		if rand.Intn(3) == 0 {
			char = fireChars[rand.Intn(len(fireChars))]
		}
		pm.particles = append(pm.particles, Particle{
			Char:    char,
			X:       0.5,
			Y:       0.5,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    life,
			MaxLife: life,
			Typ:     ParticleSpark,
		})
	}
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
	if pm.overheatGlow > 0 {
		pm.overheatGlow--
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
	smokeN := 0
	switch pm.level {
	case LevelWarm:
		n = 2
		smokeN = 1
	case LevelHot:
		n = 5
		smokeN = 2
	case LevelPower:
		n = 10
		smokeN = 4
	default:
		return
	}

	for i := 0; i < n && len(pm.particles) < maxParticles; i++ {
		char := sparkChars[rand.Intn(len(sparkChars))]
		if rand.Intn(4) == 0 {
			char = fireChars[rand.Intn(len(fireChars))]
		}
		pm.particles = append(pm.particles, Particle{
			Char:    char,
			X:       0.3 + rand.Float64()*0.4,
			Y:       0.5 + rand.Float64()*0.3,
			VX:      (rand.Float64() - 0.5) * 0.12,
			VY:      -(0.04 + rand.Float64()*0.07),
			Life:    8 + rand.Intn(12),
			MaxLife: 20,
			Typ:     ParticleSpark,
		})
	}

	for i := 0; i < smokeN && len(pm.particles) < maxParticles; i++ {
		pm.particles = append(pm.particles, Particle{
			Char:    smokeChars[rand.Intn(len(smokeChars))],
			X:       0.3 + rand.Float64()*0.4,
			Y:       0.4 + rand.Float64()*0.2,
			VX:      (rand.Float64() - 0.5) * 0.02,
			VY:      -(0.01 + rand.Float64()*0.02),
			Life:    20 + rand.Intn(15),
			MaxLife: 35,
			Typ:     ParticleSmoke,
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

		switch p.Typ {
		case ParticleSpark:
			p.VY += 0.003
			p.VX *= 0.98
		case ParticleSmoke:
			p.VX += (rand.Float64() - 0.5) * 0.004
			p.VY -= 0.001
		}

		p.X += p.VX
		p.Y += p.VY
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

	prefix := "⚡"
	if pm.level == LevelHot {
		prefix = "🔥"
	}
	if pm.level == LevelPower {
		prefix = "💥"
	}

	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	return style.Render(fmt.Sprintf("%s x%d", prefix, pm.combo))
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
			switch {
			case pm.level >= LevelPower && pm.overheatGlow%2 == 0:
				bar.WriteString("█")
			case pm.level >= LevelHot:
				bar.WriteString("▓")
			default:
				bar.WriteString("▓")
			}
		} else {
			if pm.level >= LevelPower && pm.energy > 0.5 {
				bar.WriteString("▒")
			} else {
				bar.WriteString("░")
			}
		}
	}

	color := comboNormal
	switch pm.level {
	case LevelWarm:
		color = comboWarm
	case LevelHot:
		color = comboHot
	case LevelPower:
		if pm.overheatGlow > 0 && pm.overheatGlow%2 == 0 {
			color = lipgloss.Color("226")
		} else {
			color = comboPower
		}
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
		if x < 0 {
			x = 0
		}
		if x >= width {
			x = width - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}

		if alpha > 0.6 {
			grid[y][x] = p.Char
		} else if alpha > 0.3 {
			if p.Typ == ParticleSmoke {
				grid[y][x] = "·"
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
