package crawler

import (
	"path/filepath"
	"testing"
)

// --- 6.1 town model ---

func TestBumpBuildingOpensPanel(t *testing.T) {
	tw := NewTown(DefaultMeta())
	smithy := mustFindBuilding(t, tw, "Smithy")

	// Stand one tile left of the Smithy and walk into it.
	tw.Pos = Cell{smithy.X - 1, smithy.Y}
	before := tw.Pos
	evs := tw.Move(Right)
	if tw.Pos != before {
		t.Error("bumping a building must not move you")
	}
	if tw.CurrentPanel().Building != "Smithy" {
		t.Errorf("panel %q, want Smithy", tw.CurrentPanel().Building)
	}
	if len(evs) == 0 {
		t.Error("expected a log line")
	}
}

func TestBumpGateNeedsSelectedClass(t *testing.T) {
	m := DefaultMeta()
	m.Selected = "" // defensive: no class chosen
	tw := NewTown(m)
	gate := mustFindBuilding(t, tw, "Gate")
	tw.Pos = Cell{gate.X - 1, gate.Y}
	tw.Move(Right)
	if tw.StartRequested {
		t.Error("gate must not start a run without a selected class")
	}

	if err := tw.Meta.SelectClass("Fighter"); err != nil {
		t.Fatalf("Fighter should be selectable as the starter: %v", err)
	}
	tw.Move(Right)
	if !tw.StartRequested {
		t.Error("gate must request a run start with a selected class")
	}
}

func TestLockedClassesNotSelectable(t *testing.T) {
	m := DefaultMeta()
	if err := m.SelectClass("Wizard"); err == nil {
		t.Error("Wizard must be locked with its Guild at level 0")
	}
	if got := m.UnlockedClasses(); len(got) != 1 || got[0] != "Fighter" {
		t.Errorf("unlocked roster %v, want exactly [Fighter]", got)
	}

	m.Buildings["Guild Wizard"] = 1
	if err := m.SelectClass("Wizard"); err != nil {
		t.Errorf("Wizard should be selectable at Guild 1: %v", err)
	}
}

// --- 6.2 purchase rules ---

func TestAffordablePurchase(t *testing.T) {
	m := DefaultMeta()
	m.Renown = 50
	if ok, _ := m.Buy("Smithy"); !ok {
		t.Fatal("50 renown should buy Smithy level 1")
	}
	if m.Buildings["Smithy"] != 1 {
		t.Errorf("Smithy level %d, want 1", m.Buildings["Smithy"])
	}
	if m.Renown != 0 {
		t.Errorf("renown %d, want 0", m.Renown)
	}
	if m.UpgradeCost("Smithy") <= 50 {
		t.Error("level-2 cost must rise above 50")
	}
}

func TestInsufficientRenownRejected(t *testing.T) {
	m := DefaultMeta()
	m.Renown = 10
	if ok, _ := m.Buy("Smithy"); ok {
		t.Error("purchase with insufficient renown must fail")
	}
	if m.Buildings["Smithy"] != 0 || m.Renown != 10 {
		t.Error("rejected purchase must change nothing")
	}
}

func TestPurchaseAtCap(t *testing.T) {
	m := DefaultMeta()
	m.Renown = 100000
	def := upgrades["Smithy"]
	for i := 0; i < def.Max; i++ {
		if ok, _ := m.Buy("Smithy"); !ok {
			t.Fatalf("buy %d should succeed below cap", i+1)
		}
	}
	if ok, msg := m.Buy("Smithy"); ok {
		t.Error("must refuse at cap")
	} else if msg == "" {
		t.Error("maxed refusal should report a message")
	}
}

// --- 6.3 save / load ---

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meta.json")
	m := DefaultMeta()
	m.Renown = 137
	m.Buildings["Clinic"] = 2
	m.FirstKills[BossKind] = true
	if err := m.SelectClass("Fighter"); err != nil {
		t.Fatal(err)
	}
	if err := SaveMeta(path, m); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := LoadMeta(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Renown != 137 || got.Buildings["Clinic"] != 2 || !got.FirstKills[BossKind] || got.Selected != "Fighter" {
		t.Errorf("round trip lost state: %+v", got)
	}
}

func TestFreshInstallDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	m, err := LoadMeta(path)
	if err != nil {
		t.Fatalf("missing save must load defaults: %v", err)
	}
	if m.Renown != 0 {
		t.Errorf("renown %d, want 0", m.Renown)
	}
	for name := range upgrades {
		if m.Buildings[name] != 0 {
			t.Errorf("%s level %d, want 0", name, m.Buildings[name])
		}
	}
	if len(m.UnlockedClasses()) != 1 {
		t.Errorf("fresh install must have exactly one unlocked class, got %v", m.UnlockedClasses())
	}
	if m.Selected == "" {
		t.Error("starter class must be pre-selected")
	}
}

// --- 6.4 run-start effects ---

func TestRunStartAppliesUpgrades(t *testing.T) {
	m := DefaultMeta()
	m.Buildings["Smithy"] = 2
	m.Buildings["Clinic"] = 1
	m.Buildings["Library"] = 3
	m.Buildings["Archive"] = 1
	m.Buildings["Guild Wizard"] = 2
	if err := m.SelectClass("Wizard"); err != nil {
		t.Fatal(err)
	}
	b := m.RunBonuses()
	r := NewRun(3, m.Selected, &b)
	if r.Player.StrikeDamageBonus != 2 {
		t.Errorf("Smithy: +2d expected, got %d", r.Player.StrikeDamageBonus)
	}
	if r.Player.Potions != 1 {
		t.Errorf("Clinic: 1 potion expected, got %d", r.Player.Potions)
	}
	if r.Player.WorldBonus != 3 {
		t.Errorf("Library: +3 expected, got %d", r.Player.WorldBonus)
	}
	if r.Player.SearchBonus != 1 {
		t.Errorf("Archive: +1 expected, got %d", r.Player.SearchBonus)
	}
	if !r.Enhanced["Wizard"] {
		t.Error("Guild Wizard level 2 must apply the class enhancement")
	}
}

func TestBonusesFrozenMidRun(t *testing.T) {
	m := DefaultMeta()
	m.Buildings["Smithy"] = 1
	b := m.RunBonuses()
	r := NewRun(4, "Fighter", &b)
	m.Buildings["Smithy"] = 3 // town changes while the dungeon holds
	if r.Player.StrikeDamageBonus != 1 {
		t.Errorf("active run bonuses must stay frozen, got %d", r.Player.StrikeDamageBonus)
	}
}

// mustFindBuilding locates the walkable-adjacent tile of a named building.
func mustFindBuilding(t *testing.T, tw *Town, name string) Cell {
	t.Helper()
	c, ok := tw.BuildingTile(name)
	if !ok {
		t.Fatalf("town has no building %q", name)
	}
	return c
}
