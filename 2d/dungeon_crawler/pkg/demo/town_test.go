package demo

import (
	"fmt"
	"math"
	"testing"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

func TestTownBoxCoversStampTiles(t *testing.T) {
	// The facade box must map a building's stamped tile rect through the
	// view offsets onto exactly the pixels those tiles occupy — the bug
	// that drew facades one offset up-and-left of their blocks.
	r := townRect{x0: 4, y0: 2, x1: 5, y1: 3}
	const ox, oy, tile = 285, 97, 32
	bx, by, bw, bh := townBox(r, ox, oy, tile)
	if bx != ox+4*tile || by != oy+2*tile {
		t.Errorf("origin (%v,%v), want (%v,%v)", bx, by, ox+4*tile, oy+2*tile)
	}
	if bw != 2*tile || bh != 2*tile {
		t.Errorf("extent (%v,%v), want (%v,%v)", bw, bh, 2*tile, 2*tile)
	}
}

func TestCondition01Ladder(t *testing.T) {
	if got := condition01(0, 3); got != 0 {
		t.Errorf("condition01(0,3) = %v, want 0 (derelict)", got)
	}
	if got := condition01(3, 3); got != 1 {
		t.Errorf("condition01(3,3) = %v, want 1 (kept)", got)
	}
	if got := condition01(1, 2); got != 0.5 {
		t.Errorf("condition01(1,2) = %v, want 0.5 (patched)", got)
	}
	if got := condition01(0, 0); got != 1 {
		t.Errorf("condition01(0,0) = %v, want 1 (maxed or unmetered)", got)
	}
	prev := condition01(0, 3)
	for lvl := 1; lvl <= 3; lvl++ {
		v := condition01(lvl, 3)
		if v <= prev {
			t.Errorf("condition01(%d,3) = %v, must climb above %v", lvl, v, prev)
		}
		prev = v
	}
}

func TestTile01DeterministicSpread(t *testing.T) {
	if tile01("Gate") != tile01("Gate") {
		t.Error("tile01 is not deterministic")
	}
	lo, hi := 1.0, 0.0
	for i := 0; i < 500; i++ {
		v := tile01(fmt.Sprintf("t%d", i))
		if v < 0 || v >= 1 {
			t.Fatalf("tile01 out of [0,1): %v", v)
		}
		lo = math.Min(lo, v)
		hi = math.Max(hi, v)
	}
	if hi-lo < 0.9 { // scatter details need range, not injectivity
		t.Errorf("tile01 never spread: [%v, %v]", lo, hi)
	}
}

func TestTownWearRisesWithUpgrades(t *testing.T) {
	m := crawler.DefaultMeta()
	low := townWear(m)
	if low < 0 || low > 1 {
		t.Fatalf("townWear out of [0,1]: %v", low)
	}
	for _, name := range []string{"Smithy", "Clinic", "Library", "Archive"} {
		m.Buildings[name] = 3
	}
	for _, name := range []string{"Guild Fighter", "Guild Cleric", "Guild Wizard", "Guild Rogue"} {
		m.Buildings[name] = 2
	}
	high := townWear(m)
	if high != 1 {
		t.Errorf("fully upgraded town wear = %v, want 1", high)
	}
	if high <= low {
		t.Errorf("wear must rise with upgrades: %v -> %v", low, high)
	}
}
