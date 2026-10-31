package demo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// The boot flow walks four text pages inside the menu panel: the main menu,
// a played-save confirm, the class roster, and the rolled flavor card. No
// town or run state exists until the card is accepted — the only exits are
// the accepted character, Continue with a banked save, or Quit.

type startPage int

const (
	pageMenu startPage = iota
	pageConfirm
	pageRoster
	pageCard
)

// startMenuRows is New Game, Continue, Quit.
const startMenuRows = 3

// startMenuText renders the three-entry menu; selected is 0-2 and a false
// continueOK greys the Continue row with its hint.
func startMenuText(selected int, continueOK bool) string {
	cont := "Continue"
	if !continueOK {
		cont += " — no save yet"
	}
	return startPageText("DUNGEON CRAWLER",
		[]string{"New Game", cont, "Quit"}, selected,
		"Up/Down — choose · Enter — select · 1-3 — jump")
}

// startConfirmText warns what a new game throws away.
func startConfirmText() string {
	return "START A NEW CHARACTER?\n\n" +
		"This discards the banked save: renown,\n" +
		"building levels, and first-kills.\n\n" +
		"Enter — reset and begin · Esc — back"
}

// startRosterText lists the unlocked classes with their blurbs.
func startRosterText(classes []string, selected int) string {
	rows := make([]string, len(classes))
	for i, class := range classes {
		rows[i] = fmt.Sprintf("%s — %s", class, crawler.ClassBlurb(class))
	}
	return startPageText("WHO DESCENDS?", rows, selected,
		"Up/Down — browse · Enter — confirm · Esc — back")
}

// startCardText shows the rolled flavor card awaiting acceptance.
func startCardText(class string, card crawler.CharacterFlavor) string {
	return "YOUR CHARACTER — " + class + "\n\n" +
		card.Card() +
		"\nEsc — reroll · Enter — accept and set out"
}

// startPageText composes a selectable page: title line, rows carrying the
// ▶ marker on the selection, and a key hint footer.
func startPageText(title string, rows []string, selected int, hint string) string {
	var b strings.Builder
	b.WriteString(title + "\n\n")
	for i, row := range rows {
		marker := "  "
		if i == selected {
			marker = "▶ "
		}
		b.WriteString(marker + row + "\n")
	}
	b.WriteString("\n" + hint)
	return b.String()
}

// continuePlayable reports whether Continue has something to load: the save
// file exists and holds progress beyond fresh-install defaults.
func continuePlayable(path string, m *crawler.Meta) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return !m.IsFresh()
}

// --- engine-side flow ---

// enterStart puts the menu page up over a black canvas.
func (c *DungeonCrawler) enterStart() {
	c.mode = modeStart
	c.page = pageMenu
	c.menuIdx = 0
	c.closeModal()
	c.logLines = nil
	c.h.setTown(c.meta)
	c.h.setLog("Hollowstone waits. Choose how to begin.")
	c.renderStart()
}

// renderStart paints the current start page into the menu panel. Start
// pages repaint themselves: the town/run refresh must never run here.
func (c *DungeonCrawler) renderStart() {
	switch c.page {
	case pageMenu:
		c.h.setPanel(startMenuText(c.menuIdx, c.saveOK))
	case pageConfirm:
		c.h.setPanel(startConfirmText())
	case pageRoster:
		c.h.setPanel(startRosterText(c.meta.UnlockedClasses(), c.rosterIdx))
	case pageCard:
		c.h.setPanel(startCardText(c.cardClass, c.card))
	}
	c.QueueRedraw()
}

// processStart drives the start pages. It returns false (pages repaint
// themselves via renderStart); a true return means the flow handed over to
// the town, which refresh() then owns.
func (c *DungeonCrawler) processStart() bool {
	switch c.page {
	case pageMenu:
		if actionJustPressed("move_up") {
			c.menuIdx = (c.menuIdx + startMenuRows - 1) % startMenuRows
		}
		if actionJustPressed("move_down") {
			c.menuIdx = (c.menuIdx + 1) % startMenuRows
		}
		for i, act := range []string{"class_1", "class_2", "class_3"} {
			if actionJustPressed(act) {
				c.menuIdx = i
			}
		}
		if actionTapped("confirm") {
			c.activateMenu()
		}
	case pageConfirm:
		if actionTapped("confirm") {
			// Memory resets now; the file survives until the new
			// character is accepted and overwrites it.
			c.meta = crawler.DefaultMeta()
			c.h.setTown(c.meta)
			c.h.setLog("The old life is dust. Begin again.")
			c.beginRoster()
		}
		if actionTapped("cancel") {
			// Backing out restores exactly what the disk still holds.
			if m, err := crawler.LoadMeta(c.savePath); err == nil {
				c.meta = m
				c.h.setTown(c.meta)
			}
			c.page = pageMenu
		}
	case pageRoster:
		roster := c.meta.UnlockedClasses()
		if actionJustPressed("move_up") && len(roster) > 0 {
			c.rosterIdx = (c.rosterIdx + len(roster) - 1) % len(roster)
		}
		if actionJustPressed("move_down") && len(roster) > 0 {
			c.rosterIdx = (c.rosterIdx + 1) % len(roster)
		}
		for i, act := range []string{"class_1", "class_2", "class_3", "class_4"} {
			if actionJustPressed(act) && i < len(roster) {
				c.rosterIdx = i
			}
		}
		if actionTapped("cancel") {
			c.page = pageMenu
		}
		if actionTapped("confirm") {
			c.rollCard()
		}
	case pageCard:
		if actionTapped("cancel") {
			c.rollCard()
		}
		if actionTapped("confirm") {
			c.acceptCharacter()
		}
	}
	if c.mode != modeStart {
		return true // Continue or accept handed the view to the town
	}
	c.renderStart()
	return false
}

// activateMenu runs the highlighted menu row.
func (c *DungeonCrawler) activateMenu() {
	switch c.menuIdx {
	case 0:
		if c.saveOK {
			c.page = pageConfirm
		} else {
			c.beginRoster()
		}
	case 1:
		if c.saveOK {
			c.enterTown("Welcome back to Hollowstone. The gate stands ready.")
		}
	case 2:
		c.GetTree().Quit(0)
	}
}

func (c *DungeonCrawler) beginRoster() {
	c.rosterIdx = 0
	c.page = pageRoster
}

// rollCard fixes the highlighted class and rolls a fresh persona card;
// calling it from the card page rerolls with a new seed.
func (c *DungeonCrawler) rollCard() {
	roster := c.meta.UnlockedClasses()
	if len(roster) == 0 {
		return // defensive: the starter class is always unlocked
	}
	if c.rosterIdx >= len(roster) {
		c.rosterIdx = len(roster) - 1
	}
	c.cardClass = roster[c.rosterIdx]
	c.cardSeed = time.Now().UnixNano()
	c.card = crawler.RollFlavor(c.cardSeed)
	c.page = pageCard
}

// acceptCharacter commits the class and flavor, banks the save, and opens
// the town — the flow's only productive exit.
func (c *DungeonCrawler) acceptCharacter() {
	if c.cardClass == "" {
		c.page = pageMenu
		return
	}
	if err := c.meta.SelectClass(c.cardClass); err != nil {
		c.pushText(err.Error())
		c.page = pageRoster
		return
	}
	card := c.card
	c.meta.Flavor = &card
	c.enterTown("Welcome to Hollowstone. Bump a building to inspect it; the red gate descends.")
	if err := crawler.SaveMeta(c.savePath, c.meta); err != nil {
		c.pushText("Could not bank the new character: " + err.Error())
	}
}
