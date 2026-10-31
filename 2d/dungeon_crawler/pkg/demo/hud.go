package demo

import (
	"fmt"
	"strings"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// hud owns every HUD widget: the top status panel (title, HP/XP bars, action
// pips, hotkey footer), the renown pill, the log strip, the building panel,
// and the end screen. Every text surface has a semi-transparent ColorRect
// backing that toggles with its content.
//
// Geometry lives in Main.tscn; this file only sets text, colors, and bar
// widths (Control offsets against the scene's fixed left/top).
type hud struct {
	title     Label
	hpText    Label
	xpText    Label
	actions   Label
	hotkeys   Label
	renown    Label
	logLbl    Label
	panelLbl  Label
	centerLbl Label

	hpFill   ColorRect
	xpFill   ColorRect
	pips     [3]ColorRect
	menuBg   ColorRect
	centerBg ColorRect

	// Modal overlays: a shared dimming scrim plus the help, sheet, and
	// inventory panels. Each panel carries its own gold title child.
	scrim      ColorRect
	helpPanel  ColorRect
	helpTitle  Label
	helpLbl    Label
	sheetPanel ColorRect
	sheetTitle Label
	sheetLbl   Label
	invPanel   ColorRect
	invTitle   Label
	invLbl     Label
	spellPanel ColorRect
	spellTitle Label
	spellLbl   Label
	menuTitle  Label

	// runOnly widgets are hidden in town mode (no dungeon stats to show).
	runOnly []CanvasItem
}

// bar geometry mirrors Main.tscn (both bars start at x=66, are 200 px wide
// and 14 px tall).
const (
	barLeft   = float32(66)
	barWidth  = float32(200)
	barHeight = float32(14)
)

func findHud(from Node) hud {
	h := hud{
		title:     findLabel(from, "HUD/TitleLabel"),
		hpText:    findLabel(from, "HUD/HpText"),
		xpText:    findLabel(from, "HUD/XpText"),
		actions:   findLabel(from, "HUD/ActionsLabel"),
		hotkeys:   findLabel(from, "HUD/HotkeyLabel"),
		renown:    findLabel(from, "HUD/RenownLabel"),
		logLbl:    findLabel(from, "HUD/LogLabel"),
		panelLbl:  findLabel(from, "HUD/PanelLabel"),
		centerLbl: findLabel(from, "HUD/CenterLabel"),

		hpFill:   findColorRect(from, "HUD/HpBarFill"),
		xpFill:   findColorRect(from, "HUD/XpBarFill"),
		menuBg:   findColorRect(from, "HUD/MenuPanel"),
		centerBg: findColorRect(from, "HUD/CenterPanel"),
	}
	h.pips = [3]ColorRect{
		findColorRect(from, "HUD/Pip1"),
		findColorRect(from, "HUD/Pip2"),
		findColorRect(from, "HUD/Pip3"),
	}
	h.scrim = findColorRect(from, "HUD/Scrim")
	h.helpPanel = findColorRect(from, "HUD/HelpPanel")
	h.helpTitle = findLabel(from, "HUD/HelpPanel/HelpTitle")
	h.helpLbl = findLabel(from, "HUD/HelpLabel")
	h.sheetPanel = findColorRect(from, "HUD/SheetPanel")
	h.sheetTitle = findLabel(from, "HUD/SheetPanel/SheetTitle")
	h.sheetLbl = findLabel(from, "HUD/SheetLabel")
	h.invPanel = findColorRect(from, "HUD/InventoryPanel")
	h.invTitle = findLabel(from, "HUD/InventoryPanel/InventoryTitle")
	h.invLbl = findLabel(from, "HUD/InventoryLabel")
	h.spellPanel = findColorRect(from, "HUD/SpellPanel")
	h.spellTitle = findLabel(from, "HUD/SpellPanel/SpellTitle")
	h.spellLbl = findLabel(from, "HUD/SpellLabel")
	h.menuTitle = findLabel(from, "HUD/MenuPanel/MenuTitle")
	h.runOnly = []CanvasItem{
		findLabel(from, "HUD/HpCaption"), findColorRect(from, "HUD/HpBarBg"), h.hpFill, h.hpText,
		findLabel(from, "HUD/XpCaption"), findColorRect(from, "HUD/XpBarBg"), h.xpFill, h.xpText,
		h.pips[0], h.pips[1], h.pips[2], h.actions,
	}
	return h
}

// setRun shows dungeon stats for the active run.
func (h *hud) setRun(r *crawler.Run, renown int) {
	for _, it := range h.runOnly {
		it.Show()
	}
	p := r.Player
	setLabel(h.title, fmt.Sprintf("%s Lv %d — Floor %d", p.Class, p.Level, r.FloorNum))
	setLabel(h.hpText, fmt.Sprintf("%d/%d", p.HP, p.MaxHP))

	next := crawler.XPCumulative(p.Level + 1)
	xpRatio := float32(1)
	if next > 0 {
		xpRatio = float32(p.XP) / float32(next)
		setLabel(h.xpText, fmt.Sprintf("%d/%d", p.XP, next))
	} else {
		setLabel(h.xpText, "MAX")
	}
	h.setBar(h.hpFill, float32(p.HP)/float32(p.MaxHP))
	h.setBar(h.xpFill, xpRatio)
	h.setPips(p.ActionsLeft)

	setLabel(h.actions, fmt.Sprintf("%d of 3 actions · %d potion(s)", p.ActionsLeft, p.Potions))
	setLabel(h.hotkeys, "Space cast · E search · Q potion · I pack · B book · Enter end turn · WASD/arrows · Shift+dir ↗↘↙↖")
	setLabel(h.renown, fmt.Sprintf("Renown %d", renown))
}

// setTown shows the hub variant: title and town hotkeys, no combat stats.
func (h *hud) setTown(m *crawler.Meta) {
	for _, it := range h.runOnly {
		it.Hide()
	}
	setLabel(h.title, "Hollowstone — "+m.Selected)
	setLabel(h.hotkeys, "1-4 pick class · bump buildings · Enter buy · Esc close · I pack · walk into the red gate")
	setLabel(h.renown, fmt.Sprintf("Renown %d", m.Renown))
}

// setPanel shows the building/menu text with its backing; "" hides both. The
// first line becomes the gold title inside the bordered panel.
func (h *hud) setPanel(text string) {
	if text == "" {
		showNode(h.panelLbl, false)
		showNode(h.menuBg, false)
		return
	}
	head, body := splitHeader(text)
	setLabel(h.menuTitle, head)
	setLabel(h.panelLbl, body)
	showNode(h.menuBg, true)
	showNode(h.panelLbl, true)
}

// setEnd shows the win/lose screen with its backing; "" hides both.
func (h *hud) setEnd(text string) {
	if text == "" {
		showNode(h.centerLbl, false)
		showNode(h.centerBg, false)
		return
	}
	setLabel(h.centerLbl, text)
	showNode(h.centerBg, true)
	showNode(h.centerLbl, true)
}

func (h *hud) setLog(text string) {
	setLabel(h.logLbl, text)
}

// splitHeader peels the first line off modal text: it becomes the panel's
// gold title while the remainder fills the body label.
func splitHeader(text string) (string, string) {
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		return text[:i], strings.TrimLeft(text[i+1:], "\n")
	}
	return text, ""
}

