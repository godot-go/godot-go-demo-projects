package fallingblocks

import (
	"math/rand"
	"testing"
)

func fillRow(b *Board, y int, except ...int) {
	for x := 0; x < b.W; x++ {
		skip := false
		for _, e := range except {
			if x == e {
				skip = true
			}
		}
		b.Set(x, y, !skip)
	}
}

// --- Bag / generation ---------------------------------------------------

func TestBagCompleteness(t *testing.T) {
	r := NewRandomizer(rand.New(rand.NewSource(42)))
	seen := map[Kind]bool{}
	for i := 0; i < 7; i++ {
		seen[r.Next()] = true
	}
	if len(seen) != 7 {
		t.Fatalf("first bag missing kinds: %v", seen)
	}
	seen = map[Kind]bool{}
	for i := 0; i < 3; i++ {
		seen[r.Next()] = true
	}
	// Second bag continues without repeats until 7 fresh kinds.
	counts := map[Kind]int{}
	for i := 0; i < 4; i++ {
		counts[r.Next()]++
	}
	for k, c := range counts {
		if c > 1 && seen[k] {
			t.Fatalf("kind %v repeated within a bag", k)
		}
	}
}

func TestPreview(t *testing.T) {
	r := NewRandomizer(rand.New(rand.NewSource(7)))
	_ = r.Next() // pop current piece
	p := r.Preview()
	if got := r.Next(); got != p {
		t.Fatalf("preview %v != next popped %v", got, p)
	}
}

// --- Movement / rotation --------------------------------------------------

func TestMoveAgainstWall(t *testing.T) {
	g := NewGame(1)
	// Move T left until rejected.
	for g.Move(-1) {
	}
	x := g.CurX
	if g.Move(-1) {
		t.Fatal("move into wall accepted")
	}
	if g.CurX != x {
		t.Fatal("position changed on rejected move")
	}
}

func TestWallKickOnRotate(t *testing.T) {
	g := NewGame(1)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindT, 0, 4, 8
	// Block the cell below the nub column so the unshifted rotation collides.
	g.Board.Set(5, 10, true)
	if !g.Rotate(1) {
		t.Fatal("expected kick rotation to succeed")
	}
	if g.CurX != 5 || g.CurRot != 1 {
		t.Fatalf("expected kick right shift: x=%d rot=%d", g.CurX, g.CurRot)
	}
}

func TestRotateRejectedWhenAllKicksBlocked(t *testing.T) {
	g := NewGame(1)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindT, 0, 3, 8
	for _, c := range []Cell{{4, 10}, {3, 8}, {3, 10}, {6, 9}, {5, 8}, {5, 10}} {
		g.Board.Set(c.X, c.Y, true)
	}
	before := [3]int{g.CurX, g.CurY, g.CurRot}
	if g.Rotate(1) {
		t.Fatal("rotation should be rejected")
	}
	if g.CurX != before[0] || g.CurY != before[1] || g.CurRot != before[2] {
		t.Fatal("piece mutated on rejected rotation")
	}
}

// --- Drops / gravity / locking --------------------------------------------

func TestGravityStep(t *testing.T) {
	g := NewGame(1)
	y := g.CurY
	if !g.GravityStep() {
		t.Fatal("gravity step should descend")
	}
	if g.CurY != y+1 {
		t.Fatalf("expected y %d, got %d", y+1, g.CurY)
	}
}

func TestSoftDropScoring(t *testing.T) {
	g := NewGame(1)
	score := g.Score
	g.SoftDrop()
	if g.Score != score+1 {
		t.Fatalf("soft drop awarded %d", g.Score-score)
	}
}

func TestHardDropLocksAndScores(t *testing.T) {
	g := NewGame(1)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindT, 0, 3, 0 // T at spawn position (bottom row y=1)
	score := g.Score
	g.HardDrop()
	dist := 20 // T bottom row lands on floor row 21
	if g.Score-score != 2*dist {
		t.Fatalf("expected %d hard-drop points, got %d", 2*dist, g.Score-score)
	}
	// The old piece is locked at the floor.
	if !g.Board.At(3, 21) || !g.Board.At(4, 21) || !g.Board.At(5, 21) {
		t.Fatal("T cells not stamped at landing row")
	}
}

func TestGravityLocksWhenResting(t *testing.T) {
	g := NewGame(1)
	for g.GravityStep() {
	}
	if res := g.LockPieceAtBottom(); res.Cleared != 0 || g.Over {
		t.Fatal("expected resting lock without game over")
	}
}

func TestPartialRowNotCleared(t *testing.T) {
	g := NewGame(1)
	fillRow(g.Board, 21, 0, 1, 2, 3) // big hole; T can never fill cols 0-3 in one lock at floor
	g.Board.Set(6, 19, false)
	res := g.Lock()
	if res.Cleared != 0 || g.Lines != 0 {
		t.Fatalf("partial row cleared: %+v", res)
	}
}

// --- Clearing / scoring / levels -------------------------------------------

func TestDoubleClearScoringAtLevel2(t *testing.T) {
	g := NewGame(1)
	g.Level = 2
	fillRow(g.Board, 20, 3)
	fillRow(g.Board, 21, 3)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindI, 1, 2, 18 // cells in column 3, rows 18..21
	res := g.Lock()
	if res.Cleared != 2 {
		t.Fatalf("expected double clear, got %d", res.Cleared)
	}
	if g.Score != 600 {
		t.Fatalf("expected 600 points (300 x level 2), got %d", g.Score)
	}
	if g.Lines != 2 {
		t.Fatalf("lines=%d", g.Lines)
	}
}

