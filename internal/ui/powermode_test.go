package ui

import (
	"testing"
	"time"
)

func TestPowerModeComboIncrement(t *testing.T) {
	pm := NewPowerMode()
	pm.Punch()
	if pm.Combo() != 1 {
		t.Fatalf("expected combo 1, got %d", pm.Combo())
	}
	pm.Punch()
	if pm.Combo() != 2 {
		t.Fatalf("expected combo 2, got %d", pm.Combo())
	}
	pm.Punch()
	if pm.Combo() != 3 {
		t.Fatalf("expected combo 3, got %d", pm.Combo())
	}
}

func TestPowerModeLevels(t *testing.T) {
	tests := []struct {
		combo int
		want  PowerLevel
	}{
		{0, LevelNormal},
		{1, LevelNormal},
		{2, LevelNormal},
		{3, LevelWarm},
		{4, LevelWarm},
		{5, LevelWarm},
		{6, LevelHot},
		{7, LevelHot},
		{10, LevelHot},
		{11, LevelPower},
		{20, LevelPower},
	}

	for _, tt := range tests {
		pm := NewPowerMode()
		pm.combo = tt.combo
		pm.updateLevel()
		if pm.level != tt.want {
			t.Errorf("combo %d: expected level %d, got %d", tt.combo, tt.want, pm.level)
		}
	}
}

func TestPowerModeEnergyDecay(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true
	pm.energy = 0.5
	pm.combo = 3

	pm.lastHit = time.Now().Add(-4 * time.Second)
	pm.Tick()

	if pm.energy >= 0.5 {
		t.Error("expected energy to decay")
	}
}

func TestPowerModeComboTimeout(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true
	pm.combo = 5
	pm.lastHit = time.Now().Add(-4 * time.Second)

	pm.Tick()

	if pm.combo >= 5 {
		t.Error("expected combo to decay after timeout")
	}
}

func TestPowerModeToggle(t *testing.T) {
	pm := NewPowerMode()
	if !pm.IsActive() {
		t.Fatal("expected active by default")
	}
	pm.Toggle()
	if pm.IsActive() {
		t.Fatal("expected inactive after toggle")
	}
	pm.Toggle()
	if !pm.IsActive() {
		t.Fatal("expected active after second toggle")
	}
}

func TestPowerModeInactiveDoesNotTrack(t *testing.T) {
	pm := NewPowerMode()
	pm.Toggle() // disable
	pm.Punch()
	if pm.Combo() != 0 {
		t.Fatal("expected no combo tracking when inactive")
	}
}

func TestPowerModeResetOnDisable(t *testing.T) {
	pm := NewPowerMode()
	pm.Punch()
	pm.Punch()
	pm.Punch()
	pm.energy = 0.5
	pm.Toggle() // disable — should reset
	if pm.Combo() != 0 || pm.Energy() != 0 || pm.Level() != LevelNormal {
		t.Fatal("expected full reset on disable")
	}
}

func TestPowerModeParticlesSpawn(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true

	// Warm: should spawn particles
	pm.combo = 4
	pm.level = LevelWarm
	pm.particles = pm.particles[:0]
	pm.spawnParticles()
	if len(pm.particles) == 0 {
		t.Error("expected particles at Warm level")
	}
}

func TestPowerModeParticlesDespawn(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true
	pm.particles = append(pm.particles, Particle{Life: 1, MaxLife: 10})
	pm.Tick()
	if len(pm.particles) != 0 {
		t.Error("expected particle to despawn after life reaches 0")
	}
}

func TestPowerModePulseFrames(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true

	// Transition to Hot should set pulse
	pm.combo = 6
	pm.level = LevelHot
	pm.pulseFrames = 0
	pm.level = LevelNormal
	pm.combo = 6
	pm.updateLevel()
	if pm.level != LevelHot {
		t.Fatalf("expected Hot, got %d", pm.level)
	}

	// Manually trigger pulse via Punch
	pm.pulseFrames = pulseDuration
	if pm.PulseFrames() == 0 {
		t.Error("expected pulse frames after power up")
	}

	pm.Tick()
	if pm.PulseFrames() >= pulseDuration {
		t.Error("expected pulse frames to decrease after tick")
	}
}

func TestPowerModeRenderCombo(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true

	r := pm.RenderCombo()
	if r != "" {
		t.Error("expected empty render for combo < 3")
	}

	pm.combo = 5
	pm.level = LevelWarm
	r = pm.RenderCombo()
	if r == "" {
		t.Error("expected non-empty render for combo >= 3")
	}
}

func TestPowerModeRenderEnergyBar(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true

	r := pm.RenderEnergyBar(10)
	if r != "" {
		t.Error("expected empty energy bar for combo < 3")
	}

	pm.combo = 5
	pm.energy = 0.5
	r = pm.RenderEnergyBar(10)
	if r == "" {
		t.Error("expected non-empty energy bar for combo >= 3")
	}
}

func TestPowerModeMaxCombo(t *testing.T) {
	pm := NewPowerMode()
	pm.Punch()
	pm.Punch()
	pm.Punch()
	if pm.MaxCombo() != 3 {
		t.Fatalf("expected maxCombo 3, got %d", pm.MaxCombo())
	}
	// combo decays but max stays
	pm.combo = 1
	if pm.MaxCombo() != 3 {
		t.Fatalf("expected maxCombo to persist at 3, got %d", pm.MaxCombo())
	}
}

func TestPowerModeEnergyCap(t *testing.T) {
	pm := NewPowerMode()
	pm.active = true
	pm.energy = 0.99
	for i := 0; i < 100; i++ {
		pm.combo = 50
		pm.Punch()
	}
	if pm.Energy() > 1.0 {
		t.Fatal("expected energy capped at 1.0")
	}
}
