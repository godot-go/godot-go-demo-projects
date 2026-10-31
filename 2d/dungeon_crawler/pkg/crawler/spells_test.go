package crawler

import (
	"strings"
	"testing"
)

// wizardRun builds a wizard run on the open corridor with the book synced
// to the given level. Callers place their own foes.
func wizardRun(t *testing.T, level int) *Run {
	t.Helper()
	r := newTestRun("Wizard")
	r.Player.Level = level
	syncWizardBook(r.Player)
	return r
}

func spellNamed(p *Player, name string) (Spell, int, bool) {
	for i, s := range p.Spells {
		if s.Name == name {
			return s, i, true
		}
	}
	return Spell{}, -1, false
}

// damageEvents returns the enemy damage events of a cast.
func struck(evs []Event) []Event {
	var hits []Event
	for _, ev := range evs {
		if ev.Damage > 0 {
			hits = append(hits, ev)
		}
	}
	return hits
}

func TestWizardStartsWithCantrips(t *testing.T) {
	p := ClassStatline("Wizard", 1)
	if len(p.Spells) != 2 {
		t.Fatalf("book %v, want Ray of Frost and Electric Arc", p.Spells)
	}
	if p.PreparedSpell().Name != "Ray of Frost" {
		t.Errorf("prepared %q, want Ray of Frost (index 0)", p.PreparedSpell().Name)
	}
	for _, s := range p.Spells {
		if !s.Cantrip {
			t.Errorf("%s should be a cantrip at level 1", s.Name)
		}
	}
}

func TestNonWizardHasNoSpells(t *testing.T) {
	for _, class := range []string{"Fighter", "Cleric", "Rogue"} {
		if p := ClassStatline(class, 1); len(p.Spells) != 0 {
			t.Errorf("%s book %v, want empty", class, p.Spells)
		}
	}
}

func TestSpellGrantsByLevel(t *testing.T) {
	p := ClassStatline("Wizard", 1)
	p.Level = 2
	syncWizardBook(p)
	if _, _, ok := spellNamed(p, "Magic Missile"); !ok {
		t.Fatalf("level 2 must grant Magic Missile, book %v", p.Spells)
	}
	p.Level = 3
	syncWizardBook(p)
	for _, name := range []string{"Burning Hands", "Fire Burst"} {
		if _, _, ok := spellNamed(p, name); !ok {
			t.Errorf("level 3 must grant %s, book %v", name, p.Spells)
		}
	}
	if len(p.Spells) != 6 {
		t.Errorf("book %v, want six spells total", p.Spells)
	}
	syncWizardBook(p) // idempotent — no duplicates on re-sync
	if len(p.Spells) != 6 {
		t.Errorf("re-sync duplicated entries: %v", p.Spells)
	}
}

func TestCantripsCastOnRepeat(t *testing.T) {
	r := wizardRun(t, 1)
	r.rollD20 = func() int { return 20 }
	for cast := 1; cast <= 3; cast++ {
		placeEnemy(r, "troll", Cell{2, 1}) // fresh target; crit 2d6+5 < troll HP
		r.Player.ActionsLeft = 3
		if len(struck(r.UseSkill("ray"))) == 0 {
			t.Fatalf("cantrip cast %d dealt no damage", cast)
		}
	}
}

func TestElectricArcStrikesTwoNearest(t *testing.T) {
	r := wizardRun(t, 1)
	r.rollD20 = func() int { return 20 }
	placeEnemy(r, "rat", Cell{2, 1}) // dist 1
	placeEnemy(r, "rat", Cell{3, 1}) // dist 2, same open row (walls block arcs)
	far := placeEnemy(r, "rat", Cell{5, 1})
	_, idx, _ := spellNamed(r.Player, "Electric Arc")
	r.PrepareSpell(idx)

	hits := struck(r.UseSkill("ray"))
	if len(hits) != 2 {
		t.Fatalf("arc hit %d foes, want the two nearest", len(hits))
	}
	for _, ev := range hits {
		if ev.Subject != "enemy" || ev.EnemyID == 0 {
			t.Errorf("arc event needs enemy subject/id for popups, got %+v", ev)
		}
		if ev.EnemyID == far.ID {
			t.Error("out-of-reach foe was struck")
		}
	}
}

