package crawler

import (
	"fmt"
	"strings"
)

// Town is the walkable hub between runs. Bumping a building opens its panel,
// bumping the gate starts a run; there is no action economy here.

// TownBuilding is what occupies a town tile.
type TownBuilding int

const (
	BldgNone TownBuilding = iota
	BldgSmithy
	BldgClinic
	BldgLibrary
	BldgArchive
	BldgGuildFighter
	BldgGuildCleric
	BldgGuildWizard
	BldgGuildRogue
	BldgGate
)

// name is the meta save key for the building ("" for the gate and empty).
func (b TownBuilding) name() string {
	switch b {
	case BldgSmithy:
		return "Smithy"
	case BldgClinic:
		return "Clinic"
	case BldgLibrary:
		return "Library"
	case BldgArchive:
		return "Archive"
	case BldgGuildFighter:
		return "Guild Fighter"
	case BldgGuildCleric:
		return "Guild Cleric"
	case BldgGuildWizard:
		return "Guild Wizard"
	case BldgGuildRogue:
		return "Guild Rogue"
	case BldgGate:
		return "Gate"
	}
	return ""
}

// upgradeDef bounds a purchasable building: level cap and base cost. The
// cost of raising from level L is Base × (L+1), so costs rise per level.
type upgradeDef struct{ Max, Base int }

var upgrades = map[string]upgradeDef{
	"Smithy":        {Max: 3, Base: 50},
	"Clinic":        {Max: 3, Base: 40},
	"Library":       {Max: 3, Base: 60},
	"Archive":       {Max: 3, Base: 60},
	"Guild Fighter": {Max: 2, Base: 75},
	"Guild Cleric":  {Max: 2, Base: 75},
	"Guild Wizard":  {Max: 2, Base: 75},
	"Guild Rogue":   {Max: 2, Base: 75},
}

const starterClass = "Fighter"

// UpgradeTotal sums bought levels against the sum of caps — the town's
// overall kept-ness, which the shell paints as mud-to-grass ground wear.
func (m *Meta) UpgradeTotal() (have, max int) {
	for name, def := range upgrades {
		have += m.Buildings[name]
		max += def.Max
	}
	return
}

// classOrder fixes the gate roster order for stable panels.
var classOrder = []string{"Fighter", "Cleric", "Wizard", "Rogue"}

// guildName maps a class to the Guild building that gates it.
func guildName(class string) string { return "Guild " + class }

// UnlockedClasses returns the roster offered at the gate: the starter plus
// every class whose Guild has reached level 1, in fixed order.
func (m *Meta) UnlockedClasses() []string {
	var out []string
	for _, c := range classOrder {
		if c == starterClass || m.Buildings[guildName(c)] >= 1 {
			out = append(out, c)
		}
	}
	return out
}

// SelectClass picks the class the next run will use; locked classes are
// refused.
func (m *Meta) SelectClass(class string) error {
	for _, c := range m.UnlockedClasses() {
		if c == class {
			m.Selected = class
			return nil
		}
	}
	return fmt.Errorf("%s is locked (its Guild is level 0)", class)
}

// UpgradeCost is the renown price of the next level, or -1 at the cap.
func (m *Meta) UpgradeCost(name string) int {
	def, ok := upgrades[name]
	if !ok {
		return -1
	}
	lvl := m.Buildings[name]
	if lvl >= def.Max {
		return -1
	}
	return def.Base * (lvl + 1)
}

// Buy raises a building one level when below its cap and affordable.
func (m *Meta) Buy(name string) (bool, string) {
	def, ok := upgrades[name]
	if !ok {
		return false, "Nothing to upgrade there."
	}
	lvl := m.Buildings[name]
	if lvl >= def.Max {
		return false, fmt.Sprintf("The %s is fully upgraded.", name)
	}
	cost := def.Base * (lvl + 1)
	if m.Renown < cost {
		return false, fmt.Sprintf("Not enough renown: level %d costs %d (you have %d).", lvl+1, cost, m.Renown)
	}
	m.Renown -= cost
	m.Buildings[name] = lvl + 1
	return true, fmt.Sprintf("%s upgraded to level %d (-%d renown).", name, lvl+1, cost)
}

// RunBonuses snapshots the town's run-start effects. The result is copied
// into the Run, so town changes mid-run never affect an active attempt.
func (m *Meta) RunBonuses() RunBonuses {
	enhanced := map[string]bool{}
	for _, c := range classOrder {
		enhanced[c] = m.Buildings[guildName(c)] >= 2
	}
	return RunBonuses{
		StrikeDamageBonus: m.Buildings["Smithy"],
		WorldBonus:        m.Buildings["Library"],
		SearchBonus:       m.Buildings["Archive"],
		Potions:           m.Buildings["Clinic"],
		Enhanced:          enhanced,
	}
}

// ClassOption is one gate roster entry.
type ClassOption struct {
	Class    string
	Selected bool
	Enhanced bool
}

// Panel describes an open building panel for the shell to render.
type Panel struct {
	Building   string
	Level      int
	Max        int
	Cost       int // renown for next level; 0 when maxed
	Affordable bool
	Maxed      bool
	Effect     string
	Classes    []ClassOption // gate only
}

