package crawler

import (
	"strings"
	"testing"
)

// --- 5.1 run lifecycle ---

func TestNewRunStartsOnFloorOneAtEntrance(t *testing.T) {
	r := NewRun(42, "Fighter", nil)
	if r.FloorNum != 1 || r.State != Running {
		t.Errorf("floor %d state %d, want 1/Running", r.FloorNum, r.State)
	}
	if !r.Floor.TileWalkable(r.Player.Pos) {
		t.Errorf("player stuck at %v", r.Player.Pos)
	}
	if len(r.Visible) == 0 {
		t.Error("run must start with computed visibility")
	}
	for _, e := range r.Floor.Enemies {
		d := abs(e.Pos.X-r.Player.Pos.X) + abs(e.Pos.Y-r.Player.Pos.Y)
		if d < 4 {
			t.Errorf("%s spawns adjacent to entrance (%d away)", e.Kind, d)
		}
	}
}

func TestWalkOntoStairsDescends(t *testing.T) {
	r := NewRun(5, "Fighter", nil)
	// Put stairs one tile to the player's right and step onto them.
	stairs := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(stairs, FeatureStairs)
	r.Move(Right)
	if r.FloorNum != 2 {
		t.Errorf("floor %d, want 2", r.FloorNum)
	}
	if r.FloorsDescended != 1 {
		t.Errorf("descended %d, want 1", r.FloorsDescended)
	}
	if countFeature(r.Floor, FeatureStairs) != 1 {
		t.Error("new floor must have its own stairs")
	}
}

func TestDeathFreezesTheRun(t *testing.T) {
	r := NewRun(6, "Fighter", nil)
	r.Player.HP = 1
	foe := placeEnemy(r, "troll", Cell{r.Player.Pos.X + 1, r.Player.Pos.Y})
	foe.AtkBonus = 50
	r.EndTurn()
	if r.State != Lost {
		t.Fatalf("state %d, want Lost", r.State)
	}
	pos := r.Player.Pos
	rp := r.Player.HP
	evs := r.Move(Right)
	if len(evs) == 0 || r.Player.Pos != pos || r.Player.HP != rp {
		t.Error("no verb may change state after death")
	}
}

func TestBossKillWinsTheRun(t *testing.T) {
	r := NewRun(7, "Fighter", nil)
	r.rollD20 = func() int { return 20 }
	boss := BossStatline()
	boss.ID = 99
	boss.Pos = Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.Enemies = append(r.Floor.Enemies, boss)
	// Power Attack crit 4d8+4 may not finish the tripled boss; swing until dead.
	for i := 0; i < 3 && r.State == Running; i++ {
		if boss.HP > 0 && r.Player.ActionsLeft >= 1 {
			r.Move(Right)
		}
	}
	if boss.HP > 0 {
		boss.HP = 0
		r.awardKill(boss)
	}
	if r.State != Won {
		t.Errorf("state %d, want Won after boss death", r.State)
	}
}

// --- 5.2 feature verbs ---

func TestBumpLockedDoorForcedByTrained(t *testing.T) {
	r := NewRun(8, "Fighter", nil)
	r.rollD20 = func() int { return 20 }
	door := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFloor(door)
	r.Floor.SetFeature(door, FeatureLockedDoor)
	actions := r.Player.ActionsLeft
	r.Move(Right)
	if r.Floor.FeatureAt(door) != FeatureOpenDoor {
		t.Errorf("door still %v after nat-20 Athletics", r.Floor.FeatureAt(door))
	}
	if r.Player.ActionsLeft != actions-1 {
		t.Error("forcing costs 1 action")
	}
	r.Move(Right) // now passable
	if r.Player.Pos != door {
		t.Error("opened door should be walkable")
	}
}

func TestBumpLockedDoorFailureKeepsItLockedAndSpends(t *testing.T) {
	r := NewRun(8, "Fighter", nil)
	r.rollD20 = func() int { return 1 }
	door := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFloor(door)
	r.Floor.SetFeature(door, FeatureLockedDoor)
	actions := r.Player.ActionsLeft
	r.Move(Right)
	if r.Floor.FeatureAt(door) != FeatureLockedDoor {
		t.Error("failed force must keep the door locked")
	}
	if r.Player.ActionsLeft != actions-1 {
		t.Error("failed force still spends the action")
	}
}

func TestUntrainedForceFailsWithoutOpening(t *testing.T) {
	r := NewRun(8, "Wizard", nil) // Wizard lacks Athletics
	r.rollD20 = func() int { return 20 }
	door := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFloor(door)
	r.Floor.SetFeature(door, FeatureLockedDoor)
	r.Move(Right)
	if r.Floor.FeatureAt(door) != FeatureLockedDoor {
		t.Error("untrained bump must not open the door")
	}
}

func TestBumpLockedDoorPickedByRogue(t *testing.T) {
	r := NewRun(8, "Rogue", nil)
	r.rollD20 = func() int { return 20 }
	door := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFloor(door)
	r.Floor.SetFeature(door, FeatureLockedDoor)
	actions := r.Player.ActionsLeft
	evs := r.Move(Right)
	if r.Floor.FeatureAt(door) != FeatureOpenDoor {
		t.Errorf("door still %v after nat-20 Thievery", r.Floor.FeatureAt(door))
	}
	if r.Player.ActionsLeft != actions-1 {
		t.Error("picking costs 1 action")
	}
	var texts strings.Builder
	for _, ev := range evs {
		texts.WriteString(ev.Text)
	}
	if !strings.Contains(texts.String(), "lock") {
		t.Errorf("rogue success should read as lockpicking, got: %s", texts.String())
	}
	r.Move(Right) // now passable
	if r.Player.Pos != door {
		t.Error("picked door should be walkable")
	}
}

