package crawler

import (
	"strings"
	"testing"
)

// The PF2e skill layer must display the exact numbers the verbs roll with,
// so several of these tests re-derive the play-tested tuned totals.

func TestSkillTrainedRanks(t *testing.T) {
	cases := []struct {
		class, skill string
		want         bool
	}{
		{"Fighter", "Athletics", true},
		{"Fighter", "Thievery", false},
		{"Rogue", "Thievery", true},
		{"Rogue", "Athletics", false},
		{"Cleric", "Athletics", false},
		{"Cleric", "Thievery", false},
		{"Wizard", "Thievery", false},
		// Perception is trained for everyone: the Search verb relies on it.
		{"Fighter", "Perception", true},
		{"Cleric", "Perception", true},
		{"Wizard", "Perception", true},
		{"Rogue", "Perception", true},
		// Everything else is untrained in the demo roster.
		{"Rogue", "Stealth", false},
		{"Wizard", "Arcana", false},
		{"Fighter", "Lore", false},
	}
	for _, tc := range cases {
		p := ClassStatline(tc.class, 1)
		if got := p.SkillTrained(tc.skill); got != tc.want {
			t.Errorf("%s SkillTrained(%s) = %v, want %v", tc.class, tc.skill, got, tc.want)
		}
	}
}

func TestSkillModPF2eFormula(t *testing.T) {
	// Trained modifier = ability + (level + 2) + the Library's WorldBonus.
	f := ClassStatline("Fighter", 5)
	f.WorldBonus = 2
	if got := f.SkillMod("Athletics"); got != 4+7+2 {
		t.Errorf("Fighter L5 Athletics = %d, want %d", got, 4+7+2)
	}
	r := ClassStatline("Rogue", 3)
	if got := r.SkillMod("Thievery"); got != 4+5 {
		t.Errorf("Rogue L3 Thievery = %d, want %d", got, 4+5)
	}
	// Untrained rows still show the bare ability modifier.
	if got := f.SkillMod("Stealth"); got != 1 {
		t.Errorf("Fighter L5 Stealth (untrained) = %d, want 1", got)
	}
	// Perception: WIS + trained for every class.
	for class, def := range Classes {
		p := ClassStatline(class, 4)
		p.WorldBonus = 1
		want := def.Wis + 6 + 1
		if got := p.SkillMod("Perception"); got != want {
			t.Errorf("%s L4 Perception = %d, want %d", class, got, want)
		}
	}
}

func TestSkillCheckMatchesTunedTotals(t *testing.T) {
	// Fighter forcing a door: roll 10 + trained(1+2) + STR 4 = 17.
	f := ClassStatline("Fighter", 1)
	if res := f.SkillCheck("Athletics", 10, DoorDC, true); res.Total != 17 {
		t.Errorf("Fighter door total = %d, want 17", res.Total)
	}
	// A trained-only Athletics attempt by another class auto-fails.
	w := ClassStatline("Wizard", 3)
	if res := w.SkillCheck("Athletics", 20, DoorDC, true); !res.Untrained {
		t.Error("Wizard Athletics on trained-only check should be Untrained")
	}
	// Rogue disarming a trap: roll 10 + trained(3+2) + DEX 4 = 19.
	r := ClassStatline("Rogue", 3)
	if res := r.SkillCheck("Thievery", 10, TrapDC, true); res.Total != 19 {
		t.Errorf("Rogue trap total = %d, want 19", res.Total)
	}
	// Cleric searching: roll 10 + trained(2+2) + WIS 4 = 18.
	c := ClassStatline("Cleric", 2)
	if res := c.SkillCheck("Perception", 10, HiddenDoorDC, false); res.Total != 18 {
		t.Errorf("Cleric search total = %d, want 18", res.Total)
	}
}

func TestKeyAbilityPerClass(t *testing.T) {
	want := map[string]string{
		"Fighter": "str", "Cleric": "wis", "Wizard": "int", "Rogue": "dex",
	}
	for class, key := range want {
		if got := ClassStatline(class, 1).KeyAbility(); got != key {
			t.Errorf("%s KeyAbility = %q, want %q", class, got, key)
		}
	}
}

func TestClassDC(t *testing.T) {
	// class DC = 10 + trained(level+2) + key ability.
	f := ClassStatline("Fighter", 1)
	if got := f.ClassDC(); got != 17 {
		t.Errorf("Fighter L1 class DC = %d, want 17", got)
	}
	if got := ClassStatline("Fighter", 5).ClassDC(); got != 21 {
		t.Errorf("Fighter L5 class DC = %d, want 21", got)
	}
	for class, def := range Classes {
		p := ClassStatline(class, 3)
		want := 10 + 5 + p.AbilityMod(p.KeyAbility())
		if def.KeyAbility == "" {
			t.Fatalf("%s has no KeyAbility in its class def", class)
		}
		if got := p.ClassDC(); got != want {
			t.Errorf("%s L3 class DC = %d, want %d", class, got, want)
		}
	}
}

func TestSaveMods(t *testing.T) {
	// Display-only saves: trained(level+2) + CON / DEX / WIS.
	f := ClassStatline("Fighter", 5)
	if got := f.SaveMod("fort"); got != 7+3 {
		t.Errorf("Fighter L5 Fort = %d, want %d", got, 7+3)
	}
	if got := f.SaveMod("ref"); got != 7+1 {
		t.Errorf("Fighter L5 Ref = %d, want %d", got, 7+1)
	}
	if got := f.SaveMod("will"); got != 7+0 {
		t.Errorf("Fighter L5 Will = %d, want %d", got, 7+0)
	}
	r := ClassStatline("Rogue", 1)
	if got := r.SaveMod("ref"); got != 3+4 {
		t.Errorf("Rogue L1 Ref = %d, want %d", got, 3+4)
	}
}

func TestSkillNamesCanonicalList(t *testing.T) {
	want := "Acrobatics,Arcana,Athletics,Crafting,Deception,Diplomacy," +
		"Intimidation,Lore,Medicine,Nature,Occultism,Performance," +
		"Religion,Society,Stealth,Survival,Thievery,Perception"
	if got := strings.Join(SkillNames, ","); got != want {
		t.Errorf("SkillNames = %s, want %s", got, want)
	}
	for _, s := range SkillNames {
		if SkillKey(s) == "" {
			t.Errorf("skill %s has no key ability", s)
		}
	}
}