// Town is the hub map: open ground studded with building blocks.
type Town struct {
	W, H  int
	bld   []TownBuilding
	Pos   Cell
	Meta  *Meta
	panel TownBuilding
	// StartRequested is set when the gate is bumped with a selected class;
	// the shell reacts by creating the run and re-seating town.
	StartRequested bool
}

// NewTown builds the fixed town layout and seats the wanderer in the square.
func NewTown(m *Meta) *Town {
	t := &Town{W: 21, H: 14, Meta: m}
	t.bld = make([]TownBuilding, t.W*t.H)

	// Trade row along the top.
	t.stamp(BldgSmithy, 2, 1, 3, 2)
	t.stamp(BldgClinic, 5, 1, 6, 2)
	t.stamp(BldgGuildFighter, 8, 1, 9, 2)
	t.stamp(BldgGuildCleric, 11, 1, 12, 2)
	t.stamp(BldgGuildWizard, 14, 1, 15, 2)
	t.stamp(BldgGuildRogue, 17, 1, 18, 2)

	// Knowledge row.
	t.stamp(BldgLibrary, 3, 6, 4, 7)
	t.stamp(BldgArchive, 8, 6, 9, 7)

	// Gate at the bottom edge.
	t.stamp(BldgGate, 9, 12, 11, 13)

	t.Pos = Cell{10, 4}
	return t
}

func (t *Town) stamp(b TownBuilding, x0, y0, x1, y1 int) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t.bld[y*t.W+x] = b
		}
	}
}

// BuildingAt returns the building occupying c (BldgNone for open ground).
func (t *Town) BuildingAt(c Cell) TownBuilding {
	if c.X < 0 || c.X >= t.W || c.Y < 0 || c.Y >= t.H {
		return BldgNone
	}
	return t.bld[c.Y*t.W+c.X]
}

// BuildingTile returns the leftmost tile of a named building, for tests and
// the shell's "how to find it" help text.
func (t *Town) BuildingTile(name string) (Cell, bool) {
	for y := 0; y < t.H; y++ {
		for x := 0; x < t.W; x++ {
			if t.bld[y*t.W+x] != BldgNone && t.bld[y*t.W+x].name() == name {
				return Cell{x, y}, true
			}
		}
	}
	return Cell{}, false
}

// Move walks one tile; bumping a building opens its panel and does not move
// the wanderer. Bumping the gate with a selected class requests a run.
func (t *Town) Move(d Dir) []Event {
	delta := d.Delta()
	target := Cell{t.Pos.X + delta.X, t.Pos.Y + delta.Y}
	if target.X < 0 || target.X >= t.W || target.Y < 0 || target.Y >= t.H {
		return nil
	}

	switch b := t.bld[target.Y*t.W+target.X]; b {
	case BldgNone:
		t.Pos = target
		return nil
	case BldgGate:
		if t.Meta.Selected == "" {
			return []Event{{Kind: "town", Text: "The gatekeeper asks your class — choose one with 1-4."}}
		}
		t.StartRequested = true
		return []Event{{Kind: "start", Text: fmt.Sprintf("You descend as a %s.", t.Meta.Selected)}}
	default:
		t.panel = b
		return []Event{{Kind: "panel", Text: fmt.Sprintf("— %s —", b.name())}}
	}
}

// ClosePanel dismisses an open building panel.
func (t *Town) ClosePanel() { t.panel = BldgNone }

// CurrentPanel renders the open building, or a zero Panel when closed.
func (t *Town) CurrentPanel() Panel {
	if t.panel == BldgNone {
		return Panel{}
	}
	if t.panel == BldgGate {
		p := Panel{Building: "Gate", Effect: "Choose a class (1-4), walk in to descend."}
		for _, c := range t.Meta.UnlockedClasses() {
			p.Classes = append(p.Classes, ClassOption{
				Class:    c,
				Selected: t.Meta.Selected == c,
				Enhanced: t.Meta.Buildings[guildName(c)] >= 2,
			})
		}
		return p
	}
	return t.Meta.PanelFor(t.panel.name())
}

// PanelFor describes a purchasable building.
func (m *Meta) PanelFor(name string) Panel {
	def := upgrades[name]
	lvl := m.Buildings[name]
	cost := m.UpgradeCost(name)
	return Panel{
		Building:   name,
		Level:      lvl,
		Max:        def.Max,
		Cost:       cost,
		Affordable: cost >= 0 && m.Renown >= cost,
		Maxed:      cost < 0,
		Effect:     buildingEffect(name),
	}
}

// BuyOpen attempts to upgrade whichever panel is currently open.
func (t *Town) BuyOpen() []Event {
	if t.panel == BldgNone || t.panel == BldgGate {
		return nil
	}
	_, msg := t.Meta.Buy(t.panel.name())
	return []Event{{Kind: "town", Text: msg}}
}

// buildingEffect is the panel's one-line promise of what a level buys.
func buildingEffect(name string) string {
	switch name {
	case "Smithy":
		return "Strikes deal +1 damage per level."
	case "Clinic":
		return "Start each run with one potion per level."
	case "Library":
		return "+1 to world skill checks per level."
	case "Archive":
		return "Search reaches 1 tile further per level."
	}
	if class := strings.TrimPrefix(name, "Guild "); class != name {
		return "Unlocks " + class + "; level 2 enhances its combat skill."
	}
	return ""
}
