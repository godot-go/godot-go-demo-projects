package demo

import (
	"fmt"
	"strings"
	"time"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// RegisterClassDungeonCrawler registers the demo's root node class.
func RegisterClassDungeonCrawler() {
	startWatchdog()
	ClassDBRegisterClass(NewDungeonCrawlerFromOwnerObject, []PropertyDescriptor{}, nil, func(t *DungeonCrawler) {
		ClassDBBindMethodVirtual(t, "V_DungeonCrawler_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_DungeonCrawler_Process", "_process", nil, nil)
		ClassDBBindMethodVirtual(t, "V_DungeonCrawler_Draw", "_draw", nil, nil)
	})
}

func NewDungeonCrawlerFromOwnerObject(owner *GodotObject) GDClass {
	obj := &DungeonCrawler{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// UnregisterClassDungeonCrawler unregisters the root node class.
func UnregisterClassDungeonCrawler() {
	ClassDBUnregisterClass[*DungeonCrawler]()
}

// Game modes cycled by the root node.
const (
	modeStart = iota
	modeTown
	modeRun
	modeEnd
)

// Window and tile sizes. The dungeon (41x21px) is centered in the band
// between the status panel (bottom 132) and the log strip (top 622); the
// town (21x14 @32px) is centered in the window.
const (
	winW, winH = 1280, 720
	townTile   = 32
	runTile    = 21

	// mapTop keeps the dungeon view clear of the status panel.
	mapTop = 135
)

// DungeonCrawler is the root of Main.tscn. It renders the current mode (town,
// dungeon run, end screen) and forwards input to the rules engine in package
// crawler. All rules live there; this type only paints and presses buttons.
type DungeonCrawler struct {
	Node2DImpl

	mode int
	meta *crawler.Meta
	town *crawler.Town
	run  *crawler.Run
	seed int64 // seed of the current/last run, shown on the end screen

	h hud

	// modal tracks the open pause overlay (help bindings or character
	// sheet); while set, _process swallows every other key.
	modal modalKind

	// Combat fx: floating damage numbers and receiver shake, advanced by
	// _process(delta) and painted over the dungeon view.
	now    float64
	popups []popup
	shakes map[string]fxSpan

	logLines []string
	savePath string

	// Start flow (startscreen.go): the visible page, its selections, the
	// character-generation scratchpad, and whether Continue has a save.
	page      startPage
	menuIdx   int
	rosterIdx int
	cardSeed  int64
	card      crawler.CharacterFlavor
	cardClass string
	saveOK    bool
}

func (c *DungeonCrawler) GetClassName() string { return "DungeonCrawler" }
func (c *DungeonCrawler) GetParentClassName() string {
	return "Node2D"
}

func (c *DungeonCrawler) V_DungeonCrawler_Ready() {
	watchdogBeat()
	c.h = findHud(c)

	c.savePath = globalizedPath("user://dungeon_crawler_save.json")
	m, err := crawler.LoadMeta(c.savePath)
	if err != nil {
		m = crawler.DefaultMeta()
	}
	c.meta = m
	c.seed = time.Now().UnixNano()
	c.saveOK = continuePlayable(c.savePath, c.meta)
	c.enterStart()
}

func (c *DungeonCrawler) V_DungeonCrawler_Process(delta float64) {
	watchdogBeat()
	c.advanceFx(delta)
	if actionTapped("help") {
		c.toggleModal(modalHelp)
		return
	}
	if actionTapped("sheet") {
		c.toggleModal(modalSheet)
		return
	}
	if actionTapped("inv") {
		c.toggleModal(modalInventory)
		return
	}
	if actionTapped("book") {
		c.toggleModal(modalSpellbook)
		return
	}
	if c.modal != modalNone {
		// The game is turn-based, so swallowing every verb while a modal
		// is up is a true pause: nothing can advance underneath. The pack
		// modal still accepts 1-4 to wield — equipping is the one verb it
		// offers, and it spends a rules action like any Interact.
		if c.modal == modalInventory && c.mode == modeRun {
			for i, act := range classKeys4 {
				if actionJustPressed(act) {
					c.push(c.run.EquipWeapon(i)...)
					c.checkRunEnd()
					c.refresh()
					if c.modal == modalInventory {
						c.h.showInventory(c.inventoryText())
					}
					return
				}
			}
		}
		// The spellbook offers preparing the Space cast — free, like
		// sorting the pack, so nothing else advances underneath. The book
		// outgrows the four class keys at level 3, so it binds class_5 too.
		if c.modal == modalSpellbook && c.mode == modeRun && c.run.Player.Class == "Wizard" {
			for i, act := range classKeys5 {
				if actionJustPressed(act) {
					c.push(c.run.PrepareSpell(i)...)
					c.refresh()
					if c.modal == modalSpellbook {
						c.h.showSpellbook(c.spellbookText())
					}
					return
				}
			}
		}
		if actionTapped("cancel") || actionTapped("confirm") {
			c.closeModal()
		}
		return
	}
	changed := false
	switch c.mode {
	case modeStart:
		changed = c.processStart()
	case modeTown:
		changed = c.processTown()
	case modeRun:
		changed = c.processRun()
	case modeEnd:
		if actionJustPressed("confirm") {
			c.enterTown("You return to Hollowstone.")
			changed = true
		}
	}
	if changed {
		c.refresh()
		c.QueueRedraw()
	}
}

// modalKind enumerates the pause overlays shown above the game.
type modalKind int

const (
	modalNone modalKind = iota
	modalHelp
	modalSheet
	modalInventory
	modalSpellbook
)

// toggleModal opens m, or closes it if it is already the open modal.
func (c *DungeonCrawler) toggleModal(m modalKind) {
	if c.modal == m {
		c.closeModal()
		return
	}
	c.modal = m
	switch m {
	case modalHelp:
		c.h.showHelp(helpText())
	case modalSheet:
		c.h.showSheet(c.sheetText())
	case modalInventory:
		c.h.showInventory(c.inventoryText())
	case modalSpellbook:
		c.h.showSpellbook(c.spellbookText())
	}
}

func (c *DungeonCrawler) closeModal() {
	c.modal = modalNone
	c.h.closeModal()
}

// sheetText builds the PF2e character sheet: the live player mid-run, or a
// level-1 preview of the adventurer the town would send down next. The
// rolled persona card from character generation rides along when stored.
func (c *DungeonCrawler) sheetText() string {
	var base string
	if c.run != nil && (c.mode == modeRun || c.mode == modeEnd) {
		p := c.run.Player
		base = crawler.CharacterSheet(p, c.run.Enhanced[p.Class])
	} else {
		p := crawler.ClassStatline(c.meta.Selected, 1)
		b := c.meta.RunBonuses()
		p.StrikeDamageBonus = b.StrikeDamageBonus
		p.WorldBonus = b.WorldBonus
		p.SearchBonus = b.SearchBonus
		p.Potions = b.Potions
		base = crawler.CharacterSheet(p, b.Enhanced[c.meta.Selected])
	}
	if c.meta != nil && c.meta.Flavor != nil {
		base += "\n— Character card —\n" + c.meta.Flavor.Card()
	}
	return base
}

// inventoryText builds the I modal: the live pack mid-run, or the kit a new
// run would start with while in town.
func (c *DungeonCrawler) inventoryText() string {
	class := c.meta.Selected
	potions := c.meta.RunBonuses().Potions
	var weapons []crawler.Weapon
	equipped := 0
	inRun := c.run != nil && (c.mode == modeRun || c.mode == modeEnd)
	if inRun {
		p := c.run.Player
		class = p.Class
		potions = p.Potions
		weapons = p.Weapons
		equipped = p.Equipped
	} else {
		weapons = []crawler.Weapon{crawler.StarterWeapons[class]}
	}

	var b strings.Builder
	if inRun {
		b.WriteString("PACK — INVENTORY\n\n")
	} else {
		b.WriteString("PACK — STARTING KIT\n\n")
	}
	fmt.Fprintf(&b, "Potions: %d (Q quaffs, heals 10)\n\nWeapons:\n", potions)
	trainedMartial := crawler.Classes[class].TrainedMartial
	for i, w := range weapons {
		mark := "  "
		if i == equipped {
			mark = "▸ "
		}
		line := fmt.Sprintf("%s%d. %-10s %dd%d %s", mark, i+1, w.Name, w.Dice.N, w.Dice.Size, w.DamageType)
		if w.Finesse {
			line += " · finesse"
		}
		if w.Deadly > 0 {
			line += fmt.Sprintf(" · deadly d%d", w.Deadly)
		}
		if w.Martial && !trainedMartial {
			line += " · untrained −2"
		}
		if i == equipped {
			line += "  (wielded)"
		}
		b.WriteString(line + "\n")
	}
	if inRun {
		b.WriteString("\n1-4 — wield a weapon (costs 1 action)")
	} else {
		b.WriteString("\nWeapons found in the dungeon are lost when you fall.")
	}
	b.WriteString("\nI or Esc — close")
	return b.String()
}

// spellbookText builds the B modal: the wizard's live grimoire mid-run, the
// starting lineup while in town, or a flavor empty state for other classes.
func (c *DungeonCrawler) spellbookText() string {
	class := c.meta.Selected
	var spells []crawler.Spell
	prepared := 0
	slots := ""
	inRun := c.run != nil && (c.mode == modeRun || c.mode == modeEnd)
	if inRun {
		p := c.run.Player
		class = p.Class
		spells = p.Spells
		prepared = p.Prepared
		slots = p.SlotSummary()
	} else if class == "Wizard" {
		spells = crawler.ClassStatline("Wizard", 1).Spells
	}

	if class != "Wizard" {
		return "SPELLBOOK — EMPTY\n\nOnly a Wizard coaxes arcana from a book.\nPick one at the Gate.\n\nB or Esc — close"
	}

	var b strings.Builder
	if inRun {
		b.WriteString("SPELLBOOK — GRIMOIRE\n\n")
	} else {
		b.WriteString("SPELLBOOK — STARTING GRIMOIRE\n\n")
	}
	if slots != "" {
		fmt.Fprintf(&b, "%s — they refresh when you descend\n\n", slots)
	}
	for i, s := range spells {
		mark := "  "
		if i == prepared {
			mark = "▸ "
		}
		kind := fmt.Sprintf("%d-level", s.SlotLevel)
		if s.Cantrip {
			kind = "cantrip"
		}
		line := fmt.Sprintf("%s%d. %-14s %s · %d action · %s", mark, i+1, s.Name, kind, s.Actions, s.Blurb)
		if i == prepared {
			line += "  (prepared)"
		}
		b.WriteString(line + "\n")
	}
	if inRun {
		fmt.Fprintf(&b, "\n1-%d — prepare a spell for Space (free)", len(spells))
	}
	b.WriteString("\nB or Esc — close")
	return b.String()
}

// helpText lists every hotkey binding for the ? modal.
func helpText() string {
	return `HOTKEY BINDINGS

WASD / Arrow keys — move · bump into things to attack or interact
Shift + a direction — diagonal step, tilted clockwise (↑↗ →↘ ↓↙ ←↖)
Space — class skill · the Wizard casts the prepared spell
E — search for hidden doors and traps
Q — quaff a healing potion
Enter — end turn · buy · confirm
Esc — close panel · dismiss these modals
1-4 — choose class (in town) · wield (in pack) · prepare (in book)
? or F1 — this help (pauses the game)
C — character sheet (pauses the game)
I — your pack: potions & weapons (pauses the game)
B — spellbook: ready a spell for Space (pauses the game)

The game pauses while any modal is up. Esc, Enter or re-pressing
the key dismisses it.`
}

func (c *DungeonCrawler) processTown() bool {
	if c.town.CurrentPanel().Building != "" {
		// The building panel is a modal like help and the sheet: the
		// wanderer waits while it is up. Enter buys, Esc dismisses, and
		// the gate roster keys still work.
		for i, act := range classKeys4 {
			if actionJustPressed(act) {
				c.selectClass(i)
				return true
			}
		}
		if actionJustPressed("confirm") || actionJustPressed("skill") {
			c.push(c.town.BuyOpen()...)
			return true
		}
		if actionTapped("cancel") {
			c.town.ClosePanel()
			return true
		}
		return false
	}
	if d, ok := directionFromInput(); ok {
		c.push(c.town.Move(d)...)
		if c.town.StartRequested {
			c.startRun()
		}
		return true
	}
	for i, act := range classKeys4 {
		if actionJustPressed(act) {
			c.selectClass(i)
			return true
		}
	}
	if actionJustPressed("confirm") || actionJustPressed("skill") {
		c.push(c.town.BuyOpen()...)
		return true
	}
	if actionJustPressed("cancel") {
		c.town.ClosePanel()
		return true
	}
	return false
}

func (c *DungeonCrawler) processRun() bool {
	if d, ok := directionFromInput(); ok {
		c.push(c.run.Move(d)...)
		c.checkRunEnd()
		return true
	}
	if actionJustPressed("skill") {
		def := crawler.Classes[c.run.Player.Class]
		c.push(c.run.UseSkill(def.CombatSkill)...)
		c.checkRunEnd()
		return true
	}
	if actionJustPressed("search") {
		c.push(c.run.Search()...)
		c.checkRunEnd()
		return true
	}
	if actionJustPressed("quaff") {
		c.push(c.run.Quaff()...)
		c.checkRunEnd()
		return true
	}
	if actionJustPressed("wait") {
		c.push(c.run.EndTurn()...)
		c.checkRunEnd()
		return true
	}
	return false
}

func (c *DungeonCrawler) selectClass(index int) {
	roster := []string{"Fighter", "Cleric", "Wizard", "Rogue"}
	if index >= len(roster) {
		return
	}
	class := roster[index]
	for _, unlocked := range c.meta.UnlockedClasses() {
		if unlocked == class {
			if err := c.meta.SelectClass(class); err == nil {
				c.pushText("You will descend as a " + class + ".")
			}
			return
		}
	}
	c.pushText(class + " is locked — buy its Guild upgrade first.")
}

func (c *DungeonCrawler) startRun() {
	c.town.StartRequested = false
	c.run = crawler.NewRun(c.seed, c.meta.Selected, ptrOf(c.meta.RunBonuses()))
	c.mode = modeRun
	c.logLines = nil
	c.pushText(fmt.Sprintf("You enter the dungeon as a %s (seed %d).", c.meta.Selected, c.seed))
	c.refresh()
}

// checkRunEnd detects terminal states and transitions to the end screen,
// banking renown into meta the moment the run resolves.
func (c *DungeonCrawler) checkRunEnd() {
	if c.run.State == crawler.Running {
		return
	}
	won := c.run.State == crawler.Won
	firstKill := false
	if won {
		firstKill = c.meta.RecordFirstKill(crawler.BossKind)
	}
	payout := c.run.Payout(firstKill)
	c.meta.Renown += payout
	if err := crawler.SaveMeta(c.savePath, c.meta); err != nil {
		c.pushText("Could not save renown: " + err.Error())
	}

	c.endRunScreen(won, payout, firstKill)
	c.mode = modeEnd
}

func (c *DungeonCrawler) endRunScreen(won bool, payout int, firstKill bool) {
	title := "YOU DIED"
	lines := []string{}
	if won {
		title = "VICTORY"
		lines = append(lines, "The Grim Warden falls — the dungeon is yours!")
	} else {
		lines = append(lines, fmt.Sprintf("You fell on floor %d.", c.run.FloorNum))
	}
	bonus := ""
	if firstKill {
		bonus = fmt.Sprintf(" (incl. +%d first-kill bonus)", crawler.BossFirstKillRenown)
	}
	lines = append(lines,
		fmt.Sprintf("Renown earned: %d%s", payout, bonus),
		fmt.Sprintf("Total renown: %d", c.meta.Renown),
		fmt.Sprintf("Run seed: %d", c.seed),
		"",
		"Enter — return to town",
	)
	c.closeModal() // the end screen owns the view from here
	c.setCenter(strings.Join(append([]string{title}, lines...), "\n"))
}

func (c *DungeonCrawler) enterTown(msg string) {
	c.mode = modeTown
	c.closeModal() // safety: never carry an overlay across a mode change
	c.run = nil
	c.town = crawler.NewTown(c.meta)
	c.town.ClosePanel()
	c.logLines = nil
	c.pushText(msg)
	c.seed++ // each run from town gets a fresh seed
	c.setCenter("")
}

// --- events and labels ---

func (c *DungeonCrawler) push(evs ...crawler.Event) {
	for _, ev := range evs {
		if ev.Text == "" {
			continue
		}
		c.logLines = append(c.logLines, ev.Text)
	}
	if n := len(c.logLines); n > 5 {
		c.logLines = c.logLines[n-5:]
	}
	c.spawnHitFx(evs)
}

func (c *DungeonCrawler) pushText(text string) {
	c.push(crawler.Event{Kind: "log", Text: text})
}

// refresh repaints the HUD from engine state via the hud widget set.
func (c *DungeonCrawler) refresh() {
	switch c.mode {
	case modeRun:
		c.h.setRun(c.run, c.meta.Renown)
	default:
		c.h.setTown(c.meta)
	}
	c.h.setLog(strings.Join(c.logLines, "\n"))

	if c.mode == modeTown && c.town.CurrentPanel().Building != "" {
		c.h.setPanel(c.panelText())
	} else {
		c.h.setPanel("")
	}
}

func (c *DungeonCrawler) panelText() string {
	p := c.town.CurrentPanel()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", strings.ToUpper(p.Building))
	if p.Building == "Gate" {
		b.WriteString(p.Effect + "\n")
		for i, opt := range p.Classes {
			mark := " "
			if opt.Selected {
				mark = ">"
			}
			enh := ""
			if opt.Enhanced {
				enh = " (enhanced)"
			}
			fmt.Fprintf(&b, "%s %d. %s%s\n", mark, i+1, opt.Class, enh)
		}
		return b.String()
	}
	b.WriteString(p.Effect + "\n")
	fmt.Fprintf(&b, "Level %d / %d\n", p.Level, p.Max)
	if p.Maxed {
		b.WriteString("Fully upgraded.\n")
	} else {
		afford := "not enough renown"
		if p.Affordable {
			afford = "affordable"
		}
		fmt.Fprintf(&b, "Next level: %d renown (%s; you have %d)\n", p.Cost, afford, c.meta.Renown)
		b.WriteString("[Enter] buy · [Esc] close\n")
	}
	return b.String()
}

func (c *DungeonCrawler) setCenter(text string) {
	c.h.setEnd(text)
}

// --- rendering ---

func (c *DungeonCrawler) V_DungeonCrawler_Draw() {
	paint(c, 0, 0, winW, winH, rgb{})
	switch c.mode {
	case modeRun:
		c.drawRun()
	case modeStart:
		// Black canvas: the menu panel carries the start pages.
	default:
		c.drawTown()
	}
}

func (c *DungeonCrawler) drawTown() {
	t := c.town
	ox := float32((winW - t.W*townTile) / 2)
	oy := float32((winH - t.H*townTile) / 2)
	paint(c, ox-4, oy-4, float32(t.W*townTile)+8, float32(t.H*townTile)+8, rgb{0.07, 0.09, 0.07})
	wear := townWear(t.Meta)
	for y := 0; y < t.H; y++ {
		for x := 0; x < t.W; x++ {
			b := t.BuildingAt(crawler.Cell{X: x, Y: y})
			gx, gy := ox+float32(x*townTile), oy+float32(y*townTile)
			if b != crawler.BldgNone {
				paint(c, gx+1, gy+1, townTile-2, townTile-2, buildingColor(b))
				continue
			}
			townGround(c, x, y, gx, gy, wear)
			if x == plazCell.X && y == plazCell.Y {
				townPlaza(c, gx, gy)
			}
			if (x == 10 && y >= 5 && y <= 11) || (y == 5 && x >= 2 && x <= 18) {
				townPathElement(c, x, y, gx, gy, wear)
			}
		}
	}
	drawTownProps(c, t, ox, oy, t.Meta, c.now)
	for b, r := range townBuildingRects(t) {
		drawTownFacade(c, b, r, ox, oy, t.Meta, c.now)
	}
	drawClassAvatar(c, c.meta.Selected,
		ox+float32(c.town.Pos.X*townTile), oy+float32(c.town.Pos.Y*townTile), townTile, 0.35, c.now, false)
}

func (c *DungeonCrawler) drawRun() {
	f := c.run.Floor
	ox := float32((winW - f.W*runTile) / 2)
	oy := float32(mapTop)
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			cell := crawler.Cell{X: x, Y: y}
			visible := c.run.Visible[cell]
			if !visible && !f.ExploredAt(cell) {
				continue
			}
			feature := f.FeatureAt(cell)
			if feature == crawler.FeatureHiddenDoor {
				feature = crawler.FeatureNone // secrets look like wall
			}
			gx, gy := ox+float32(x*runTile), oy+float32(y*runTile)
			switch {
			case f.IsSolid(cell):
				drawArtTile(c, dungeonWallRuns(cell), gx, gy, !visible)
			case feature != crawler.FeatureNone:
				col := featureColor(feature)
				if !visible {
					col = col.dimmed()
				}
				paint(c, gx+1, gy+1, runTile-2, runTile-2, col)
			default:
				drawArtTile(c, dungeonFloorRuns(cell), gx, gy, !visible)
			}
		}
	}
	// Items get a figure of their own over the tinted floor. Explored but
	// unseen memories keep the shape frozen (t=0) and dimmed.
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			cell := crawler.Cell{X: x, Y: y}
			item := f.FeatureAt(cell)
			if item == crawler.FeatureNone || item == crawler.FeatureHiddenDoor {
				continue
			}
			if item == crawler.FeatureTrapArmed && !f.TrapSpotted(cell) {
				continue // armed traps stay invisible until Search spots them
			}
			visible := c.run.Visible[cell]
			if !visible && !f.ExploredAt(cell) {
				continue
			}
			t := c.now
			if !visible {
				t = 0
			}
			drawItemAvatar(c, item, ox+float32(x*runTile), oy+float32(y*runTile), runTile,
				fxPhase(fmt.Sprintf("c%d,%d", x, y)), t, !visible)
		}
	}
	for _, e := range c.run.Floor.Enemies {
		if e.HP <= 0 || !c.run.Visible[e.Pos] {
			continue
		}
		hostileOutline(c, ox+float32(e.Pos.X*runTile), oy+float32(e.Pos.Y*runTile), runTile)
		pad := runTile / 4
		if e.IsBoss {
			pad = 2
		}
		sx, sy := c.shakeOffset(fmt.Sprintf("e%d", e.ID))
		drawEnemyAvatar(c, e.Kind,
			ox+float32(e.Pos.X*runTile)+float32(pad)+sx, oy+float32(e.Pos.Y*runTile)+float32(pad)+sy,
			float32(runTile-2*pad), fxPhase(fmt.Sprintf("e%d", e.ID)), c.now, false)
	}
	psx, psy := c.shakeOffset("player")
	drawClassAvatar(c, c.run.Player.Class,
		ox+float32(c.run.Player.Pos.X*runTile)+psx, oy+float32(c.run.Player.Pos.Y*runTile)+psy,
		float32(runTile), fxPhase("player"), c.now, false)
	c.drawPopups()
}

