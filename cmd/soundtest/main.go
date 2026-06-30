package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/ui"
)

func main() {
	flag.Parse()
	if flag.NArg() > 0 && flag.Arg(0) == "chord" {
		fmt.Println("Playing chord test (5 sec)...")
		playChord()
		return
	}
	p := ui.NewSoundPlayer()
	defer p.Close()
	fmt.Println("Testing PowerMode sounds — listen for tones")
	fmt.Println()
	fmt.Println("1. Warm (C5 523Hz)...")
	p.PlayLevel(ui.LevelWarm, ui.LevelNormal)
	time.Sleep(400 * time.Millisecond)
	fmt.Println("2. Hot (C5→E5 arpeggio)...")
	p.PlayLevel(ui.LevelHot, ui.LevelWarm)
	time.Sleep(500 * time.Millisecond)
	fmt.Println("3. Power (C5→E5→G5→C6 arpeggio)...")
	p.PlayLevel(ui.LevelPower, ui.LevelHot)
	time.Sleep(600 * time.Millisecond)
	fmt.Println("4. Combo blips (ascending)...")
	for i := 1; i <= 8; i++ {
		p.PlayComboUp(i)
		time.Sleep(150 * time.Millisecond)
	}
	fmt.Println("Done")
}

func playChord() {
	p := ui.NewSoundPlayer()
	defer p.Close()
	for i := 0; i < 10; i++ {
		p.PlayLevel(ui.LevelPower, ui.LevelNormal)
		time.Sleep(500 * time.Millisecond)
	}
}
