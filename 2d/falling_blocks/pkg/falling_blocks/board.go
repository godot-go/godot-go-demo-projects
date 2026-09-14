package fallingblocks

import "math/rand"

// Board is the playfield: W columns, H rows indexed top-down. The top Buffer
// rows are hidden (spawn area); the remaining rows are visible.
type Board struct {
	W      int
	H      int
	Buffer int
	cells  []bool
	rnd    *rand.Rand
}

// NewBoard creates an empty board.
func NewBoard(w, h, buffer int, rnd *rand.Rand) *Board {
	return &Board{W: w, H: h, Buffer: buffer, cells: make([]bool, w*h), rnd: rnd}
}

// At reports whether a locked cell occupies (x, y). Out-of-bounds horizontally
// and below the floor count as occupied; above the top does not.
func (b *Board) At(x, y int) bool {
	if x < 0 || x >= b.W || y >= b.H {
		return true
	}
	if y < 0 {
		return false
	}
	return b.cells[y*b.W+x]
}

// Set writes a locked cell.
func (b *Board) Set(x, y int, v bool) {
	if x < 0 || x >= b.W || y < 0 || y >= b.H {
		return
	}
	b.cells[y*b.W+x] = v
}

// PieceCells returns the absolute cells of a piece at (px, py, rot).
func PieceCells(k Kind, rot, px, py int) []Cell {
	rel := RotationCells(k, rot)
	out := make([]Cell, len(rel))
	for i, c := range rel {
		out[i] = Cell{px + c.X, py + c.Y}
	}
	return out
}

// Fits reports whether every cell is inside empty space.
func (b *Board) Fits(cells []Cell) bool {
	for _, c := range cells {
		if b.At(c.X, c.Y) {
			return false
		}
	}
	return true
}

// Spawn places kind at its standard spawn position; false means the spawn
// cells are blocked (game over).
func (b *Board) Spawn(k Kind) (px, py int, ok bool) {
	px, py = SpawnX(k), 0
	if !b.Fits(PieceCells(k, 0, px, py)) {
		return px, py, false
	}
	return px, py, true
}

// Lock stamps the piece cells into the board.
func (b *Board) Lock(k Kind, rot, px, py int) {
	for _, c := range PieceCells(k, rot, px, py) {
		b.Set(c.X, c.Y, true)
	}
}

// ClearLines removes every completely filled row (shifting the stack down)
// and returns how many rows were removed.
func (b *Board) ClearLines() int {
	cleared := 0
	write := b.H - 1
	for read := b.H - 1; read >= 0; read-- {
		full := true
		for x := 0; x < b.W; x++ {
			if !b.cells[read*b.W+x] {
				full = false
				break
			}
		}
		if full {
			cleared++
			continue
		}
		if write != read {
			copy(b.cells[write*b.W:(write+1)*b.W], b.cells[read*b.W:(read+1)*b.W])
		}
		write--
	}
	for ; write >= 0; write-- {
		for x := 0; x < b.W; x++ {
			b.cells[write*b.W+x] = false
		}
	}
	return cleared
}

// AddGarbageLines pushes n rows of garbage up from the bottom. Each inserted
// row is full except one random gap column, aligned across the inserted rows.
// Returns the new topmost occupied row (0..Buffer means spawn area is
// breached).
func (b *Board) AddGarbageLines(n int) int {
	if n <= 0 {
		return b.TopmostRow()
	}
	gap := b.rnd.Intn(b.W)
	// Shift the existing stack up by n rows; rows pushed off the top are lost.
	for y := 0; y < b.H-n; y++ {
		copy(b.cells[y*b.W:(y+1)*b.W], b.cells[(y+n)*b.W:(y+n+1)*b.W])
	}
	for y := b.H - n; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			b.cells[y*b.W+x] = x != gap
		}
	}
	return b.TopmostRow()
}

// TopmostRow returns the index of the highest row containing a locked cell,
// or b.H when the board is empty.
func (b *Board) TopmostRow() int {
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if b.cells[y*b.W+x] {
				return y
			}
		}
	}
	return b.H
}

// LockedInBuffer reports whether any locked cell sits in the hidden buffer
// rows (the stack breached the spawn area).
func (b *Board) LockedInBuffer() bool {
	return b.TopmostRow() < b.Buffer
}