// drawArtTile blits a run list at tile scale: one fill per inked run, dimmed
// through the pen for explored-but-unseen memory. Runs paint in slice order,
// so hashed decor lands on top of the base pattern.
func drawArtTile(c *DungeonCrawler, runs []artRun, gx, gy float32, dim bool) {
	p := pen{c, dim}
	for _, r := range runs {
		p.fill(gx+float32(r.dx*runPx), gy+float32(r.dy*runPx),
			float32(r.w*runPx), float32(r.h*runPx), dungeonPalette[r.k])
	}
}

// --- colors ---

// rgb is a plain color triple; godot-go's Color is a raw byte blob, so we
// keep float math here and convert only when painting.
type rgb struct{ r, g, b float32 }

func (v rgb) dimmed() rgb { return rgb{v.r * 0.35, v.g * 0.35, v.b * 0.4} }

func (v rgb) color() Color {
	return NewColorWithFloat32Float32Float32Float32(v.r, v.g, v.b, 1)
}

func paint(c *DungeonCrawler, x, y, w, h float32, col rgb) {
	drawFilledRect(c, x, y, w, h, col.color())
}

// hostileOutline frames a visible enemy's tile with a thin red border:
// anything wearing it can hurt you right now. Memory ghosts never get one
// (the caller already filters to Visible cells).
func hostileOutline(c *DungeonCrawler, x, y, size float32) {
	const line = 1.5
	col := rgb{0.85, 0.25, 0.2}
	x, y, size = x+0.5, y+0.5, size-1
	paint(c, x, y, size, line, col)
	paint(c, x, y+size-line, size, line, col)
	paint(c, x, y+line, line, size-2*line, col)
	paint(c, x+size-line, y+line, line, size-2*line, col)
}

