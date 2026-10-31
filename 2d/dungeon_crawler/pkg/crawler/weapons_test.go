package crawler

import (
	"math/rand"
	"strings"
	"testing"
)

// giveWeapon appends loot to the pack and returns its index.
func giveWeapon(p *Player, w Weapon) int {
	p.Weapons = append(p.Weapons, w)
	return len(p.Weapons) - 1
}

func lootByName(t *testing.T, name string) Weapon {
	t.Helper()
	for _, w := range LootWeapons {
		if w.Name == name {
			return w
		}
	}
	t.Fatalf("no loot weapon %q", name)
	return Weapon{}
}

func TestStarterWeaponsMatchClassDice(t *testing.T) {
	want := map[string]string{
		"Fighter": "Longsword",
		"Cleric":  "Mace",
		"Wizard":  "Club",
		"Rogue":   "Short Sword",
	}
	for class, name := range want {
		p := ClassStatline(class, 1)
		if len(p.Weapons) != 1 || p.Equipped != 0 {
			t.Fatalf("%s: pack %v equipped %d, want one starter at 0", class, p.Weapons, p.Equipped)
		}
		w := p.EquippedWeapon()
		if w.Name != name {
			t.Errorf("%s: starter %q, want %q", class, w.Name, name)
		}
		if w.Dice != Classes[class].Weapon {
			t.Errorf("%s: starter dice %v, want tuned %v", class, w.Dice, Classes[class].Weapon)
		}
	}
}

func TestStrikeDamageFollowsEquipped(t *testing.T) {
	p := ClassStatline("Fighter", 1)
	if got := p.StrikeDamage(); got != (Dice{1, 8}) {
		t.Fatalf("starter strike dice %v, want 1d8", got)
	}
	p.Equipped = giveWeapon(p, lootByName(t, "Greataxe"))
	if got := p.StrikeDamage(); got != (Dice{1, 12}) {
		t.Errorf("greataxe strike dice %v, want 1d12", got)
	}
}

func TestFinesseUsesBestAbility(t *testing.T) {
	rogue := ClassStatline("Rogue", 1) // short sword is finesse, DEX 4
	if got := rogue.StrikeMeleeMod(); got != 4 {
		t.Errorf("rogue finesse mod %d, want 4", got)
	}
	wiz := ClassStatline("Wizard", 1) // club (STR -1); rapier would use DEX +1
	if got := wiz.StrikeMeleeMod(); got != -1 {
		t.Errorf("wizard club mod %d, want -1", got)
	}
	wiz.Equipped = giveWeapon(wiz, lootByName(t, "Rapier"))
	if got := wiz.StrikeMeleeMod(); got != 1 {
		t.Errorf("wizard rapier mod %d, want 1 (finesse picks DEX over STR)", got)
	}
}

func TestWeaponProfMod(t *testing.T) {
	axe := lootByName(t, "Battle Axe")
	cases := []struct {
		class  string
		weapon Weapon
		want   int
	}{
		{"Fighter", axe, 0}, // trained in martial
		{"Fighter", ClassStatline("Fighter", 1).EquippedWeapon(), 0},
		{"Wizard", axe, -2}, // untrained martial penalty
		{"Rogue", axe, -2},  // PF2e rogues get simple weapons only
		{"Wizard", ClassStatline("Wizard", 1).EquippedWeapon(), 0},
	}
	for _, c := range cases {
		p := ClassStatline(c.class, 1)
		p.Equipped = giveWeapon(p, c.weapon)
		if got := p.WeaponProfMod(); got != c.want {
			t.Errorf("%s with %s: prof mod %d, want %d", c.class, c.weapon.Name, got, c.want)
		}
	}
}

func TestEquipWeaponActionEconomy(t *testing.T) {
	r := newTestRun("Fighter")
	p := r.Player
	axeIdx := giveWeapon(p, lootByName(t, "Battle Axe"))

	evs := r.EquipWeapon(axeIdx)
	if p.ActionsLeft != 2 {
		t.Errorf("equip cost %d actions, want 1 (PF2e Interact)", 3-p.ActionsLeft)
	}
	if p.Equipped != axeIdx {
		t.Errorf("equipped %d, want %d", p.Equipped, axeIdx)
	}
	if len(evs) == 0 || !strings.Contains(evs[0].Text, "Battle Axe") {
		t.Errorf("events %v, want one naming the weapon", evs)
	}

	before := p.ActionsLeft
	evs = r.EquipWeapon(axeIdx) // already wielded — free
	if p.ActionsLeft != before {
		t.Error("re-equipping the same weapon must not cost an action")
	}
	if len(evs) == 0 || !strings.Contains(strings.ToLower(evs[0].Text), "already") {
		t.Errorf("events %v, want an 'already' notice", evs)
	}

	p.ActionsLeft = 0
	greatIdx := giveWeapon(p, lootByName(t, "Greataxe"))
	evs = r.EquipWeapon(greatIdx)
	if p.Equipped != axeIdx {
		t.Error("equip must be refused at 0 actions")
	}
	if len(evs) == 0 || !strings.Contains(evs[0].Text, "No actions") {
		t.Errorf("events %v, want a refusal", evs)
	}
}

