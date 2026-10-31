package crawler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- 17.1 flavored character ---

func flavorEq(a, b CharacterFlavor) bool {
	if a.Hometown != b.Hometown || a.Boon != b.Boon || a.Sigil != b.Sigil || len(a.Tales) != len(b.Tales) {
		return false
	}
	for i := range a.Tales {
		if a.Tales[i] != b.Tales[i] {
			return false
		}
	}
	return true
}

func TestRollFlavorDeterministic(t *testing.T) {
	a, b := RollFlavor(123), RollFlavor(123)
	if !flavorEq(a, b) {
		t.Errorf("same seed rolled two cards: %+v vs %+v", a, b)
	}
	if len(a.Tales) != 4 {
		t.Fatalf("card carries %d tale rolls, want 4", len(a.Tales))
	}
	for i, roll := range a.Tales {
		if roll < 1 || roll > 20 {
			t.Errorf("tale roll %d = %d, want d20 range", i, roll)
		}
	}
}

func TestRollFlavorBounds(t *testing.T) {
	card := RollFlavor(7)
	if card.Hometown == "" || card.Boon == "" {
		t.Fatalf("card missing flavor lines: %+v", card)
	}
	if card.Sigil < 0 || card.Sigil >= len(flavorSigils) {
		t.Errorf("sigil %d unmapped, want [0,%d)", card.Sigil, len(flavorSigils))
	}
}

func TestRollFlavorVaries(t *testing.T) {
	homes := map[string]bool{}
	boons := map[string]bool{}
	for seed := int64(1); seed <= 40; seed++ {
		card := RollFlavor(seed)
		homes[card.Hometown] = true
		boons[card.Boon] = true
	}
	if len(homes) < 3 {
		t.Errorf("hometowns too uniform: %d distinct over 40 seeds", len(homes))
	}
	if len(boons) < 3 {
		t.Errorf("boons too uniform: %d distinct over 40 seeds", len(boons))
	}
}

func TestMetaFlavorRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	m := DefaultMeta()
	card := RollFlavor(99)
	m.Flavor = &card
	if err := SaveMeta(path, m); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadMeta(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Flavor == nil || !flavorEq(*loaded.Flavor, card) {
		t.Errorf("flavor did not survive the round trip: %+v", loaded.Flavor)
	}
}

func TestMetaWithoutFlavorLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	// A pre-flavor save: no Flavor key at all.
	if err := os.WriteFile(path, []byte(`{"Renown":5,"Buildings":{"Smithy":1},"FirstKills":{},"Selected":"Fighter"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadMeta(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Flavor != nil {
		t.Errorf("old save grew a flavor: %+v", m.Flavor)
	}
	if fmt.Sprint(m.Selected) != "Fighter" {
		t.Errorf("save body corrupted: %+v", m.Selected)
	}
}

func TestMetaIsFresh(t *testing.T) {
	if !DefaultMeta().IsFresh() {
		t.Error("fresh-install defaults must read fresh")
	}
	m := DefaultMeta()
	m.Renown = 12
	if m.IsFresh() {
		t.Error("banked renown is not fresh")
	}
	m = DefaultMeta()
	m.Buildings["Smithy"] = 1
	if m.IsFresh() {
		t.Error("bought levels are not fresh")
	}
	m = DefaultMeta()
	m.FirstKills["Warden"] = true
	if m.IsFresh() {
		t.Error("first kills are not fresh")
	}
	m = DefaultMeta()
	m.Selected = "Fighter" // starter selection alone never counts as progress
	if !m.IsFresh() {
		t.Error("bare class selection must not count as a played save")
	}
}

func TestFlavorCardRenders(t *testing.T) {
	card := RollFlavor(5)
	text := card.Card()
	for _, want := range []string{card.Hometown, card.Boon} {
		if !strings.Contains(text, want) {
			t.Errorf("card text missing %q:\n%s", want, text)
		}
	}
	if len(text) < 40 {
		t.Errorf("card text too thin:\n%s", text)
	}
	// The sheet modal fits ~23 lines; the card block must stay a tight
	// 3 lines so the sheet never spills past its panel.
	if n := len(strings.Split(strings.TrimSpace(text), "\n")); n > 3 {
		t.Errorf("card runs %d lines, want 3:\n%s", n, text)
	}
}
