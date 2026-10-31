package crawler

import (
	"strings"
	"testing"
)

// --- 18.1 hidden traps ---

func TestTrapSpottedSet(t *testing.T) {
	r := NewRun(21, "Fighter", nil)
	trap := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	if r.Floor.TrapSpotted(trap) {
		t.Error("fresh traps start unspotted")
	}
	r.Floor.SpotTrap(trap)
	if !r.Floor.TrapSpotted(trap) {
		t.Error("SpotTrap must stick")
	}
	other := Cell{r.Player.Pos.X + 2, r.Player.Pos.Y}
	if r.Floor.TrapSpotted(other) {
		t.Error("spotting one trap must not spot its neighbour")
	}
}

func TestSearchSpotsTrapInRange(t *testing.T) {
	r := NewRun(22, "Rogue", nil)
	r.rollD20 = func() int { return 20 }
	trap := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	evs := r.Search()
	if !r.Floor.TrapSpotted(trap) {
		t.Fatal("nat-20 Search must spot the nearby trap")
	}
	if len(evs) == 0 {
		t.Fatal("expected a log line")
	}
	if !strings.Contains(strings.ToLower(evs[0].Text), "trap") {
		t.Errorf("trap find must name the trap: %q", evs[0].Text)
	}
	if evs[0].DC != TrapSpotDC {
		t.Errorf("spot log DC = %d, want %d", evs[0].DC, TrapSpotDC)
	}
}

func TestSearchMissedTrapStaysHidden(t *testing.T) {
	r := NewRun(23, "Fighter", nil)
	r.rollD20 = func() int { return 1 }
	trap := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	r.Search()
	if r.Floor.TrapSpotted(trap) {
		t.Error("failed Search must not spot")
	}
}

func TestSearchRevealsNearestHiddenThing(t *testing.T) {
	r := NewRun(24, "Rogue", nil)
	r.rollD20 = func() int { return 20 }
	near := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(near, FeatureTrapArmed)
	far := Cell{r.Player.Pos.X + 2, r.Player.Pos.Y}
	r.Floor.SetFeature(far, FeatureHiddenDoor)
	r.Search()
	if !r.Floor.TrapSpotted(near) {
		t.Error("nearest hidden thing must be the trap")
	}
	if r.Floor.FeatureAt(far) != FeatureHiddenDoor {
		t.Error("one Search roll reveals one thing, not both")
	}
}

func TestUnspottedTrapStillTriggers(t *testing.T) {
	r := NewRun(25, "Fighter", nil)
	trap := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	hp := r.Player.HP
	r.Move(Right)
	if r.Player.HP >= hp {
		t.Error("walking onto an unspotted trap must still bite")
	}
	if r.Floor.FeatureAt(trap) != FeatureTrapDisabled {
		t.Error("triggered trap must be spent")
	}
}
