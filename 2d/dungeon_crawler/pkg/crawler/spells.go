package crawler

import (
	"fmt"
	"sort"
)

// Spell is one Wizard spell. The demo models five resolution shapes:
// "ray" (the existing spell attack down the corridor), "arc" (attack rolls
// against up to two nearby foes), "missiles" (auto-hitting bolts), and the
// two no-roll area blasts that catch every visible foe within their radius.
// Cantrips cast for free; leveled spells spend a slot of their SlotLevel,
// and slots refresh when the wizard descends — the dungeon's stand-in for
// PF2e's daily preparation.
type Spell struct {
	Name      string
	Cantrip   bool
	SlotLevel int // 0 for cantrips
	Actions   int
	Shape     string
	Blurb     string // one-line summary for the spellbook and the sheet
}

// wizardSpellTiers is what the grimoire gains at each character level.
var wizardSpellTiers = map[int][]Spell{
	1: {
		{Name: "Ray of Frost", Cantrip: true, Actions: 1, Shape: "ray",
			Blurb: "frost ray, reach 3 — 1d6 + level"},
		{Name: "Electric Arc", Cantrip: true, Actions: 1, Shape: "arc",
			Blurb: "arc to the 2 nearest foes within 2 — 1d4 + level each"},
	},
	2: {
		{Name: "Magic Missile", SlotLevel: 1, Actions: 1, Shape: "missiles",
			Blurb: "sure bolts — 1 + level/2 at 1d4+1 each"},
		{Name: "Knock", Cantrip: true, Actions: 1, Shape: "knock",
			Blurb: "swings one locked door within a step open — 1 action"},
	},
	3: {
		{Name: "Burning Hands", SlotLevel: 1, Actions: 1, Shape: "adjacent",
			Blurb: "2d6 fire on every adjacent foe"},
		{Name: "Fire Burst", SlotLevel: 2, Actions: 2, Shape: "burst2",
			Blurb: "4d6 fire to distance 2 — 2 actions"},
	},
}

// wizardSlotTotals is the slot budget per slot level at a character level:
// one 1st-level slot at level 2, and two 1st plus one 2nd from level 3 —
// PF2e's cadence of unlocking the next slot tier at levels 2 and 3, kept
// small for a one-dungeon run.
func wizardSlotTotals(level int) map[int]int {
	switch {
	case level >= 3:
		return map[int]int{1: 2, 2: 1}
	case level == 2:
		return map[int]int{1: 1}
	default:
		return map[int]int{}
	}
}

// syncWizardBook grants every spell the wizard's level has reached and
// raises remaining slots toward the new totals (never lowering them). It is
// idempotent and returns the newly learned spells so level-ups can log them.
// Non-wizards are left untouched.
func syncWizardBook(p *Player) []Spell {
	if p.Class != "Wizard" {
		return nil
	}
	if p.Slots == nil {
		p.Slots = map[int]int{}
	}
	var learned []Spell
	for lvl := 1; lvl <= p.Level; lvl++ {
		for _, s := range wizardSpellTiers[lvl] {
			known := false
			for _, have := range p.Spells {
				if have.Name == s.Name {
					known = true
					break
				}
			}
			if !known {
				p.Spells = append(p.Spells, s)
				learned = append(learned, s)
			}
		}
	}
	for sl, total := range wizardSlotTotals(p.Level) {
		if p.Slots[sl] < total {
			p.Slots[sl] = total
		}
	}
	return learned
}

// refreshWizardSlots refills every slot to its level total — the descend
// ritual that stands in for a night's re-preparation.
func refreshWizardSlots(p *Player) {
	if p.Class != "Wizard" || p.Slots == nil {
		return
	}
	for sl, total := range wizardSlotTotals(p.Level) {
		p.Slots[sl] = total
	}
}

// PreparedSpell returns the spell Space casts, falling back to the first
// entry when the index is unset.
func (p *Player) PreparedSpell() Spell {
	if p.Prepared >= 0 && p.Prepared < len(p.Spells) {
		return p.Spells[p.Prepared]
	}
	if len(p.Spells) > 0 {
		return p.Spells[0]
	}
	return Spell{}
}