func buildingColor(b crawler.TownBuilding) rgb {
	switch b {
	case crawler.BldgSmithy:
		return rgb{0.62, 0.35, 0.15}
	case crawler.BldgClinic:
		return rgb{0.55, 0.72, 0.80}
	case crawler.BldgLibrary:
		return rgb{0.32, 0.55, 0.28}
	case crawler.BldgArchive:
		return rgb{0.55, 0.40, 0.72}
	case crawler.BldgGuildFighter:
		return rgb{0.72, 0.28, 0.22}
	case crawler.BldgGuildCleric:
		return rgb{0.80, 0.66, 0.28}
	case crawler.BldgGuildWizard:
		return rgb{0.32, 0.45, 0.75}
	case crawler.BldgGuildRogue:
		return rgb{0.28, 0.55, 0.48}
	case crawler.BldgGate:
		return rgb{0.65, 0.14, 0.14}
	}
	return rgb{0.52, 0.46, 0.2}
}

func featureColor(f crawler.Feature) rgb {
	switch f {
	case crawler.FeatureStairs:
		return rgb{0.2, 0.7, 0.75}
	case crawler.FeatureLockedDoor:
		return rgb{0.55, 0.38, 0.2}
	case crawler.FeatureOpenDoor:
		return rgb{0.42, 0.32, 0.2}
	case crawler.FeatureTrapArmed:
		return rgb{0.85, 0.75, 0.25}
	case crawler.FeatureTrapDisabled:
		return rgb{0.45, 0.42, 0.3}
	case crawler.FeaturePotion:
		return rgb{0.85, 0.35, 0.6}
	case crawler.FeatureWeapon:
		return rgb{0.55, 0.66, 0.85}
	}
	return rgb{}
}

