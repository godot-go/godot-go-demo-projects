package demo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// --- 17.2 start-screen pages (pure text parts) ---

func TestStartMenuRows(t *testing.T) {
	text := startMenuText(0, true)
	for _, want := range []string{"New Game", "Continue", "Quit"} {
		if !strings.Contains(text, want) {
			t.Errorf("menu missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "▶ New Game") {
		t.Errorf("row 0 must carry the marker:\n%s", text)
	}
	if strings.Contains(text, "no save yet") {
		t.Errorf("playable continue must not advertise no save:\n%s", text)
	}

	text = startMenuText(1, false)
	if !strings.Contains(text, "▶ Continue") {
		t.Errorf("row 1 must carry the marker:\n%s", text)
	}
	if !strings.Contains(text, "no save yet") {
		t.Errorf("blocked continue must say so:\n%s", text)
	}
}

func TestStartConfirmNamesTheLoss(t *testing.T) {
	text := startConfirmText()
	for _, want := range []string{"renown", "Esc"} {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(want)) {
			t.Errorf("confirm missing %q:\n%s", want, text)
		}
	}
}

func TestStartRosterListsClassesWithBlurbs(t *testing.T) {
	classes := []string{"Fighter", "Wizard"}
	text := startRosterText(classes, 0)
	for _, class := range classes {
		if !strings.Contains(text, class) {
			t.Errorf("roster missing %s:\n%s", class, text)
		}
	}
	if !strings.Contains(text, "▶ Fighter") {
		t.Errorf("roster marker wrong:\n%s", text)
	}
	if blurb := crawler.ClassBlurb("Fighter"); !strings.Contains(text, blurb) {
		t.Errorf("Fighter row missing its blurb %q:\n%s", blurb, text)
	}
}

func TestStartCardPageShowsCard(t *testing.T) {
	card := crawler.RollFlavor(11)
	text := startCardText("Fighter", card)
	if !strings.Contains(text, "Fighter") || !strings.Contains(text, card.Hometown) {
		t.Errorf("card page incomplete:\n%s", text)
	}
	if !strings.Contains(text, "reroll") {
		t.Errorf("card page must advertise Esc reroll:\n%s", text)
	}
}

func TestContinuePlayable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	if continuePlayable(path, crawler.DefaultMeta()) {
		t.Error("missing save is not playable")
	}
	if err := crawler.SaveMeta(path, crawler.DefaultMeta()); err != nil {
		t.Fatal(err)
	}
	if continuePlayable(path, crawler.DefaultMeta()) {
		t.Error("fresh-install save is not playable either")
	}
	m := crawler.DefaultMeta()
	m.Renown = 40
	if err := crawler.SaveMeta(path, m); err != nil {
		t.Fatal(err)
	}
	if !continuePlayable(path, m) {
		t.Error("a save with banked progress must be playable")
	}
	os.Remove(path)
	if continuePlayable(path, m) {
		t.Error("deleting the file blocks continue again")
	}
}
