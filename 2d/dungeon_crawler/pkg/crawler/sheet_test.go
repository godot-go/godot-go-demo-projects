package crawler

import (
	"strings"
	"testing"
)

func TestCharacterSheetShowsRolledTruth(t *testing.T) {
	f := ClassStatline("Fighter", 5)
	f.WorldBonus = 2
	f.StrikeDamageBonus = 3
	text := CharacterSheet(f, true)
	for _, want := range []string{
		"FIGHTER — Level 5",
		"Class DC 21",
		// Sheet totals must equal what the verbs roll:
		"T Athletics +13", // STR 4 + trained 7 + Library 2
		"T Perception +9", // WIS 0 + trained 7 + Library 2
		"STR +4",
		"Guild-enhanced",
		"Smithy +3 strike damage",
		"Library +2 skill checks",
		"Longsword 1d8 +4",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("Fighter sheet missing %q, got:\n%s", want, text)
		}
	}
	// Untrained rows carry the bare ability (no Library bonus).
	if !strings.Contains(text, "Stealth +1") {
		t.Errorf("Fighter sheet should show untrained Stealth +1, got:\n%s", text)
	}
}

func TestCharacterSheetWithoutEnhancement(t *testing.T) {
	r := ClassStatline("Rogue", 1)
	text := CharacterSheet(r, false)
	if strings.Contains(text, "Guild-enhanced") {
		t.Errorf("unenhanced sheet mentions Guild-enhanced:\n%s", text)
	}
	if !strings.Contains(text, "Sneak Attack") {
		t.Errorf("Rogue sheet missing class skill blurb:\n%s", text)
	}
	// No town bonuses: the active-bonuses line disappears entirely.
	if strings.Contains(text, "Active bonuses") {
		t.Errorf("bare sheet shows Active bonuses line:\n%s", text)
	}
}