// showModal paints one bordered overlay on the dimming scrim and hides the
// panels of every other modal.
func (h *hud) showModal(panel ColorRect, title Label, lbl Label, text string, others []CanvasItem) {
	head, body := splitHeader(text)
	setLabel(title, head)
	setLabel(lbl, body)
	showNode(h.scrim, true)
	showNode(panel, true)
	showNode(lbl, true)
	for _, it := range others {
		showNode(it, false)
	}
}

// showHelp opens the bindings modal over the dimming scrim.
func (h *hud) showHelp(text string) {
	h.showModal(h.helpPanel, h.helpTitle, h.helpLbl, text,
		[]CanvasItem{h.sheetPanel, h.sheetLbl, h.invPanel, h.invLbl, h.spellPanel, h.spellLbl})
}

// showSheet opens the character sheet modal over the dimming scrim.
func (h *hud) showSheet(text string) {
	h.showModal(h.sheetPanel, h.sheetTitle, h.sheetLbl, text,
		[]CanvasItem{h.helpPanel, h.helpLbl, h.invPanel, h.invLbl, h.spellPanel, h.spellLbl})
}

// showInventory opens the pack modal over the dimming scrim.
func (h *hud) showInventory(text string) {
	h.showModal(h.invPanel, h.invTitle, h.invLbl, text,
		[]CanvasItem{h.helpPanel, h.helpLbl, h.sheetPanel, h.sheetLbl, h.spellPanel, h.spellLbl})
}

// showSpellbook opens the grimoire modal over the dimming scrim.
func (h *hud) showSpellbook(text string) {
	h.showModal(h.spellPanel, h.spellTitle, h.spellLbl, text,
		[]CanvasItem{h.helpPanel, h.helpLbl, h.sheetPanel, h.sheetLbl, h.invPanel, h.invLbl})
}

// closeModal hides the scrim and every modal panel and label.
func (h *hud) closeModal() {
	showNode(h.scrim, false)
	for _, it := range []CanvasItem{h.helpPanel, h.helpLbl, h.sheetPanel, h.sheetLbl, h.invPanel, h.invLbl, h.spellPanel, h.spellLbl} {
		showNode(it, false)
	}
}

// setBar resizes a fill rectangle from the shared left edge to width×ratio.
// It resizes via SetSize because godot-go mis-marshals Control.set_offset
// against Godot 4.8 — every write lands as 0 — the same binding quirk that
// forced drawFilledRect off draw_rect. The fill nodes sit at the bar's left
// edge in Main.tscn with zero anchors, so size alone drives the fill.
func (h *hud) setBar(fill ColorRect, ratio float32) {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	fill.SetSize(NewVector2WithFloat32Float32(ratio*barWidth, barHeight), false)
}

// setPips lights the first n action pips amber, dims the rest.
func (h *hud) setPips(n int) {
	for i, pip := range h.pips {
		col := rgb{0.18, 0.18, 0.22}
		if i < n {
			col = rgb{0.95, 0.78, 0.25}
		}
		pip.SetColor(col.color())
	}
}
