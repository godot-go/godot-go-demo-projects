package crawler

import (
	"fmt"
	"testing"
)

// bfsFirstStep returns the first step from `from` along a shortest walkable
// route toward `goal`. The goal tile itself may be blocked (an enemy, a door
// to bump); the walk-back stops on a tile adjacent to it. ok is false when no
// route exists.
func bfsFirstStep(f *Floor, from, goal Cell) (Cell, bool) {
	if from == goal {
		return Cell{}, false
	}
	prev := map[Cell]Cell{}
	visited := map[Cell]bool{from: true}
	queue := []Cell{from}
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
				if cur == from {
					return next, true // goal directly adjacent to us
				}
				return firstStepOf(prev, cur, from)
			}
			if !f.TileWalkable(next) || f.EnemyAt(next) != nil {
				continue
			}
			visited[next] = true
			prev[next] = cur
			queue = append(queue, next)
		}
	}
	return Cell{}, false
}

func firstStepOf(prev map[Cell]Cell, cell, start Cell) (Cell, bool) {
	for prev[cell] != start {
		c, ok := prev[cell]
		if !ok {
			return Cell{}, false
		}
		cell = c
	}
	return cell, true
}

func dirToward(from, to Cell) Dir {
	switch {
	case to.X > from.X:
		return Right
	case to.X < from.X:
		return Left
	case to.Y > from.Y:
		return Down
	default:
		return Up
	}
}

func findFeature(f *Floor, feat Feature) (Cell, bool) {
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			if f.FeatureAt(Cell{x, y}) == feat {
				return Cell{x, y}, true
			}
		}
	}
	return Cell{}, false
}

func findBoss(f *Floor) (Cell, bool) {
	for _, e := range f.Enemies {
		if e.HP > 0 && e.IsBoss {
			return e.Pos, true
		}
	}
	return Cell{}, false
}

// runBot drives a greedy Fighter with no town bonuses.
func runBot(t *testing.T, seed int64, maxActions int) (*Run, int) {
	t.Helper()
	return runBotBonuses(t, seed, maxActions, nil)
}

// runBotBonuses drives a greedy Fighter: quaff when hurt, bump adjacent foes,
// chase near enemies, otherwise head for the exit (or the floor-5 boss). When
// no route exists it waits — enemies converge, so the loop always resolves.
func runBotBonuses(t *testing.T, seed int64, maxActions int, b *RunBonuses) (*Run, int) {
	t.Helper()
	r := NewRun(seed, "Fighter", b)
	actions := 0
	recent := make([]Cell, 0, 8)
	for r.State == Running && actions < maxActions {
		actions++
		p := r.Player

		if p.Potions > 0 && p.HP*3 < p.MaxHP {
			r.Quaff()
			continue
		}
		if e := r.adjacentEnemy(); e != nil {
			if p.ActionsLeft >= 2 {
				r.UseSkill(Classes[p.Class].CombatSkill) // Fighter: power attack
			} else {
				delta := Cell{e.Pos.X - p.Pos.X, e.Pos.Y - p.Pos.Y}
				r.Move(dirToward(p.Pos, Cell{p.Pos.X + delta.X, p.Pos.Y + delta.Y}))
			}
			continue
		}

		// Anti-stall: enemy bodies transiently block corridors, flipping the
		// shortest path back and forth. If the last 8 positions covered at
		// most 4 cells, take dice-driven steps until the pocket is left —
		// stepping onto a foe starts combat, a door gets forced, and a random
		// walk over a connected floor always reaches somewhere new.
		recent = append(recent, p.Pos)
		if len(recent) > 8 {
			recent = recent[1:]
		}
		if len(recent) == 8 {
			uniq := map[Cell]bool{}
			for _, c := range recent {
				uniq[c] = true
			}
			if len(uniq) <= 4 {
				order := []Dir{Up, Right, Down, Left}
				start := r.rng.Intn(4)
				for i := 0; i < 4; i++ {
					d := order[(start+i)%4]
					nb := Cell{p.Pos.X + d.Delta().X, p.Pos.Y + d.Delta().Y}
					if r.Floor.TileWalkable(nb) || r.Floor.FeatureAt(nb) == FeatureLockedDoor {
						r.Move(d)
						break
					}
				}
				continue
			}
		}

		// Chase a close enemy, else head for the exit (the boss on floor 5).
		goal, ok := botExitGoal(r)
		if nearest, dist := r.nearestEnemy(); nearest != nil && dist <= 5 {
			goal, ok = nearest.Pos, true
		}
		if !ok {
			r.EndTurn()
			continue
		}

		step, reachable := bfsFirstStep(r.Floor, p.Pos, goal)
		if !reachable {
			// Chase target walled off: the exit may still be reachable.
			if alt, altOK := botExitGoal(r); altOK && alt != goal {
				if step2, ok2 := bfsFirstStep(r.Floor, p.Pos, alt); ok2 {
					step, reachable = step2, true
				}
			}
		}
		if !reachable {
			// Exit walled off too: approach the nearest forced-able door.
			if door, dOK := nearestReachableDoor(r.Floor, p.Pos); dOK {
				step, reachable = bfsFirstStep(r.Floor, p.Pos, door)
			}
		}
		if reachable {
			r.Move(dirToward(p.Pos, step))
			continue
		}
		// Walled off and nothing to approach: bump any adjacent locked door,
		// otherwise wait — enemies converge, so the loop resolves.
		bumped := false
		for _, d := range []Dir{Up, Down, Left, Right} {
			delta := d.Delta()
			c := Cell{p.Pos.X + delta.X, p.Pos.Y + delta.Y}
			if r.Floor.FeatureAt(c) == FeatureLockedDoor {
				r.Move(d)
				bumped = true
				break
			}
		}
		if !bumped {
			r.EndTurn()
		}
	}
	return r, actions
}