func TestRoguePickFailureKeepsDoorLockedAndSpends(t *testing.T) {
	r := NewRun(8, "Rogue", nil)
	r.rollD20 = func() int { return 1 }
	door := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFloor(door)
	r.Floor.SetFeature(door, FeatureLockedDoor)
	actions := r.Player.ActionsLeft
	r.Move(Right)
	if r.Floor.FeatureAt(door) != FeatureLockedDoor {
		t.Error("failed pick must keep the door locked")
	}
	if r.Player.ActionsLeft != actions-1 {
		t.Error("failed pick still spends the action")
	}
}

func TestTrapTriggersOnceOnEntry(t *testing.T) {
	r := NewRun(9, "Fighter", nil)
	trap := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	hp := r.Player.HP
	r.Move(Right)
	if r.Player.HP >= hp {
		t.Error("armed trap should damage the walker")
	}
	if r.Floor.FeatureAt(trap) != FeatureTrapDisabled {
		t.Error("triggered trap must be spent")
	}
}

func TestTrapDisabledByTrainedThievery(t *testing.T) {
	r := NewRun(10, "Rogue", nil)
	r.rollD20 = func() int { return 20 }
	trap := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	hp := r.Player.HP
	r.Move(Right) // bump the trap tile: Thievery attempt, no entry
	if r.Floor.FeatureAt(trap) != FeatureTrapDisabled {
		t.Error("nat-20 Thievery should disarm")
	}
	if r.Player.HP != hp {
		t.Error("disarming must not trigger the trap")
	}
	if r.Player.Pos == trap {
		t.Error("disarming must not move onto the tile")
	}
}

func TestHiddenDoorRevealedBySearch(t *testing.T) {
	r := NewRun(11, "Rogue", nil)
	r.rollD20 = func() int { return 20 }
	hidden := Cell{r.Player.Pos.X + 2, r.Player.Pos.Y}
	r.Floor.SetFeature(hidden, FeatureHiddenDoor)
	r.Search()
	if r.Floor.FeatureAt(hidden) != FeatureOpenDoor {
		t.Errorf("hidden door still %v after nat-20 Search", r.Floor.FeatureAt(hidden))
	}
	if !r.Floor.TileWalkable(hidden) {
		t.Error("revealed door must be walkable")
	}
}

func TestSearchMissesAndCostsAction(t *testing.T) {
	r := NewRun(12, "Rogue", nil)
	r.rollD20 = func() int { return 1 }
	hidden := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(hidden, FeatureHiddenDoor)
	actions := r.Player.ActionsLeft
	evs := r.Search()
	if r.Floor.FeatureAt(hidden) != FeatureHiddenDoor {
		t.Error("failed Search must not reveal")
	}
	if r.Player.ActionsLeft != actions-1 {
		t.Error("Search costs 1 action")
	}
	if len(evs) == 0 {
		t.Error("expected a log line")
	}
}

func TestPotionPickupAndQuaff(t *testing.T) {
	r := NewRun(13, "Fighter", nil)
	pot := Cell{r.Player.Pos.X + 1, r.Player.Pos.Y}
	r.Floor.SetFeature(pot, FeaturePotion)
	r.Move(Right)
	if r.Player.Potions != 1 {
		t.Errorf("carrying %d potions, want 1", r.Player.Potions)
	}
	if r.Floor.FeatureAt(pot) != FeatureNone {
		t.Error("picked tile must be empty")
	}
	r.Player.HP = 5
	actions := r.Player.ActionsLeft
	r.Quaff()
	if r.Player.HP <= 5 {
		t.Error("quaff must heal")
	}
	if r.Player.Potions != 0 {
		t.Error("quaff consumes the potion")
	}
	if r.Player.ActionsLeft != actions {
		t.Error("quaffing is free")
	}
}

// --- 5.3 renown payout ---

func TestPayoutDeathSumsKillsAndFloors(t *testing.T) {
	r := NewRun(14, "Fighter", nil)
	for _, kind := range []string{"rat", "goblin"} {
		e := placeEnemy(r, kind, Cell{r.Player.Pos.X, r.Player.Pos.Y + 1})
		e.HP = 0
		r.awardKill(e)
	}
	r.FloorsDescended = 2
	want := RenownPerKill(0) + RenownPerKill(1) + 2*FloorRenown
	if got := r.Payout(false); got != want {
		t.Errorf("payout %d, want %d", got, want)
	}
}

func TestPayoutVictoryFirstKillBonus(t *testing.T) {
	r := NewRun(15, "Fighter", nil)
	boss := BossStatline()
	boss.ID = 1
	boss.Pos = Cell{2, 2}
	r.Floor.Enemies = append(r.Floor.Enemies, boss)
	boss.HP = 0
	r.awardKill(boss)
	if got, with := r.Payout(false), r.Payout(true); with-got != BossFirstKillRenown {
		t.Errorf("first-kill bonus missing: %d vs %d", got, with)
	}
}
