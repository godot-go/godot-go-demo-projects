package crawler

import "math/rand"

// Dungeon geometry constants: 24 px tiles on a 41x23 grid fit the 1280x720
// window with room for the HUD strip.
const (
	MapW = 41
	MapH = 23
)

// GenResult is one generated (bare) floor plus its anchor rooms.
type GenResult struct {
	Floor    *Floor
	Entrance Cell
	Stairs   Cell
	Rooms    []rect
}

// rect is an axis-aligned map region.
type rect struct{ x, y, w, h int }

func (r rect) center() Cell { return Cell{r.x + r.w/2, r.y + r.h/2} }

// contains reports whether the cell lies inside the rect.
func (r rect) contains(c Cell) bool {
	return c.X >= r.x && c.X < r.x+r.w && c.Y >= r.y && c.Y < r.y+r.h
}

// bsp leaf of the split tree.
type leaf struct {
	area        rect
	left, right *leaf
	room        rect
	hasRoom     bool
	rep         Cell // any room center in this subtree (self if a leaf)
}

// GenerateFloor runs recursive BSP splitting, carves a room in each leaf, and
// connects sibling rooms with L-corridors so the floor is connected by
// construction. Entrance is the first room; Stairs is the farthest room.
func GenerateFloor(rng *rand.Rand) GenResult {
	f := NewFloor(MapW, MapH)
	root := splitLeaf(rect{1, 1, MapW - 2, MapH - 2}, rng, 0)

	var rooms []rect
	var carveRooms func(l *leaf)
	carveRooms = func(l *leaf) {
		if l == nil {
			return
		}
		if l.left == nil && l.right == nil {
			l.room = carveRoom(l.area, rng)
			l.hasRoom = true
			l.rep = l.room.center()
			for y := l.room.y; y < l.room.y+l.room.h; y++ {
				for x := l.room.x; x < l.room.x+l.room.w; x++ {
					f.SetFloor(Cell{x, y})
				}
			}
			rooms = append(rooms, l.room)
			return
		}
		carveRooms(l.left)
		carveRooms(l.right)
	}
	carveRooms(root)

	var connect func(l *leaf)
	connect = func(l *leaf) {
		if l == nil || l.left == nil {
			return
		}
		connect(l.left)
		connect(l.right)
		l.rep = l.left.rep
		carveCorridor(f, l.left.rep, l.right.rep, rng)
	}
	connect(root)

	entrance := rooms[0].center()
	stairs := rooms[0].center()
	best := -1
	for _, r := range rooms {
		c := r.center()
		d := abs(c.X-entrance.X) + abs(c.Y-entrance.Y)
		if d > best {
			best, stairs = d, c
		}
	}
	return GenResult{Floor: f, Entrance: entrance, Stairs: stairs, Rooms: rooms}
}

// splitLeaf recurses until leaves are small enough for one room each.
func splitLeaf(r rect, rng *rand.Rand, depth int) *leaf {
	l := &leaf{area: r}
	const minDim, maxDepth = 8, 6
	if depth >= maxDepth || (r.w <= minDim && r.h <= minDim) {
		return l
	}

	// Split the longer side (with a coin flip on squares).
	splitVertical := r.w > r.h
	if r.w == r.h {
		splitVertical = rng.Intn(2) == 0
	}
	if splitVertical {
		if r.w < 2*minDim {
			splitVertical = false
		}
	}
	if !splitVertical && r.h < 2*minDim {
		return l // no room to split either way
	}

	if splitVertical {
		left := r.w/2 - 2 + rng.Intn(5)
		if left < minDim || r.w-left < minDim {
			left = r.w / 2
		}
		l.left = splitLeaf(rect{r.x, r.y, left, r.h}, rng, depth+1)
		l.right = splitLeaf(rect{r.x + left, r.y, r.w - left, r.h}, rng, depth+1)
	} else {
		up := r.h/2 - 2 + rng.Intn(5)
		if up < minDim || r.h-up < minDim {
			up = r.h / 2
		}
		l.left = splitLeaf(rect{r.x, r.y, r.w, up}, rng, depth+1)
		l.right = splitLeaf(rect{r.x, r.y + up, r.w, r.h - up}, rng, depth+1)
	}
	return l
}

// carveRoom picks a 3..6 room inside the leaf leaving the leaf border solid.
func carveRoom(a rect, rng *rand.Rand) rect {
	w := 3 + rng.Intn(4)
	if w > a.w-1 {
		w = a.w - 1
	}
	h := 3 + rng.Intn(3)
	if h > a.h-1 {
		h = a.h - 1
	}
	x := a.x + rng.Intn(a.w-w+1)
	y := a.y + rng.Intn(a.h-h+1)
	return rect{x, y, w, h}
}

