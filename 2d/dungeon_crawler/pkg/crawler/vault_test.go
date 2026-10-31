package crawler

import "testing"

// --- vault-placed locked doors ---

// floodSealed walks the open floor from the entrance treating the sealed
// cell (the locked door) as a wall, and reports which targets stay reachable.
type floodSealed map[Cell]bool

func floodFromSealed(f *Floor, from Cell, sealed Cell) map[Cell]bool {
	seen := map[Cell]bool{}
	if from == sealed {
		return seen
	}
	queue := []Cell{from}
	seen[from] = true
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range [4]Cell{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			n := Cell{c.X + d.X, c.Y + d.Y}
			if n.X < 0 || n.Y < 0 || n.X >= f.W || n.Y >= f.H {
				continue
			}
			if n == sealed || seen[n] || !f.TileWalkable(n) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	return seen
}

// TestLockedDoorSealsVaultDoorway pins the whole contract: the lock sits on
// a room's border (a doorway, never room interior), that room is optional
// (neither entrance nor stairs), and opening nothing the main route needs —
// with the door treated as a wall the stairs stay reachable.
func TestLockedDoorSealsVaultDoorway(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		for _, floorNum := range []int{1, 2, 3, 4} {
			checkVaultContract(t, seed, floorNum, true)
		}
		checkVaultContract(t, seed, 5, false) // the boss floor has no stairs
	}
}

// checkVaultContract pins the whole contract for one generated floor: the
// lock sits on a room's border ring (a doorway, never room interior), the
// room holds the floor's potion and is neither entrance nor stairs, and —
// when the floor has stairs — opening nothing the main route needs keeps
// the stairs reachable with the door treated as a wall.
func checkVaultContract(t *testing.T, seed int64, floorNum int, expectStairs bool) {
	t.Helper()
	f, g := placeTestFloor(seed, floorNum)
	locks := 0
	for i, feat := range f.Features {
		if feat != FeatureLockedDoor {
			continue
		}
		c := Cell{i % f.W, i / f.W}
		locks++
		var vault *rect
		for i := range g.Rooms {
			r := g.Rooms[i]
			if !onRing(r, c) {
				continue
			}
			if r.contains(g.Entrance) || r.contains(g.Stairs) {
				t.Fatalf("seed %d floor %d: lock %v on entrance/stairs room", seed, floorNum, c)
			}
			vault = &g.Rooms[i]
		}
		if vault == nil {
			t.Fatalf("seed %d floor %d: lock %v is on no room border", seed, floorNum, c)
		}
		if countFeature(f, FeaturePotion) < 1 {
			t.Fatalf("seed %d floor %d: no potion", seed, floorNum)
		}
		pot := firstCellOf(f, FeaturePotion)
		inVault := false
		for _, in := range cellsIn(*vault) {
			if in == pot {
				inVault = true
			}
		}
		if !inVault {
			t.Errorf("seed %d floor %d: potion %v outside vault", seed, floorNum, pot)
		}
		for _, in := range cellsIn(*vault) {
			if in == c {
				continue // the doorway itself is stamped with the lock
			}
			if f.FeatureAt(in) == FeatureLockedDoor {
				t.Errorf("seed %d floor %d: lock inside room interior at %v", seed, floorNum, in)
			}
		}
		if expectStairs {
			reach := floodFromSealed(f, g.Entrance, c)
			if !reach[g.Stairs] {
				t.Errorf("seed %d floor %d: lock %v blocks the stairs", seed, floorNum, c)
			}
		}
	}
	if locks == 0 {
		t.Errorf("seed %d floor %d: no locked door — vault missing", seed, floorNum)
	}
}

func onRing(r rect, c Cell) bool {
	return c.X >= r.x && c.X < r.x+r.w && c.Y >= r.y && c.Y < r.y+r.h &&
		(c.X == r.x || c.X == r.x+r.w-1 || c.Y == r.y || c.Y == r.y+r.h-1)
}

func cellsIn(r rect) []Cell {
	var out []Cell
	for y := r.y; y < r.y+r.h; y++ {
		for x := r.x; x < r.x+r.w; x++ {
			out = append(out, Cell{x, y})
		}
	}
	return out
}

func firstCellOf(f *Floor, feat Feature) Cell {
	for i, ft := range f.Features {
		if ft == feat {
			return Cell{i % f.W, i / f.W}
		}
	}
	return Cell{}
}