func TestLevelUpAt10Lines(t *testing.T) {
	g := NewGame(1)
	g.Lines = 9
	fillRow(g.Board, 21, 0)
	g.Cur, g.CurRot, g.CurX, g.CurY = KindI, 1, -1, 18 // column 0 rows 18..21
	g.Lock()
	if g.Level != 2 {
		t.Fatalf("level should be 2, got %d", g.Level)
	}
	if g.GravityInterval() >= GravityInterval(1) {
		t.Fatal("gravity should speed up with level")
	}
}

func TestGarbageAttackTableAndCombo(t *testing.T) {
	cases := []struct {
		clears, combo, want int
	}{
		{1, 0, 0}, {2, 0, 1}, {3, 0, 2}, {4, 0, 4}, {2, 3, 4},
	}
	for _, c := range cases {
		g := NewGame(1)
		g.Combo = c.combo
		// Fill rows to clear c.clears rows with a vertical I in column 9.
		for i := 0; i < c.clears; i++ {
			fillRow(g.Board, 21-i, 9)
		}
		g.Cur, g.CurRot, g.CurX, g.CurY = KindI, 1, 8, 18 // column 9, rows 18..21
		res := g.Lock()
		if res.Cleared != c.clears {
			t.Fatalf("clears=%d wanted: %+v", res.Cleared, c)
		}
		if res.GarbageSent != c.want {
			t.Fatalf("clears=%d combo=%d: sent %d want %d", c.clears, c.combo, res.GarbageSent, c.want)
		}
	}
}

// --- Game over --------------------------------------------------------------

func TestGameOverOnSpawnOverlap(t *testing.T) {
	g := NewGame(1)
	// Block the whole spawn region (cols 3..6, buffer rows + first visible row)
	// so any piece kind collides at spawn.
	for y := 0; y <= 2; y++ {
		for x := 3; x <= 6; x++ {
			g.Board.Set(x, y, true)
		}
	}
	g.spawn()
	if !g.Over {
		t.Fatal("expected game over on blocked spawn")
	}
	// Inputs frozen after game over.
	if g.Move(1) || g.Rotate(1) || g.GravityStep() {
		t.Fatal("inputs accepted after game over")
	}
}

// --- Garbage ----------------------------------------------------------------

func TestGarbageAppliesOnNextLock(t *testing.T) {
	g := NewGame(1)
	fillRow(g.Board, 10, 0) // near-full stack row (not full: survives clear)
	g.AddGarbage(2)
	beforeTop := g.Board.TopmostRow()
	g.Cur, g.CurRot, g.CurX, g.CurY = KindO, 0, 0, 5 // lock somewhere harmless
	res := g.Lock()
	if res.Cleared != 0 {
		t.Fatalf("unexpected clear: %+v", res)
	}
	if g.Pending != 0 {
		t.Fatal("pending garbage not consumed")
	}
	afterTop := g.Board.TopmostRow()
	if afterTop != 3 || !g.Board.At(5, 8) {
		t.Fatalf("stack should rise by 2: before=%d after=%d", beforeTop, afterTop)
	}
	// Two bottom rows are garbage with exactly one gap each.
	for y := g.Board.H - 2; y < g.Board.H; y++ {
		gaps := 0
		for x := 0; x < g.Board.W; x++ {
			if !g.Board.At(x, y) {
				gaps++
			}
		}
		if gaps != 1 {
			t.Fatalf("garbage row %d has %d gaps", y, gaps)
		}
	}
}

func TestGarbageInducedTopOut(t *testing.T) {
	g := NewGame(1)
	// Pre-fill from first visible row down, leaving cols 8-9 open.
	for y := g.Board.Buffer; y < g.Board.H; y++ {
		fillRow(g.Board, y, 8, 9)
	}
	g.AddGarbage(3)
	// Vertical I in the two-col shaft: fills one column, so no row completes
	// and the +3 push deterministically breaches the spawn area.
	g.Cur, g.CurRot, g.CurX, g.CurY = KindI, 1, 8, 4
	g.Lock()
	if !g.Over {
		t.Fatal("expected garbage push to breach the spawn area")
	}
}

// --- Pause -------------------------------------------------------------

func TestPauseFreezes(t *testing.T) {
	g := NewGame(1)
	g.Paused = true
	y := g.CurY
	if g.GravityStep() || g.Move(1) || g.Rotate(1) {
		t.Fatal("action accepted while paused")
	}
	if g.CurY != y {
		t.Fatal("gravity moved piece while paused")
	}
	g.Paused = false
	if !g.GravityStep() {
		t.Fatal("gravity should resume after unpause")
	}
}

func TestAllKindsOrderStable(t *testing.T) {
	want := []Kind{KindI, KindO, KindT, KindS, KindZ, KindJ, KindL}
	for seed := int64(1); seed <= 20; seed++ {
		_ = NewRandomizer(rand.New(rand.NewSource(seed)))
	}
	got := AllKinds()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AllKinds order mutated by shuffling: got %v, want %v", got, want)
		}
	}
}