func TestMagicMissileAutoHitsAndSpendsSlot(t *testing.T) {
	r := wizardRun(t, 2)
	placeEnemy(r, "troll", Cell{4, 1})
	r.rollD20 = func() int { return 1 } // would miss if missiles rolled
	_, idx, _ := spellNamed(r.Player, "Magic Missile")
	r.PrepareSpell(idx)

	hits := struck(r.UseSkill("ray"))
	if len(hits) != 1 {
		t.Fatalf("missiles produced %d damage events, want 1 struck foe", len(hits))
	}
	// 2 bolts at level 2: [4, 10] force, guaranteed even on a scripted 1.
	if dmg := hits[0].Damage; dmg < 4 || dmg > 10 {
		t.Fatalf("missile damage %d outside [4,10]", dmg)
	}
	if r.Player.Slots[1] != 0 {
		t.Errorf("slots after cast %d, want 0 remaining", r.Player.Slots[1])
	}

	r.Player.ActionsLeft = 3
	before := r.Floor.Enemies[0].HP
	evs := r.UseSkill("ray") // dry
	if r.Floor.Enemies[0].HP != before {
		t.Error("cast with no slots still damaged the foe")
	}
	joined := ""
	for _, ev := range evs {
		joined += ev.Text
	}
	if !strings.Contains(strings.ToLower(joined), "slot") {
		t.Errorf("expected a slot-exhausted notice, got %q", joined)
	}
}

func TestSlotsRefreshOnDescend(t *testing.T) {
	r := wizardRun(t, 2)
	placeEnemy(r, "troll", Cell{2, 1})
	_, idx, _ := spellNamed(r.Player, "Magic Missile")
	r.PrepareSpell(idx)
	r.UseSkill("ray") // spend the single 1st-level slot
	if r.Player.Slots[1] != 0 {
		t.Fatalf("slot not spent")
	}
	r.descend()
	if r.Player.Slots[1] != 1 {
		t.Errorf("after descend slots %d, want refreshed to 1", r.Player.Slots[1])
	}
}

func TestSlotTotalsScaleWithLevel(t *testing.T) {
	want := wizardSlotTotals(3)
	if want[1] != 2 || want[2] != 1 {
		t.Errorf("level 3 slot totals %v, want 2 first / 1 second", want)
	}
	if len(wizardSlotTotals(1)) != 0 {
		t.Error("level 1 should have no leveled slots yet")
	}
}

func TestBurningHandsHitsEveryAdjacentFoe(t *testing.T) {
	r := wizardRun(t, 3)
	placeEnemy(r, "rat", Cell{2, 1}) // dist 1
	placeEnemy(r, "rat", Cell{1, 2}) // dist 1
	far := placeEnemy(r, "rat", Cell{3, 3})
	_, idx, _ := spellNamed(r.Player, "Burning Hands")
	r.PrepareSpell(idx)

	hits := struck(r.UseSkill("ray"))
	if len(hits) != 2 {
		t.Fatalf("burning hands hit %d, want both adjacent foes only", len(hits))
	}
	for _, ev := range hits {
		if ev.EnemyID == far.ID {
			t.Error("distant foe burned")
		}
	}
}

func TestFireBurstHitsWithinTwoForTwoActions(t *testing.T) {
	r := wizardRun(t, 3)
	placeEnemy(r, "rat", Cell{2, 1})
	placeEnemy(r, "rat", Cell{2, 2}) // dist 2
	safe := placeEnemy(r, "rat", Cell{5, 1})
	_, idx, _ := spellNamed(r.Player, "Fire Burst")
	r.PrepareSpell(idx)

	before := r.Player.ActionsLeft
	hits := struck(r.UseSkill("ray"))
	if len(hits) != 2 {
		t.Fatalf("fire burst hit %d, want the two foes within distance 2", len(hits))
	}
	for _, ev := range hits {
		if ev.EnemyID == safe.ID {
			t.Error("foe outside the burst was struck")
		}
	}
	if spent := before - r.Player.ActionsLeft; spent != 2 {
		t.Errorf("Fire Burst cost %d, want 2 actions", spent)
	}
}