func TestEquipWeaponInvalidIndex(t *testing.T) {
	r := newTestRun("Fighter")
	before := r.Player.ActionsLeft
	if evs := r.EquipWeapon(42); len(evs) == 0 {
		t.Fatal("want a notice for invalid index")
	}
	if r.Player.ActionsLeft != before || r.Player.Equipped != 0 {
		t.Error("invalid equip must not spend actions or swap weapons")
	}
}

func TestUntrainedMartialPenaltyOnRoll(t *testing.T) {
	r := newTestRun("Wizard")
	r.rollD20 = func() int { return 10 }
	foe := placeEnemy(r, "goblin", Cell{2, 1})

	hit := func(t *testing.T) Event {
		t.Helper()
		for _, ev := range r.Move(Right) {
			if ev.Kind == "attack" && ev.Total != 0 {
				return ev
			}
		}
		t.Fatal("no attack roll event")
		return Event{}
	}
	club := hit(t)
	if want := 10 + p0(r) + r.Player.Level + 2; club.Total != want {
		t.Fatalf("club total %d, want d20 10 + STR -1 + trained prof (want %d)", club.Total, want)
	}
	// reset for the second swing
	r.Player.Pos = Cell{1, 1}
	r.Player.StrikesThisTurn = 0
	r.Player.Equipped = giveWeapon(r.Player, lootByName(t, "Battle Axe"))
	axe := hit(t)
	if axe.Total != club.Total-2 {
		t.Errorf("battle axe total %d, want untrained penalty: %d (club %d minus 2)", axe.Total, club.Total-2, club.Total)
	}
	_ = foe
}

func p0(r *Run) int { return r.Player.AbilityMod("str") }

func TestDeadlyCritAddsBonusDie(t *testing.T) {
	axe := lootByName(t, "Battle Axe") // 1d8, deadly d12
	anyAboveBase := false
	for seed := int64(1); seed <= 25; seed++ {
		r := newTestRun("Fighter")
		r.rng = rand.New(rand.NewSource(seed))
		r.rollD20 = func() int { return 20 } // guaranteed crit
		r.Player.Equipped = giveWeapon(r.Player, axe)
		placeEnemy(r, "troll", Cell{2, 1}) // beefy enough to survive
		var dmg int
		for _, ev := range r.Move(Right) {
			if ev.Kind == "attack" && ev.Damage > 0 {
				dmg = ev.Damage
			}
		}
		if dmg == 0 {
			t.Fatalf("seed %d: no damage event", seed)
		}
		// crit 2d8 + STR 4 + deadly 1d12 ∈ [7, 32]; no-deadly max is 20.
		if dmg < 7 || dmg > 32 {
			t.Fatalf("seed %d: crit damage %d outside [7,32]", seed, dmg)
		}
		if dmg > 20 {
			anyAboveBase = true
		}
	}
	if !anyAboveBase {
		t.Error("deadly d12 never boosted a crit above the 1d8-only maximum — bonus die missing")
	}
}

func TestWeaponDropsPlacedPerFloor(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		rng := rand.New(rand.NewSource(seed))
		for floorNum := 1; floorNum <= 5; floorNum++ {
			g := GenerateFloor(rng)
			nextID := 1
			PlaceFeatures(rng, g, floorNum, func(kind string) *Enemy {
				e := MonsterStatline(kind)
				e.ID = nextID
				nextID++
				return e
			})
			count := 0
			for _, f := range g.Floor.Features {
				if f == FeatureWeapon {
					count++
				}
			}
			if count < 1 {
				t.Errorf("seed %d floor %d: %d weapon drops, want at least 1", seed, floorNum, count)
			}
		}
	}
}

func TestWeaponPickupGrowsInventory(t *testing.T) {
	r := newTestRun("Fighter")
	before := len(r.Player.Weapons)
	r.Floor.SetFeature(Cell{2, 1}, FeatureWeapon)
	evs := r.Move(Right)
	if len(r.Player.Weapons) != before+1 {
		t.Fatalf("pack %v, want one more weapon", r.Player.Weapons)
	}
	if r.Floor.FeatureAt(Cell{2, 1}) != FeatureNone {
		t.Error("weapon tile must be cleared after pickup")
	}
	joined := ""
	for _, ev := range evs {
		joined += ev.Text
	}
	if !strings.Contains(joined, "You find") {
		t.Errorf("events %q, want a find notice", joined)
	}
}

func TestDamageEventsCarrySubjectAndLocation(t *testing.T) {
	r := newTestRun("Fighter")
	r.rollD20 = func() int { return 20 }
	foe := placeEnemy(r, "troll", Cell{2, 1})
	for _, ev := range r.Move(Right) {
		if ev.Kind == "attack" && ev.Damage > 0 {
			if ev.Subject != "enemy" || ev.At != foe.Pos || ev.EnemyID != foe.ID {
				t.Errorf("player strike event subject %q at %v id %d, want enemy at %v id %d", ev.Subject, ev.At, ev.EnemyID, foe.Pos, foe.ID)
			}
			break
		}
	}

	r2 := newTestRun("Fighter")
	r2.rollD20 = func() int { return 20 }
	placeEnemy(r2, "troll", Cell{2, 1})
	for _, ev := range r2.EndTurn() {
		if ev.Kind == "attack" && ev.Damage > 0 {
			if ev.Subject != "player" || ev.At != r2.Player.Pos {
				t.Errorf("enemy strike event subject %q at %v, want player at %v", ev.Subject, ev.At, r2.Player.Pos)
			}
			return
		}
	}
	t.Error("enemy phase produced no damage event")
}
