package crawler

import "fmt"

// Move is the single directional verb: it Strides into an empty tile, Strikes
// by bumping an enemy, and bumps interactable features. Moving into a wall or
// a blocked feature costs nothing.
func (r *Run) Move(d Dir) []Event {
	defer r.updateVisibility()
	if r.State != Running {
		return []Event{{Kind: "log", Text: "The run is over."}}
	}
	p := r.Player
	if p.ActionsLeft <= 0 {
		return []Event{{Kind: "log", Text: "No actions left this turn."}}
	}
	target := Cell{p.Pos.X + d.Delta().X, p.Pos.Y + d.Delta().Y}

	if e := r.Floor.EnemyAt(target); e != nil {
		return r.strikeAction(e, nil, "Strike", 1, 0)
	}
	if !r.Floor.TileWalkable(target) {
		return r.bumpFeature(target)
	}
	if d.isDiagonal() && r.cornerClipped(p.Pos, d) {
		return nil // no squeezing through wall corners; costs nothing
	}
	if r.Floor.FeatureAt(target) == FeatureTrapArmed {
		if evs := r.stepIntoTrapAttempt(target); evs != nil {
			return evs
		}
	}

	p.Pos = target
	p.ActionsLeft--
	evs := []Event{{Kind: "log", Text: "You stride."}}
	evs = append(evs, r.arriveOnTile()...)
	return append(evs, r.maybeEndTurn()...)
}

// cornerClipped reports whether a diagonal stride from pos along d would cut
// the corner between two unwalkable tiles — the two orthogonal neighbours on
// that corner. At least one open side is enough to slip through.
func (r *Run) cornerClipped(pos Cell, d Dir) bool {
	dx, dy := d.Delta().X, d.Delta().Y
	sideA := Cell{pos.X + dx, pos.Y}
	sideB := Cell{pos.X, pos.Y + dy}
	return !r.Floor.TileWalkable(sideA) && !r.Floor.TileWalkable(sideB)
}

// strikeAction resolves one melee Strike (optionally the Fighter's Power
// Attack with diceOverride and cost 2).
func (r *Run) strikeAction(e *Enemy, dice *Dice, label string, cost int, flatBonus int) []Event {
	p := r.Player
	pen := mapPenalty(p.StrikesThisTurn)
	roll := r.rollD20()
	res := RunCheck(roll, p.Level, true, p.StrikeMeleeMod()+p.WeaponProfMod()+pen, e.AC, false)
	p.StrikesThisTurn++

	ev := Event{
		Kind: "attack", Roll: res.Roll, Total: res.Total, DC: res.DC,
		Penalty: pen, Degree: res.Degree, Subject: "enemy", At: e.Pos, EnemyID: e.ID,
	}

	if res.Degree <= Failure {
		ev.Text = fmt.Sprintf("%s: d20=%d %+d = %d vs AC %d — miss", label, res.Roll, res.Total-res.Roll, res.Total, e.AC)
		p.ActionsLeft -= cost
		return append([]Event{ev}, r.maybeEndTurn()...)
	}

	d := p.StrikeDamage()
	if dice != nil {
		d = *dice
	}
	if res.Degree == CriticalSuccess {
		d.N *= 2
	}
	dmg := rollDice(r.rng, d) + p.StrikeMeleeMod() + p.StrikeDamageBonus + flatBonus
	if res.Degree == CriticalSuccess {
		if w := p.EquippedWeapon(); w.Deadly > 0 {
			dmg += rollDice(r.rng, Dice{1, w.Deadly})
		}
	}

	// Rogue Sneak Attack: precision dice against an enemy that has not
	// wounded anyone since the last enemy phase.
	if sneak := SneakAttackDice(p, r.Enhanced["Rogue"]); sneak > 0 && !e.AttackedRecently {
		sneakDmg := rollDice(r.rng, Dice{sneak, 6})
		dmg += sneakDmg
		ev.Text = fmt.Sprintf("Sneak attack! +%dd6 (%d) ", sneak, sneakDmg)
	}

	e.HP -= dmg
	ev.Damage = dmg
	verb := "hits"
	if res.Degree == CriticalSuccess {
		verb = "CRITS"
	}
	if ev.Text != "" { // sneak prefix
		ev.Text += fmt.Sprintf("%s: %s %s for %d", label, e.Kind, verb, dmg)
	} else {
		ev.Text = fmt.Sprintf("%s: d20=%d %+d = %d vs AC %d — %s %s for %d",
			label, res.Roll, res.Total-res.Roll, res.Total, e.AC, verb, e.Kind, dmg)
	}

	var post []Event
	if e.HP <= 0 {
		post = r.awardKill(e)
	}
	p.ActionsLeft -= cost
	return append(append([]Event{ev}, post...), r.maybeEndTurn()...)
}