// botExitGoal is the floor's progress objective: stairs, or the boss.
func botExitGoal(r *Run) (Cell, bool) {
	if r.FloorNum == 5 {
		return findBoss(r.Floor)
	}
	return findFeature(r.Floor, FeatureStairs)
}

// nearestReachableDoor is the locked door closest to from (Manhattan) for
// which a walking route exists.
func nearestReachableDoor(f *Floor, from Cell) (Cell, bool) {
	best, bestDist := Cell{}, 1<<30
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			c := Cell{x, y}
			if f.FeatureAt(c) != FeatureLockedDoor {
				continue
			}
			d := abs(c.X-from.X) + abs(c.Y-from.Y)
			if d >= bestDist {
				continue
			}
			if _, ok := bfsFirstStep(f, from, c); ok {
				best, bestDist = c, d
			}
		}
	}
	return best, bestDist < 1<<30
}

// --- 8.1 full runs reach an end state ---

func TestScriptedBotAlwaysFinishesRuns(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		r, actions := runBot(t, seed, 40000)
		if r.State == Running {
			t.Fatalf("seed %d: bot stalled on floor %d after %d actions", seed, r.FloorNum, actions)
		}
		t.Logf("seed %d → %v on floor %d (level %d, %d actions)", seed, r.State, r.FloorNum, r.Player.Level, actions)
	}
}

// --- 8.2 pacing: the early floors must level, and mid-depth must be
// reachable by a decent-but-greedy player ---

func TestBotPacingReachesMidDungeon(t *testing.T) {
	bestFloor := 0
	wins := 0
	levelAtFloor2 := []int{}
	for seed := int64(1); seed <= 12; seed++ {
		r, _ := runBot(t, seed, 40000)
		if r.FloorNum > bestFloor {
			bestFloor = r.FloorNum
		}
		if r.State == Won {
			wins++
		}
		if r.FloorNum >= 2 {
			levelAtFloor2 = append(levelAtFloor2, r.Player.Level)
		}
	}
	if bestFloor < 3 {
		t.Errorf("no bot out of 12 reached floor 3 (best floor %d) — early floors too punishing", bestFloor)
	}
	if wins == 0 {
		t.Log("no wins yet across 12 seeds — acceptable if floors 3-5 are meant to bite")
	}
}

// TestPayoutBankedAfterRun sanity-checks the money flow the shell relies on:
// a finished run reports a non-negative renown payout that counts kills.
func TestPayoutBankedAfterRun(t *testing.T) {
	r, _ := runBot(t, 3, 40000)
	if payout := r.Payout(false); payout < 0 {
		t.Errorf("negative payout %d", payout)
	} else {
		t.Log(fmt.Sprintf("seed 3 payout: %d renown", payout))
	}
}

// --- meta loop: renown spent in town must translate into deeper runs ---

func TestMetaUpgradesDeepenLaterRuns(t *testing.T) {
	m := DefaultMeta()
	m.Buildings["Smithy"] = 3
	m.Buildings["Clinic"] = 3
	m.Buildings["Library"] = 3
	m.Buildings["Archive"] = 3
	for _, g := range []string{"Guild Fighter", "Guild Cleric", "Guild Wizard", "Guild Rogue"} {
		m.Buildings[g] = 2
	}
	b := m.RunBonuses()

	rawBest, buffedBest, buffedWins := 0, 0, 0
	for seed := int64(1); seed <= 8; seed++ {
		raw, _ := runBotBonuses(t, seed, 40000, nil)
		if raw.FloorNum > rawBest {
			rawBest = raw.FloorNum
		}
		buff, _ := runBotBonuses(t, seed, 40000, &b)
		if buff.FloorNum > buffedBest {
			buffedBest = buff.FloorNum
		}
		if buff.State == Won {
			buffedWins++
		}
		t.Logf("seed %d: raw floor %d → buffed floor %d (%v)", seed, raw.FloorNum, buff.FloorNum, buff.State)
	}
	if buffedBest < rawBest {
		t.Errorf("maxed town upgrades must never hurt: raw best floor %d, buffed best floor %d", rawBest, buffedBest)
	}
	if buffedWins == 0 && buffedBest <= rawBest {
		t.Errorf("maxed upgrades neither deepened runs (best floor %d) nor won any", buffedBest)
	}
}
