package crawler

import "fmt"

// enemyPhase runs every living enemy in order. Each gets the same 3-action
// budget: an enemy adjacent at the start of its turn Strikes once; a distant
// enemy Strides up to 3 tiles toward the player and does not attack on the
// turn it arrives.
func (r *Run) enemyPhase() []Event {
	evs := []Event{enemyPhaseMarker()}
	foes := make([]*Enemy, len(r.Floor.Enemies))
	copy(foes, r.Floor.Enemies) // a trap may remove enemies mid-phase
	for _, e := range foes {
		e.AttackedRecently = false
	}
	for _, e := range foes {
		if r.State != Running || e.HP <= 0 {
			continue
		}
		evs = append(evs, r.enemyAct(e)...)
	}
	return evs
}

func (r *Run) enemyAct(e *Enemy) []Event {
	if e.Pos == r.Player.Pos || r.enemyAdjacent(e) {
		return r.enemyStrike(e)
	}
	return r.enemyMarch(e, 3)
}

// enemyMarch walks up to n strides toward the player, stopping as soon as it
// is adjacent (so the enemy phase ends with the foe ready to strike) and
// resolving any traps crossed.
func (r *Run) enemyMarch(e *Enemy, n int) []Event {
	var evs []Event
	for i := 0; i < n; i++ {
		if r.enemyAdjacent(e) {
			break
		}
		next, ok := r.stepToward(e.Pos)
		if !ok {
			break
		}
		e.Pos = next
		if extra := r.enemyTriggersTrap(e); len(extra) > 0 {
			evs = append(evs, extra...)
			if e.HP <= 0 {
				break
			}
		}
	}
	return evs
}

func (r *Run) enemyAdjacent(e *Enemy) bool {
	p := r.Player
	return abs(e.Pos.X-p.Pos.X)+abs(e.Pos.Y-p.Pos.Y) == 1
}

func (r *Run) enemyStrike(e *Enemy) []Event {
	p := r.Player
	roll := r.rollD20()
	total := roll + e.AtkBonus

	switch {
	case roll == 1 || (roll != 20 && total < p.AC):
		return []Event{{Kind: "attack", Roll: roll, Total: total, DC: p.AC,
			Degree: Failure, Text: fmt.Sprintf("The %s swings — miss.", e.Kind)}}
	}

	dice := Dice{e.DmgDice, 6}
	degree := Success
	if roll == 20 {
		dice.N *= 2
		degree = CriticalSuccess
	}
	dmg := rollDice(r.rng, dice) + e.DmgMod
	p.HP -= dmg
	e.AttackedRecently = true

	ev := Event{Kind: "attack", Roll: roll, Total: total, DC: p.AC, Degree: degree,
		Damage: dmg, Text: fmt.Sprintf("The %s hits you for %d.", e.Kind, dmg),
		At: p.Pos, Subject: "player"}
	if p.HP <= 0 {
		p.HP = 0
		r.State = Lost
		ev.Text = "The " + e.Kind + " strikes you down. You die."
		return append([]Event{ev}, Event{Kind: "lose", Text: "You have fallen."})
	}
	return []Event{ev}
}

// stepToward returns the next cell along a shortest walkable route from `from`
// toward the player, avoiding walls and other enemies. The player's own tile
// is never entered. ok is false when no progress is possible.
func (r *Run) stepToward(from Cell) (Cell, bool) {
	p := r.Player
	goal := p.Pos
	prev := map[Cell]Cell{}
	visited := map[Cell]bool{from: true}
	queue := []Cell{from}

	occupied := func(c Cell) bool {
		if c == goal {
			return true
		}
		return r.Floor.EnemyAt(c) != nil
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range []Dir{Up, Down, Left, Right} {
			delta := d.Delta()
			next := Cell{cur.X + delta.X, cur.Y + delta.Y}
			if visited[next] {
				continue
			}
			if next == goal {
				// cur == from means we are orthogonally adjacent: firstStep
				// below will fail (the player tile is never a legal step),
				// which is exactly "stop here and strike".
				return r.firstStep(prev, cur, from)
			}
			if !r.Floor.TileWalkable(next) || occupied(next) {
				continue
			}
			visited[next] = true
			prev[next] = cur
			queue = append(queue, next)
		}
	}
	return Cell{}, false
}

// firstStep walks the prev-chain from cell back to start and returns the cell
// immediately after start.
func (r *Run) firstStep(prev map[Cell]Cell, cell, start Cell) (Cell, bool) {
	for prev[cell] != start {
		c, ok := prev[cell]
		if !ok {
			return Cell{}, false
		}
		cell = c
	}
	return cell, true
}
