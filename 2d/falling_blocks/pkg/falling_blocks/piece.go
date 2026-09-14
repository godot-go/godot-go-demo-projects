package fallingblocks

// Kind identifies one of the seven standard tetrominoes.
type Kind int

const (
	KindI Kind = iota
	KindO
	KindT
	KindS
	KindZ
	KindJ
	KindL
)

var allKinds = [7]Kind{KindI, KindO, KindT, KindS, KindZ, KindJ, KindL}

// AllKinds returns the seven standard tetromino kinds.
func AllKinds() []Kind {
	return allKinds[:]
}

// KindName returns a single letter display name for a piece.
func KindName(k Kind) string {
	switch k {
	case KindI:
		return "I"
	case KindO:
		return "O"
	case KindT:
		return "T"
	case KindS:
		return "S"
	case KindZ:
		return "Z"
	case KindJ:
		return "J"
	case KindL:
		return "L"
	}
	return "?"
}

// Cell is a grid coordinate.
type Cell struct {
	X, Y int
}

// pieceLayout describes a piece's bounding box size, spawn column, and the
// cells of its spawn (rotation 0) orientation.
type pieceLayout struct {
	boxSize int
	spawnX  int
	cells   []Cell
}

var layouts = map[Kind]pieceLayout{
	KindI: {boxSize: 4, spawnX: 3, cells: []Cell{{0, 2}, {1, 2}, {2, 2}, {3, 2}}},
	KindO: {boxSize: 2, spawnX: 4, cells: []Cell{{0, 0}, {1, 0}, {0, 1}, {1, 1}}},
	KindT: {boxSize: 3, spawnX: 3, cells: []Cell{{1, 0}, {0, 1}, {1, 1}, {2, 1}}},
	KindS: {boxSize: 3, spawnX: 3, cells: []Cell{{1, 0}, {2, 0}, {0, 1}, {1, 1}}},
	KindZ: {boxSize: 3, spawnX: 3, cells: []Cell{{0, 0}, {1, 0}, {1, 1}, {2, 1}}},
	KindJ: {boxSize: 3, spawnX: 3, cells: []Cell{{0, 0}, {0, 1}, {1, 1}, {2, 1}}},
	KindL: {boxSize: 3, spawnX: 3, cells: []Cell{{2, 0}, {0, 1}, {1, 1}, {2, 1}}},
}

// shapes[k][r] holds the cells of kind k in rotation state r (0..3), where
// each clockwise step maps (x, y) -> (size-1-y, x) within the bounding box.
var shapes [7][4][]Cell

func init() {
	for k, lay := range layouts {
		base := lay.cells
		for r := 0; r < 4; r++ {
			shapes[k][r] = base
			next := make([]Cell, len(base))
			for i, c := range base {
				next[i] = Cell{lay.boxSize - 1 - c.Y, c.X}
			}
			base = next
		}
	}
}

// RotationCells returns the cells of a kind in a rotation state.
func RotationCells(k Kind, rot int) []Cell {
	return shapes[k][(rot%4+4)%4]
}

// SpawnX returns the left column of a piece's bounding box at spawn.
func SpawnX(k Kind) int { return layouts[k].spawnX }