// UseSkill activates the class combat skill.
func (r *Run) UseSkill(skill string) []Event {
	defer r.updateVisibility()
	if r.State != Running {
		return []Event{{Kind: "log", Text: "The run is over."}}
	}
	p := r.Player
	cost := ClassSkillCost(p.Class, skill)
	switch {
	case cost < 0:
		return []Event{{Kind: "log", Text: "Your class does not know " + skill + "."}}
	case cost == 0:
		return []Event{{Kind: "log", Text: "Sneak Attack rides your Strikes — it is not activated."}}
	case p.ActionsLeft < cost:
		return []Event{{Kind: "log", Text: "Not enough actions."}}
	}

	switch skill {
	case "power_attack":
		adj := r.adjacentEnemy()
		if adj == nil {
			return []Event{{Kind: "log", Text: "Power Attack needs an adjacent foe."}}
		}
		dice := PowerAttackDamage(p)
		bonus := 0
		if r.Enhanced["Fighter"] {
			bonus = 4 // Guild level-2 enhancement
		}
		return r.strikeAction(adj, &dice, "Power Attack", 2, bonus)
	case "lay_on_hands":
		if p.HP >= p.MaxHP {
			return []Event{{Kind: "log", Text: "You are already whole."}}
		}
		heal := LayOnHandsHeal(p, r.Enhanced["Cleric"])
		before := p.HP
		p.HP += heal
		if p.HP > p.MaxHP {
			p.HP = p.MaxHP
		}
		p.ActionsLeft -= cost
		ev := Event{Kind: "log", Text: fmt.Sprintf("Lay on Hands restores %d HP.", p.HP-before)}
		return append([]Event{ev}, r.maybeEndTurn()...)
	case "ray":
		return r.castPreparedSpell()
	}
	return []Event{{Kind: "log", Text: "Unknown skill."}}
}

// castRayOfFrost resolves the wizard's signature cantrip: an INT spell
// attack straight down the corridor, heightened by the Library.
func (r *Run) castRayOfFrost(cost int) []Event {
	p := r.Player
	target := r.rayTarget(RayReach(p, r.Enhanced["Wizard"]))
	if target == nil {
		return r.rayTrapFallback(RayReach(p, r.Enhanced["Wizard"]), cost)
	}
	roll := r.rollD20()
	res := RunCheck(roll, p.Level, true, p.AbilityMod("int"), target.AC, false)
	p.StrikesThisTurn++
	ev := Event{Kind: "attack", Roll: res.Roll, Total: res.Total, DC: res.DC, Degree: res.Degree, Subject: "enemy", At: target.Pos, EnemyID: target.ID}
	if res.Degree <= Failure {
		ev.Text = fmt.Sprintf("Ray: d20=%d %+d = %d vs AC %d — miss", res.Roll, res.Total-res.Roll, res.Total, target.AC)
		p.ActionsLeft -= cost
		return append([]Event{ev}, r.maybeEndTurn()...)
	}
	d := RayDamage(p, r.Enhanced["Wizard"])
	if res.Degree == CriticalSuccess {
		d.N *= 2
	}
	dmg := rollDice(r.rng, d) + p.Level + p.AbilityMod("int")
	target.HP -= dmg
	ev.Damage = dmg
	ev.Text = fmt.Sprintf("Ray sears the %s for %d", target.Kind, dmg)
	var post []Event
	if target.HP <= 0 {
		post = r.awardKill(target)
	}
	p.ActionsLeft -= cost
	return append(append([]Event{ev}, post...), r.maybeEndTurn()...)
}

// EndTurn forfeits remaining actions and hands over to the enemy phase.
func (r *Run) EndTurn() []Event {
	defer r.updateVisibility()
	if r.State != Running {
		return []Event{{Kind: "log", Text: "The run is over."}}
	}
	return r.finishPlayerTurn()
}

func (r *Run) finishPlayerTurn() []Event {
	evs := r.enemyPhase()
	r.Player.resetTurn()
	return append(evs, yourTurnMarker())
}