// carveCorridor carves an L corridor between two centers.
func carveCorridor(f *Floor, a, b Cell, rng *rand.Rand) {
	horizontalFirst := rng.Intn(2) == 0
	if horizontalFirst {
		for x := min(a.X, b.X); x <= max(a.X, b.X); x++ {
			f.SetFloor(Cell{x, a.Y})
		}
		for y := min(a.Y, b.Y); y <= max(a.Y, b.Y); y++ {
			f.SetFloor(Cell{b.X, y})
		}
	} else {
		for y := min(a.Y, b.Y); y <= max(a.Y, b.Y); y++ {
			f.SetFloor(Cell{a.X, y})
		}
		for x := min(a.X, b.X); x <= max(a.X, b.X); x++ {
			f.SetFloor(Cell{x, b.Y})
		}
	}
}

// PickSpawn returns a random species eligible for a floor.
func PickSpawn(rng *rand.Rand, floorNum int) string {
	kind := FloorSpawns[floorNum]
	return kind[rng.Intn(len(kind))]
}

// ringCells returns the perimeter cells of a room.
func ringCells(r rect) []Cell {
	var out []Cell
	for x := r.x; x < r.x+r.w; x++ {
		out = append(out, Cell{x, r.y}, Cell{x, r.y + r.h - 1})
	}
	for y := r.y + 1; y < r.y+r.h-1; y++ {
		out = append(out, Cell{r.x, y}, Cell{r.x + r.w - 1, y})
	}
	return out
}

// hasInteriorFloor reports whether the border cell c opens into the room —
// a doorway punched from outside must be walkable on the room side too, or
// locking it would choke a corridor instead of sealing a room.
func hasInteriorFloor(f *Floor, r rect, c Cell) bool {
	var in []Cell
	if c.X == r.x {
		in = append(in, Cell{c.X + 1, c.Y})
	}
	if c.X == r.x+r.w-1 {
		in = append(in, Cell{c.X - 1, c.Y})
	}
	if c.Y == r.y {
		in = append(in, Cell{c.X, c.Y + 1})
	}
	if c.Y == r.y+r.h-1 {
		in = append(in, Cell{c.X, c.Y - 1})
	}
	for _, n := range in {
		if f.TileWalkable(n) {
			return true
		}
	}
	return false
}