func enemyColor(kind string) rgb {
	switch kind {
	case crawler.BossKind:
		return rgb{0.85, 0.18, 0.18}
	case "rat":
		return rgb{0.55, 0.42, 0.3}
	case "goblin":
		return rgb{0.3, 0.65, 0.3}
	case "skeleton":
		return rgb{0.85, 0.85, 0.9}
	case "orc":
		return rgb{0.2, 0.5, 0.22}
	case "shadow":
		return rgb{0.32, 0.22, 0.45}
	case "troll":
		return rgb{0.42, 0.55, 0.35}
	case "hag":
		return rgb{0.6, 0.3, 0.6}
	default: // wraith
		return rgb{0.55, 0.72, 0.82}
	}
}

// --- helpers ---

// directionFromInput reads the directional verbs: plain keys step
// orthogonally; with Shift held the pressed direction tilts one step
// clockwise onto its diagonal (↑↗ →↘ ↓↙ ←↖).
// Poll tables hoisted out of the per-frame handlers: the hot path allocates
// nothing steady-state, which keeps GC cycles (and gctrace output) rare.
var (
	classKeys4 = []string{"class_1", "class_2", "class_3", "class_4"}
	classKeys5 = []string{"class_1", "class_2", "class_3", "class_4", "class_5"}

	directionMoves = []struct {
		act string
		dir crawler.Dir
	}{
		{"move_up", crawler.Up},
		{"move_down", crawler.Down},
		{"move_left", crawler.Left},
		{"move_right", crawler.Right},
	}
)

func directionFromInput() (crawler.Dir, bool) {
	shift := shiftHeld()
	for _, m := range directionMoves {
		if directionJustPressed(m.act) {
			if !shift {
				return m.dir, true
			}
			return diagonalFor(m.dir), true
		}
	}
	return crawler.Up, false
}

// diagonalFor rotates a direction one step clockwise onto the diagonal it
// points at.
func diagonalFor(pressed crawler.Dir) crawler.Dir {
	switch pressed {
	case crawler.Up:
		return crawler.UpRight
	case crawler.Right:
		return crawler.DownRight
	case crawler.Down:
		return crawler.DownLeft
	case crawler.Left:
		return crawler.UpLeft
	}
	return pressed
}

func globalizedPath(path string) string {
	ps := GetProjectSettingsSingleton()
	s := NewStringWithLatin1Chars(path)
	defer s.Destroy()
	gs := ps.GlobalizePath(s)
	defer gs.Destroy()
	return gs.ToUtf8()
}

func ptrOf(b crawler.RunBonuses) *crawler.RunBonuses { return &b }