// PrepareSpell selects which spell Space casts. It is free — the spell
// slots are the resource — the way sorting the pack is its own quiet verb.
func (r *Run) PrepareSpell(i int) []Event {
	p := r.Player
	if i < 0 || i >= len(p.Spells) {
		return []Event{{Kind: "log", Text: "You do not know that spell."}}
	}
	if i == p.Prepared {
		return []Event{{Kind: "log", Text: "Already prepared."}}
	}
	p.Prepared = i
	return []Event{{Kind: "log", Text: "You ready " + p.Spells[i].Name + "."}}
}

// castPreparedSpell resolves the Space key for the wizard: it dispatches
// the prepared spell's shape after checking actions and slots. Slots are
// only spent when the spell actually takes effect.
func (r *Run) castPreparedSpell() []Event {
	p := r.Player
	sp := p.PreparedSpell()
	if len(p.Spells) == 0 {
		return []Event{{Kind: "log", Text: "You have no spells."}}
	}
	if sp.SlotLevel > 0 && p.Slots[sp.SlotLevel] <= 0 {
		return []Event{{Kind: "log", Text: fmt.Sprintf("No %s-level slots remain — they refresh when you descend.", ordinal(sp.SlotLevel))}}
	}
	if p.ActionsLeft < sp.Actions {
		return []Event{{Kind: "log", Text: fmt.Sprintf("No actions left this turn — %s costs %d.", sp.Name, sp.Actions)}}
	}
	switch sp.Shape {
	case "ray":
		return r.castRayOfFrost(sp.Actions)
	case "arc":
		return r.castElectricArc(sp)
	case "missiles":
		return r.castMagicMissile(sp)
	case "knock":
		return r.castKnock(sp)
	case "adjacent":
		return r.castAreaSpell(sp, 1, Dice{2, 6}, "engulfs", "fire")
	case "burst2":
		return r.castAreaSpell(sp, 2, Dice{4, 6}, "detonates on", "fire")
	}
	return []Event{{Kind: "log", Text: "Unknown spell."}}
}

// spellAttack rolls one INT spell attack against an enemy and applies the
// dice, returning the resolved event plus any kill events.
func (r *Run) spellAttack(e *Enemy, d Dice, hitText, missText string) (Event, []Event) {
	p := r.Player
	roll := r.rollD20()
	res := RunCheck(roll, p.Level, true, p.AbilityMod("int"), e.AC, false)
	ev := Event{Kind: "attack", Roll: res.Roll, Total: res.Total, DC: res.DC,
		Degree: res.Degree, Subject: "enemy", At: e.Pos, EnemyID: e.ID}
	if res.Degree <= Failure {
		ev.Text = missText
		return ev, nil
	}
	if res.Degree == CriticalSuccess {
		d.N *= 2
	}
	dmg := rollDice(r.rng, d) + p.Level + p.AbilityMod("int")
	e.HP -= dmg
	ev.Damage = dmg
	ev.Text = fmt.Sprintf(hitText, e.Kind, dmg)
	if e.HP <= 0 {
		return ev, r.awardKill(e)
	}
	return ev, nil
}

// foesWithin returns live enemies within the given Manhattan distance
// (excluding the player's own tile), nearest first, capped at limit. No
// line-of-sight filter — area blasts like Burning Hands catch foes around
// corners the way they do in PF2e.
func (r *Run) foesWithin(dist, limit int) []*Enemy {
	p := r.Player
	type entry struct {
		e *Enemy
		d int
	}
	var list []entry
	for _, e := range r.Floor.Enemies {
		if e.HP <= 0 {
			continue
		}
		d := abs(e.Pos.X-p.Pos.X) + abs(e.Pos.Y-p.Pos.Y)
		if d >= 1 && d <= dist {
			list = append(list, entry{e, d})
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].d < list[j].d })
	if len(list) > limit {
		list = list[:limit]
	}
	out := make([]*Enemy, len(list))
	for i, en := range list {
		out[i] = en.e
	}
	return out
}

// foesWithinLos is the aimed-spell filter: same geometry, but walls block
// just like they do for the ray.
func (r *Run) foesWithinLos(dist, limit int) []*Enemy {
	var out []*Enemy
	for _, e := range r.foesWithin(dist, 100) {
		if lineClear(r.Floor, r.Player.Pos, e.Pos) {
			out = append(out, e)
			if len(out) == limit {
				break
			}
		}
	}
	return out
}

