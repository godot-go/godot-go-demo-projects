package fallingblocks

import "testing"

func cellSet(cells []Cell) map[Cell]bool {
	m := make(map[Cell]bool, len(cells))
	for _, c := range cells {
		m[c] = true
	}
	return m
}

func equalCells(a, b []Cell) bool {
	if len(a) != len(b) {
		return false
	}
	m := cellSet(a)
	for _, c := range b {
		if !m[c] {
			return false
		}
	}
	return true
}

func TestGhostCellsLandsOnFloor(t *testing.T) {
	g := NewGame(1)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindO, 0, 4, 0
	want := []Cell{{4, 20}, {5, 20}, {4, 21}, {5, 21}}
	if got := g.GhostCells(); !equalCells(got, want) {
		t.Fatalf("ghost = %v, want %v", got, want)
	}
}

func TestGhostCellsStacksOnFilledRows(t *testing.T) {
	g := NewGame(1)
	fillRow(g.Board, 21) // occupied floor: the O piece comes to rest on top
	g.Cur, g.CurRot, g.CurX, g.CurY = KindO, 0, 4, 0
	want := []Cell{{4, 19}, {5, 19}, {4, 20}, {5, 20}}
	if got := g.GhostCells(); !equalCells(got, want) {
		t.Fatalf("ghost = %v, want %v", got, want)
	}
}

func TestGhostCellsMatchHardDropLanding(t *testing.T) {
	g1 := NewGame(9)
	g2 := NewGame(9) // identical sequence
	ghost := g1.GhostCells()
	g2.HardDrop() // locks the same piece; empty board so no line clear
	for _, c := range ghost {
		if !g2.Board.At(c.X, c.Y) {
			t.Fatalf("hard drop did not land on ghost cell %v (ghost %v)", c, ghost)
		}
	}
}

func TestGhostCellsRestingEqualsCurrent(t *testing.T) {
	g := NewGame(1)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindO, 0, 4, 20 // already resting
	if got, want := g.GhostCells(), g.CurrentCells(); !equalCells(got, want) {
		t.Fatalf("resting ghost = %v, want current %v", got, want)
	}
}