func (r *Run) maybeEndTurn() []Event {
	if r.State == Running && r.Player.ActionsLeft <= 0 {
		return r.finishPlayerTurn()
	}
	return nil
}

// awardKill grants XP, tracks renown, levels the character across PF2e
// thresholds, and ends the run when the boss falls.
func (r *Run) awardKill(e *Enemy) []Event {
	p := r.Player
	r.kills++
	p.XP += e.XP
	r.renownEarned += RenownPerKill(e.Level)
	evs := []Event{{Kind: "log", Text: fmt.Sprintf("The %s dies. +%d XP.", e.Kind, e.XP)}}

	for p.Level < len(xpCumulative) && p.XP >= xpCumulative[p.Level] {
		def := p.classOf()
		p.Level++
		newMax := def.HitBase + def.HitPerLevel*(p.Level-1) + def.Con*p.Level
		p.HP += newMax - p.MaxHP
		p.MaxHP = newMax
		p.AC = 12 + p.Level + def.Dex
		evs = append(evs, Event{Kind: "level_up", Text: fmt.Sprintf("You reach level %d!", p.Level)})
		for _, s := range syncWizardBook(p) {
			evs = append(evs, Event{Kind: "level_up", Text: "You learn " + s.Name + "!"})
		}
	}

	if e.IsBoss {
		r.bossDefeated = true
		r.State = Won
		evs = append(evs, Event{Kind: "win", Text: "The Grim Warden falls. The dungeon is yours!"})
	}
	r.removeEnemy(e)
	return evs
}

func (r *Run) removeEnemy(e *Enemy) {
	kept := r.Floor.Enemies[:0]
	for _, x := range r.Floor.Enemies {
		if x != e {
			kept = append(kept, x)
		}
	}
	r.Floor.Enemies = kept
}

// adjacentEnemy returns the first enemy orthogonally adjacent to the player.
func (r *Run) adjacentEnemy() *Enemy {
	for _, d := range []Dir{Up, Down, Left, Right} {
		delta := d.Delta()
		if e := r.Floor.EnemyAt(Cell{r.Player.Pos.X + delta.X, r.Player.Pos.Y + delta.Y}); e != nil {
			return e
		}
	}
	return nil
}

// nearestEnemy returns the living enemy closest to the player and its
// Manhattan distance.
func (r *Run) nearestEnemy() (*Enemy, int) {
	var best *Enemy
	bestDist := 1 << 30
	for _, e := range r.Floor.Enemies {
		if e.HP <= 0 {
			continue
		}
		d := abs(e.Pos.X-r.Player.Pos.X) + abs(e.Pos.Y-r.Player.Pos.Y)
		if d < bestDist {
			best, bestDist = e, d
		}
	}
	return best, bestDist
}

// rayTarget picks the nearest enemy within reach (orthogonal distance) whose
// straight line to the caster is not blocked by terrain.
func (r *Run) rayTarget(reach int) *Enemy {
	p := r.Player
	var best *Enemy
	bestDist := reach + 1
	for _, e := range r.Floor.Enemies {
		if e.HP <= 0 {
			continue
		}
		dist := abs(e.Pos.X-p.Pos.X) + abs(e.Pos.Y-p.Pos.Y)
		if dist > reach || dist >= bestDist {
			continue
		}
		if !lineClear(r.Floor, p.Pos, e.Pos) {
			continue
		}
		best, bestDist = e, dist
	}
	return best
}

// lineClear is a Bresenham line-of-sight test between two cells; walls block.
// The endpoints are not tested.
func lineClear(f *Floor, a, b Cell) bool {
	x0, y0, x1, y1 := a.X, a.Y, b.X, b.Y
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := step(x0, x1), step(y0, y1)
	err := dx + dy
	for {
		if x0 == x1 && y0 == y1 {
			return true
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
		if x0 == x1 && y0 == y1 {
			return true
		}
		if f.IsSolid(Cell{x0, y0}) {
			return false
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func step(a, b int) int {
	switch {
	case a < b:
		return 1
	case a > b:
		return -1
	default:
		return 0
	}
}

// rollDice sums count dsize rolls using the run's RNG.
func rollDice(rng interface{ Intn(int) int }, d Dice) int {
	total := 0
	for i := 0; i < d.N; i++ {
		total += rng.Intn(d.Size) + 1
	}
	return total
}