func TestSpellCastNeedsActions(t *testing.T) {
	r := wizardRun(t, 1)
	placeEnemy(r, "rat", Cell{2, 1})
	before := r.Floor.Enemies[0].HP
	r.Player.ActionsLeft = 0
	if evs := r.UseSkill("ray"); len(evs) == 0 || r.Floor.Enemies[0].HP != before {
		t.Error("spell fired at 0 actions or produced no refusal")
	}
}

func TestPrepareSpellIsValidated(t *testing.T) {
	r := wizardRun(t, 1)
	before := r.Player.ActionsLeft
	if len(r.PrepareSpell(1)) == 0 {
		t.Error("valid prepare should log a line")
	}
	if r.Player.Prepared != 1 {
		t.Errorf("prepared index %d, want 1", r.Player.Prepared)
	}
	if r.Player.ActionsLeft != before {
		t.Error("preparing a spell must be free")
	}
	r.PrepareSpell(42)
	if r.Player.Prepared != 1 {
		t.Error("invalid prepare must not change the prepared spell")
	}
}

func TestKnockLearnsAtLevelTwo(t *testing.T) {
	p := ClassStatline("Wizard", 1)
	p.Level = 2
	syncWizardBook(p)
	s, _, ok := spellNamed(p, "Knock")
	if !ok {
		t.Fatalf("level 2 must grant Knock, book %v", p.Spells)
	}
	if !s.Cantrip || s.Actions != 1 {
		t.Errorf("Knock %+v, want a free one-action cantrip", s)
	}
}

func TestKnockSwingsAdjacentLockedDoor(t *testing.T) {
	r := wizardRun(t, 2)
	_, idx, _ := spellNamed(r.Player, "Knock")
	r.PrepareSpell(idx)
	door := Cell{2, 1}
	r.Floor.SetFeature(door, FeatureLockedDoor)
	actions := r.Player.ActionsLeft
	slotBefore := r.Player.Slots[1]
	r.UseSkill("ray") // dispatches to the prepared spell
	if r.Floor.FeatureAt(door) != FeatureOpenDoor {
		t.Errorf("door still %v after Knock", r.Floor.FeatureAt(door))
	}
	if r.Player.ActionsLeft != actions-1 {
		t.Error("Knock costs 1 action")
	}
	if r.Player.Slots[1] != slotBefore {
		t.Error("cantrip Knock must not spend a slot")
	}
	r.Move(Right) // now passable
	if r.Player.Pos != door {
		t.Error("knocked-open door should be walkable")
	}
}

func TestKnockSwingsDiagonalDoor(t *testing.T) {
	r := wizardRun(t, 2)
	_, idx, _ := spellNamed(r.Player, "Knock")
	r.PrepareSpell(idx)
	door := Cell{2, 2}
	r.Floor.SetFeature(door, FeatureLockedDoor)
	r.UseSkill("ray")
	if r.Floor.FeatureAt(door) != FeatureOpenDoor {
		t.Errorf("diagonal door still %v after Knock", r.Floor.FeatureAt(door))
	}
}

// --- spells destroy armed traps ---

func TestRayDestroysArmedTrapWhenNoFoes(t *testing.T) {
	r := wizardRun(t, 1)
	trap := Cell{3, 1}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	before := r.Player.ActionsLeft

	evs := r.UseSkill("ray")

	if got := r.Floor.FeatureAt(trap); got != FeatureTrapDisabled {
		t.Fatalf("ray left trap %v, want disabled", got)
	}
	if r.Player.ActionsLeft != before-1 {
		t.Errorf("actions %d, want %d spent", r.Player.ActionsLeft, before-1)
	}
	if len(struck(evs)) != 0 {
		t.Errorf("ray damaged a foe while hitting a trap: %+v", struck(evs))
	}
}