// castElectricArc lashes up to two nearby foes. Its attack rolls skip the
// multiple-attack bookkeeping — the slot-free cantrip trades MAP parity for
// breadth. The cantrip heightens to 2d4 at level 3. With no foes in reach
// the arc falls back to an armed trap and fuses it shut.
func (r *Run) castElectricArc(sp Spell) []Event {
	targets := r.foesWithinLos(2, 2)
	if len(targets) == 0 {
		traps := r.trapsWithin(2, true)
		if len(traps) == 0 {
			return []Event{{Kind: "log", Text: "No foes within the arc's reach."}}
		}
		r.Player.ActionsLeft -= sp.Actions
		r.Floor.SetFeature(traps[0], FeatureTrapDisabled)
		return append([]Event{{Kind: "log", At: traps[0],
			Text: "The Electric Arc fuses the trap's mechanism shut."}}, r.maybeEndTurn()...)
	}
	r.Player.ActionsLeft -= sp.Actions
	d := Dice{1, 4}
	if r.Player.Level >= 3 {
		d.N = 2
	}
	var evs []Event
	for _, e := range targets {
		ev, post := r.spellAttack(e, d,
			"Electric Arc lashes the %s for %d lightning.",
			"Electric Arc fizzes against the foe.")
		evs = append(evs, ev)
		evs = append(evs, post...)
	}
	return append(evs, r.maybeEndTurn()...)
}

// castMagicMissile never misses: PF2e's signature bolts strike the nearest
// visible foe for 1d4+1 force each, one extra bolt every two levels.
func (r *Run) castMagicMissile(sp Spell) []Event {
	targets := r.foesWithinLos(5, 1)
	if len(targets) == 0 {
		traps := r.trapsWithin(5, true)
		if len(traps) == 0 {
			return []Event{{Kind: "log", Text: "No target within the missiles' reach."}}
		}
		p := r.Player
		p.ActionsLeft -= sp.Actions
		p.Slots[sp.SlotLevel]--
		r.Floor.SetFeature(traps[0], FeatureTrapDisabled)
		bolts := 1 + p.Level/2
		return append([]Event{{Kind: "log", At: traps[0],
			Text: fmt.Sprintf("Magic Missile's %d bolts blast the trap apart.", bolts)}}, r.maybeEndTurn()...)
	}
	p := r.Player
	p.ActionsLeft -= sp.Actions
	p.Slots[sp.SlotLevel]--
	bolts := 1 + p.Level/2
	dmg := 0
	for i := 0; i < bolts; i++ {
		dmg += rollDice(r.rng, Dice{1, 4}) + 1
	}
	e := targets[0]
	e.HP -= dmg
	ev := Event{Kind: "attack", DC: e.AC, Degree: Success, Damage: dmg,
		Subject: "enemy", At: e.Pos, EnemyID: e.ID,
		Text: fmt.Sprintf("Magic Missile's %d bolts strike the %s for %d force.", bolts, e.Kind, dmg)}
	evs := []Event{ev}
	if e.HP <= 0 {
		evs = append(evs, r.awardKill(e)...)
	}
	return append(evs, r.maybeEndTurn()...)
}

// castKnock is the wizard's surgical cantrip: the locked door within a step
// swings open — no roll, no slot. With no door in reach it fizzles and
// commits nothing, exactly like the other cantrips missing a target.
func (r *Run) castKnock(sp Spell) []Event {
	for _, d := range [8]Cell{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		c := Cell{r.Player.Pos.X + d.X, r.Player.Pos.Y + d.Y}
		if r.Floor.InBounds(c) && r.Floor.FeatureAt(c) == FeatureLockedDoor {
			r.Player.ActionsLeft -= sp.Actions
			r.Floor.SetFeature(c, FeatureOpenDoor)
			return append([]Event{{Kind: "log",
				Text: "You knock — the lock bursts open."}}, r.maybeEndTurn()...)
		}
	}
	return []Event{{Kind: "log", Text: "Knock finds no locked door within a step."}}
}

