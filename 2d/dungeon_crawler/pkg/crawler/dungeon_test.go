package crawler

import (
	"math/rand"
	"testing"
)

// --- 4.1 generation & connectivity ---

func TestGeneratedFloorsAreConnected(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		rng := rand.New(rand.NewSource(seed))
		g := GenerateFloor(rng)

		reachable := bfsCells(g.Floor, g.Entrance)
		if !reachable[g.Stairs] {
			t.Fatalf("seed %d: stairs %v unreachable from entrance %v", seed, g.Stairs, g.Entrance)
		}
		// Every carved tile must be reachable.
		for y := 0; y < g.Floor.H; y++ {
			for x := 0; x < g.Floor.W; x++ {
				c := Cell{x, y}
				if !g.Floor.IsSolid(c) && !reachable[c] {
					t.Fatalf("seed %d: carved tile %v is isolated", seed, c)
				}
			}
		}
	}
}

func TestEntranceAndStairsAreDistinctAndSpacious(t *testing.T) {
	rng := rand.New(rand.NewSource(9))
	g := GenerateFloor(rng)
	if g.Entrance == g.Stairs {
		t.Error("entrance and stairs must differ")
	}
	if !g.Floor.TileWalkable(g.Entrance) || !g.Floor.TileWalkable(g.Stairs) {
		t.Error("both must be walkable")
	}
}

// bfsCells flood-fills walkable terrain (ignoring enemies) from start.
func bfsCells(f *Floor, start Cell) map[Cell]bool {
	seen := map[Cell]bool{start: true}
	queue := []Cell{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range []Dir{Up, Down, Left, Right} {
			delta := d.Delta()
			n := Cell{cur.X + delta.X, cur.Y + delta.Y}
			if seen[n] || !f.TileWalkable(n) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	return seen
}

// --- 4.2 feature placement ---

func placeTestFloor(seed int64, floorNum int) (*Floor, GenResult) {
	rng := rand.New(rand.NewSource(seed))
	g := GenerateFloor(rng)
	next := 0
	PlaceFeatures(rng, g, floorNum, func(kind string) *Enemy {
		next++
		var e *Enemy
		if kind == BossKind {
			e = BossStatline()
		} else {
			e = MonsterStatline(kind)
		}
		e.ID = next
		return e
	})
	return g.Floor, g
}

func countFeature(f *Floor, feat Feature) int {
	n := 0
	for _, ft := range f.Features {
		if ft == feat {
			n++
		}
	}
	return n
}

func TestFeaturePresence(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		f, _ := placeTestFloor(seed, 1)
		if countFeature(f, FeatureStairs) != 1 {
			t.Errorf("seed %d: stairs count %d", seed, countFeature(f, FeatureStairs))
		}
		if countFeature(f, FeatureLockedDoor) < 1 {
			t.Errorf("seed %d: no locked door", seed)
		}
		if countFeature(f, FeatureTrapArmed) < 1 {
			t.Errorf("seed %d: no trap", seed)
		}
		if countFeature(f, FeatureHiddenDoor) < 1 {
			t.Errorf("seed %d: no hidden door", seed)
		}
		if countFeature(f, FeaturePotion) < 1 {
			t.Errorf("seed %d: no potion", seed)
		}
	}
}

func TestMonstersRespectFloorTierAndStayAwayFromEntrance(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		f, g := placeTestFloor(seed, 2)
		allowed := FloorSpawns[2]
		if len(f.Enemies) == 0 {
			t.Fatalf("seed %d: no enemies", seed)
		}
		for _, e := range f.Enemies {
			ok := false
			for _, k := range allowed {
				if e.Kind == k {
					ok = true
				}
			}
			if !ok {
				t.Errorf("seed %d: %s spawned on floor 2", seed, e.Kind)
			}
			if !f.TileWalkable(e.Pos) {
				t.Errorf("seed %d: %s in solid tile %v", seed, e.Kind, e.Pos)
			}
			d := abs(e.Pos.X-g.Entrance.X) + abs(e.Pos.Y-g.Entrance.Y)
			if d < 4 {
				t.Errorf("seed %d: %s spawns %d tiles from entrance", seed, e.Kind, d)
			}
		}
	}
}

func TestFloorFiveHasBossNotStairs(t *testing.T) {
	f, g := placeTestFloor(3, 5)
	if countFeature(f, FeatureStairs) != 0 {
		t.Error("floor 5 must have no stairs")
	}
	var boss *Enemy
	for _, e := range f.Enemies {
		if e.IsBoss {
			boss = e
		}
	}
	if boss == nil {
		t.Fatal("floor 5 must have the boss")
	}
	d := abs(boss.Pos.X-g.Stairs.X) + abs(boss.Pos.Y-g.Stairs.Y)
	if d > 4 {
		t.Errorf("boss should guard the exit area, is %d tiles away", d)
	}
}

// --- 4.3 visibility ---

func TestFOVRadiusAndWalls(t *testing.T) {
	f := NewFloor(25, 5)
	for x := 0; x < 25; x++ {
		f.SetFloor(Cell{x, 2})
	}
	vis := ComputeFOV(f, Cell{5, 2}, 6)
	if !vis[Cell{5, 2}] {
		t.Error("origin always visible")
	}
	if !vis[Cell{10, 2}] {
		t.Error("tile 5 away down the corridor should be visible")
	}
	if vis[Cell{14, 2}] {
		t.Error("tile 9 away exceeds radius 6")
	}

	// Wall in the corridor blocks sight beyond it.
	f.solid[f.index(Cell{8, 2})] = true
	vis = ComputeFOV(f, Cell{5, 2}, 6)
	if vis[Cell{10, 2}] {
		t.Error("sight must not pass through a wall")
	}
}

func TestExploredMemoryPersists(t *testing.T) {
	f := NewFloor(25, 5)
	for x := 0; x < 25; x++ {
		f.SetFloor(Cell{x, 2})
	}
	vis := ComputeFOV(f, Cell{5, 2}, 6)
	f.MarkExplored(vis)
	if !f.ExploredAt(Cell{10, 2}) {
		t.Error("seen tile must be remembered")
	}
	if f.ExploredAt(Cell{22, 2}) {
		t.Error("never-seen tile must not be explored")
	}

	vis2 := ComputeFOV(f, Cell{22, 2}, 6)
	if vis2[Cell{5, 2}] {
		t.Error("old position should now be out of sight")
	}
	if !f.ExploredAt(Cell{5, 2}) {
		t.Error("memory must survive leaving the area")
	}
}