func TestRayStillTargetsFoesOverTraps(t *testing.T) {
	r := wizardRun(t, 1)
	r.rollD20 = func() int { return 20 }
	foe := placeEnemy(r, "rat", Cell{2, 1})
	r.Floor.SetFeature(Cell{5, 1}, FeatureTrapArmed)

	r.UseSkill("ray")

	if foe.HP >= 5 {
		t.Errorf("ray ignored the foe in reach (HP %d)", foe.HP)
	}
	if got := r.Floor.FeatureAt(Cell{5, 1}); got != FeatureTrapArmed {
		t.Errorf("trap was disturbed while a foe stood in reach: %v", got)
	}
}

func TestElectricArcDestroysTrapWhenNoFoes(t *testing.T) {
	r := wizardRun(t, 1)
	trap := Cell{3, 1}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	_, idx, _ := spellNamed(r.Player, "Electric Arc")
	r.PrepareSpell(idx)

	r.UseSkill("ray")

	if got := r.Floor.FeatureAt(trap); got != FeatureTrapDisabled {
		t.Fatalf("arc left trap %v, want disabled", got)
	}
}

func TestMagicMissileDestroysTrapAndSpendsSlot(t *testing.T) {
	r := wizardRun(t, 2)
	trap := Cell{5, 1}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	_, idx, _ := spellNamed(r.Player, "Magic Missile")
	r.PrepareSpell(idx)
	slots := r.Player.Slots[1]

	r.UseSkill("ray")

	if got := r.Floor.FeatureAt(trap); got != FeatureTrapDisabled {
		t.Fatalf("missile left trap %v, want disabled", got)
	}
	if r.Player.Slots[1] != slots-1 {
		t.Errorf("slots %d, want %d spent", r.Player.Slots[1], slots-1)
	}
}

func TestBurningHandsCatchesAdjacentTrapWithFoes(t *testing.T) {
	r := wizardRun(t, 3)
	placeEnemy(r, "rat", Cell{2, 1})
	trap := Cell{0, 1}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	_, idx, _ := spellNamed(r.Player, "Burning Hands")
	r.PrepareSpell(idx)

	evs := r.UseSkill("ray")

	if len(struck(evs)) == 0 {
		t.Errorf("adjacent foe took no damage: %+v", evs)
	}
	if got := r.Floor.FeatureAt(trap); got != FeatureTrapDisabled {
		t.Errorf("blast left trap %v, want disabled", got)
	}
}

func TestFireBurstCatchTrapsAmongFoes(t *testing.T) {
	r := wizardRun(t, 3)
	placeEnemy(r, "rat", Cell{2, 1})
	trap := Cell{3, 1}
	r.Floor.SetFeature(trap, FeatureTrapArmed)
	_, idx, _ := spellNamed(r.Player, "Fire Burst")
	r.PrepareSpell(idx)

	r.UseSkill("ray")

	if got := r.Floor.FeatureAt(trap); got != FeatureTrapDisabled {
		t.Errorf("burst left trap %v, want disabled", got)
	}
}

func TestTrapOnlyFizzleCommitsNothing(t *testing.T) {
	r := wizardRun(t, 2) // holds a slot to prove it stays untouched
	trap := Cell{5, 1}
	r.Floor.SetFeature(trap, FeatureTrapDisabled) // no armed trap left
	_, idx, _ := spellNamed(r.Player, "Magic Missile")
	r.PrepareSpell(idx)
	acts, slots := r.Player.ActionsLeft, r.Player.Slots[1]

	r.UseSkill("ray")

	if got := r.Floor.FeatureAt(trap); got != FeatureTrapDisabled {
		t.Errorf("feature changed: %v", got)
	}
	if r.Player.ActionsLeft != acts || r.Player.Slots[1] != slots {
		t.Errorf("fizzle spent resources: actions %d, slots %d", r.Player.ActionsLeft, r.Player.Slots[1])
	}
}

func TestKnockFizzleSpendsNothing(t *testing.T) {
	r := wizardRun(t, 2)
	_, idx, _ := spellNamed(r.Player, "Knock")
	r.PrepareSpell(idx)
	actions := r.Player.ActionsLeft
	slotBefore := r.Player.Slots[1]
	r.UseSkill("ray")
	if r.Player.ActionsLeft != actions {
		t.Error("a fizzle must not spend the action")
	}
	if r.Player.Slots[1] != slotBefore {
		t.Error("a fizzle must not spend a slot")
	}
}