// castAreaSpell is the no-roll blast template: every visible foe within
// reach takes the dice, and armed traps caught in the same radius burn or
// freeze out of the tile — the fire does not stop at tripwires. The slot
// is spent when anything landed.
func (r *Run) castAreaSpell(sp Spell, reach int, d Dice, verb, dtype string) []Event {
	targets := r.foesWithin(reach, 100)
	traps := r.trapsWithin(reach, false)
	if len(targets) == 0 && len(traps) == 0 {
		return []Event{{Kind: "log", Text: sp.Name + " blooms harmlessly — nothing in reach."}}
	}
	p := r.Player
	p.ActionsLeft -= sp.Actions
	if sp.SlotLevel > 0 {
		p.Slots[sp.SlotLevel]--
	}
	var evs []Event
	for _, e := range targets {
		dmg := rollDice(r.rng, d)
		e.HP -= dmg
		ev := Event{Kind: "attack", DC: e.AC, Degree: Success, Damage: dmg,
			Subject: "enemy", At: e.Pos, EnemyID: e.ID,
			Text: fmt.Sprintf("%s %s the %s for %d %s.", sp.Name, verb, e.Kind, dmg, dtype)}
		evs = append(evs, ev)
		if e.HP <= 0 {
			evs = append(evs, r.awardKill(e)...)
		}
	}
	for _, c := range traps {
		r.Floor.SetFeature(c, FeatureTrapDisabled)
		evs = append(evs, Event{Kind: "log", At: c,
			Text: fmt.Sprintf("%s guts the armed trap — its mechanism melts away.", sp.Name)})
	}
	return append(evs, r.maybeEndTurn()...)
}

// trapsWithin lists armed trap cells nearest-first under a spell's reach
// geometry: LOS-filtered for aimed shapes, plain reach for area blasts.
func (r *Run) trapsWithin(reach int, los bool) []Cell {
	p := r.Player
	type hit struct {
		c Cell
		d int
	}
	var found []hit
	for y := p.Pos.Y - reach; y <= p.Pos.Y+reach; y++ {
		for x := p.Pos.X - reach; x <= p.Pos.X+reach; x++ {
			c := Cell{X: x, Y: y}
			d := abs(x-p.Pos.X) + abs(y-p.Pos.Y)
			if d == 0 || d > reach || !r.Floor.InBounds(c) {
				continue
			}
			if r.Floor.FeatureAt(c) != FeatureTrapArmed {
				continue
			}
			if los && !lineClear(r.Floor, p.Pos, c) {
				continue
			}
			found = append(found, hit{c, d})
		}
	}
	sort.SliceStable(found, func(i, j int) bool { return found[i].d < found[j].d })
	cells := make([]Cell, len(found))
	for i, h := range found {
		cells[i] = h.c
	}
	return cells
}

// rayTrapFallback handles the ray's empty-target branch: its reach also
// touches an armed trap, which the frost shatters; with neither foe nor
// trap it reports the old fizzle.
func (r *Run) rayTrapFallback(reach, cost int) []Event {
	traps := r.trapsWithin(reach, true)
	if len(traps) == 0 {
		return []Event{{Kind: "log", Text: "No target within the ray's reach."}}
	}
	c := traps[0]
	r.Player.ActionsLeft -= cost
	r.Player.StrikesThisTurn++
	r.Floor.SetFeature(c, FeatureTrapDisabled)
	return append([]Event{{Kind: "log", At: c,
		Text: "The ray's frost cracks the trap's mechanism open."}}, r.maybeEndTurn()...)
}

// SlotSummary renders the remaining-spell-slots line used by the sheet and
// the spellbook, e.g. "slots: 1st 1/2, 2nd 1/1".
func (p *Player) SlotSummary() string {
	totals := wizardSlotTotals(p.Level)
	if len(totals) == 0 {
		return ""
	}
	levels := make([]int, 0, len(totals))
	for sl := range totals {
		levels = append(levels, sl)
	}
	sort.Ints(levels)
	out := "slots:"
	for _, sl := range levels {
		out += fmt.Sprintf(" %s-level %d/%d", ordinal(sl), p.Slots[sl], totals[sl])
	}
	return out
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "first"
	case 2:
		return "second"
	default:
		return fmt.Sprintf("%dth", n)
	}
}