// floodCells walks every reachable cell from from (TileWalkable, which
// already treats a locked door as blocking).
func floodCells(f *Floor, from Cell) map[Cell]bool {
	seen := map[Cell]bool{from: true}
	queue := []Cell{from}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range [4]Cell{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			n := Cell{c.X + d.X, c.Y + d.Y}
			if !f.InBounds(n) || seen[n] || !f.TileWalkable(n) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	return seen
}

// containsInterior reports whether c lies strictly inside the room.
func (r rect) containsInterior(c Cell) bool {
	return c.X > r.x && c.X < r.x+r.w-1 && c.Y > r.y && c.Y < r.y+r.h-1
}

// findVault walls surplus doorways of an optional room, locks the kept one,
// and verifies the seal — backtracking across every room and doorway until
// one seals safely. The room's border keeps exactly one walkable doorway,
// the locked-door tile itself; everything else is interior-loot only.
func findVault(f *Floor, g GenResult, rng *rand.Rand) (rect, Cell, bool) {
	before := floodCells(f, g.Entrance)
	for _, i := range rng.Perm(len(g.Rooms)) {
		r := g.Rooms[i]
		if r.contains(g.Entrance) || r.contains(g.Stairs) {
			continue
		}
		var holes []Cell
		for _, edge := range ringCells(r) {
			if f.TileWalkable(edge) && hasInteriorFloor(f, r, edge) {
				holes = append(holes, edge)
			}
		}
		if len(holes) == 0 {
			continue
		}
		for _, keep := range rng.Perm(len(holes)) {
			var walls []Cell
			for j, h := range holes {
				if j != keep {
					walls = append(walls, h)
				}
			}
			door := holes[keep]
			for _, w := range walls {
				f.SetWall(w)
			}
			f.SetFeature(door, FeatureLockedDoor)
			after := floodCells(f, g.Entrance)
			broken := !after[g.Stairs]
			for in := range before {
				if r.containsInterior(in) {
					if after[in] {
						broken = true // loot leaks through a seam
					}
				} else if !r.contains(in) && !after[in] {
					broken = true // the walls severed the rest of the floor
				}
			}
			if !broken {
				return r, door, true
			}
			for _, w := range walls { // revert and try the next doorway
				f.SetFloor(w)
			}
			f.SetFeature(door, FeatureNone)
		}
	}
	return rect{}, Cell{}, false
}

// freeInterior lists the unoccupied walkable cells strictly inside a room.
func freeInterior(f *Floor, r rect) []Cell {
	var out []Cell
	for y := r.y + 1; y < r.y+r.h-1; y++ {
		for x := r.x + 1; x < r.x+r.w-1; x++ {
			c := Cell{x, y}
			if f.TileWalkable(c) && f.FeatureAt(c) == FeatureNone {
				out = append(out, c)
			}
		}
	}
	return out
}

// PlaceFeatures stamps stairs, the locked-door vault, traps, a hidden door, a potion,
// and the tier's monsters onto a generated floor. newEnemy builds the enemy
// for a species (the Run assigns IDs through its own factory).
func PlaceFeatures(rng *rand.Rand, g GenResult, floorNum int, newEnemy func(kind string) *Enemy) {
	f := g.Floor

	if floorNum < 5 {
		f.SetFeature(g.Stairs, FeatureStairs)
	}
	var vault *rect
	free := func(minEntranceDist int) (Cell, bool) {
		for attempt := 0; attempt < 200; attempt++ {
			c := Cell{rng.Intn(f.W), rng.Intn(f.H)}
			if !f.TileWalkable(c) || f.FeatureAt(c) != FeatureNone || f.EnemyAt(c) != nil {
				continue
			}
			if c == g.Entrance || c == g.Stairs {
				continue
			}
			if vault != nil && vault.contains(c) {
				continue // the sealed vault holds only its loot
			}
			if abs(c.X-g.Entrance.X)+abs(c.Y-g.Entrance.Y) < minEntranceDist {
				continue
			}
			return c, true
		}
		return Cell{}, false
	}

	// The locked door seals an optional vault: an away room with one
	// locked doorway on its border, promised loot inside — placed by
	// findVault, which backtracks (walling surplus doorways) until a seal
	// provably leaves the rest of the floor intact.
	// Traps and monsters later spawn around the vault, not inside it.
	vaultPotion, vaultWeapon := false, false
	if v, door, ok := findVault(f, g, rng); ok {
		f.SetFeature(door, FeatureLockedDoor)
		vault = &v
		if inside := freeInterior(f, v); len(inside) > 0 {
			rng.Shuffle(len(inside), func(i, j int) { inside[i], inside[j] = inside[j], inside[i] })
			f.SetFeature(inside[0], FeaturePotion)
			vaultPotion = true
			if len(inside) > 1 && rng.Intn(2) == 0 {
				f.SetFeature(inside[1], FeatureWeapon)
				vaultWeapon = true
			}
		}
	}
	if !vaultPotion {
		if c, ok := free(3); ok {
			f.SetFeature(c, FeaturePotion)
		}
	}
	if !vaultWeapon {
		if c, ok := free(3); ok {
			f.SetFeature(c, FeatureWeapon)
		}
	}
	for i := 0; i < 2+floorNum/2; i++ {
		if c, ok := free(4); ok {
			f.SetFeature(c, FeatureTrapArmed)
		}
	}
	if c, ok := findHiddenDoorSpot(rng, f, g); ok {
		f.SetFeature(c, FeatureHiddenDoor)
	}

	count := 2 + floorNum
	if floorNum == 5 {
		count = 3 // the boss is the floor's centerpiece; escorts stay sparse
	}
	for i := 0; i < count; i++ {
		if c, ok := free(4); ok {
			e := newEnemy(PickSpawn(rng, floorNum))
			e.Pos = c
			f.Enemies = append(f.Enemies, e)
		}
	}
	if floorNum == 5 {
		boss := newEnemy(BossKind)
		boss.Pos = guardSpot(f, g.Stairs)
		f.Enemies = append(f.Enemies, boss)
	}
}

// findHiddenDoorSpot picks a solid tile that borders open floor away from the
// entrance: a wall someone could search and find.
func findHiddenDoorSpot(rng *rand.Rand, f *Floor, g GenResult) (Cell, bool) {
	var candidates []Cell
	for y := 1; y < f.H-1; y++ {
		for x := 1; x < f.W-1; x++ {
			c := Cell{x, y}
			if !f.IsSolid(c) || f.FeatureAt(c) != FeatureNone {
				continue
			}
			if abs(c.X-g.Entrance.X)+abs(c.Y-g.Entrance.Y) < 4 {
				continue
			}
			openNeighbors := 0
			for _, d := range []Dir{Up, Down, Left, Right} {
				n := Cell{c.X + d.Delta().X, c.Y + d.Delta().Y}
				if f.InBounds(n) && !f.IsSolid(n) && f.FeatureAt(n) == FeatureNone {
					openNeighbors++
				}
			}
			if openNeighbors >= 1 {
				candidates = append(candidates, c)
			}
		}
	}
	if len(candidates) == 0 {
		return Cell{}, false
	}
	return candidates[rng.Intn(len(candidates))], true
}

// guardSpot places the boss on or next to the exit area.
func guardSpot(f *Floor, exit Cell) Cell {
	if f.TileWalkable(exit) && f.EnemyAt(exit) == nil {
		return exit
	}
	for _, d := range []Dir{Up, Down, Left, Right} {
		n := Cell{exit.X + d.Delta().X, exit.Y + d.Delta().Y}
		if f.TileWalkable(n) && f.EnemyAt(n) == nil {
			return n
		}
	}
	return exit
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
