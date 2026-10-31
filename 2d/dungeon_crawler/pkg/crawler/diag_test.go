package crawler

import "testing"

// --- 19.1 eight-direction movement ---

func TestDiagonalDeltas(t *testing.T) {
	cases := []struct {
		d    Dir
		want Cell
	}{
		{Up, Cell{0, -1}},
		{Down, Cell{0, 1}},
		{Left, Cell{-1, 0}},
		{Right, Cell{1, 0}},
		{UpLeft, Cell{-1, -1}},
		{UpRight, Cell{1, -1}},
		{DownLeft, Cell{-1, 1}},
		{DownRight, Cell{1, 1}},
	}
	for _, c := range cases {
		if got := c.d.Delta(); got != c.want {
			t.Errorf("Delta(%v) = %v, want %v", c.d, got, c.want)
		}
	}
}

func TestDiagonalStrideSpendsActionAndMoves(t *testing.T) {
	r := newTestRun("Fighter")
	r.Floor.SetFloor(Cell{2, 0}) // one open side (Right) is enough
	before := r.Player.ActionsLeft
	evs := r.Move(UpRight)
	if r.Player.Pos != (Cell{2, 0}) {
		t.Fatalf("pos %v, want (2,0)", r.Player.Pos)
	}
	if r.Player.ActionsLeft != before-1 {
		t.Errorf("actions %d, want %d", r.Player.ActionsLeft, before-1)
	}
	if len(evs) == 0 {
		t.Error("expected stride log event")
	}
}

func TestDiagonalStrideClippedCornerCostsNothing(t *testing.T) {
	r := newTestRun("Fighter")
	r.Player.Pos = Cell{5, 1}
	r.Floor.SetFloor(Cell{6, 2}) // diagonal target is open...
	r.Floor.SetWall(Cell{6, 1})  // ...but both ortho neighbours are solid
	before := r.Player.ActionsLeft
	r.Move(DownRight)
	if r.Player.Pos != (Cell{5, 1}) {
		t.Error("corner-clip must not move the player")
	}
	if r.Player.ActionsLeft != before {
		t.Error("corner-clip must not spend an action")
	}
}

func TestDiagonalBumpStrikesEnemy(t *testing.T) {
	r := newTestRun("Fighter")
	r.rollD20 = func() int { return 20 }
	r.Floor.SetFloor(Cell{2, 0})
	foe := placeEnemy(r, "rat", Cell{2, 0})
	before := r.Player.ActionsLeft
	evs := r.Move(UpRight)
	if r.Player.Pos != (Cell{1, 1}) {
		t.Error("a diagonal bump must strike in place, not move")
	}
	if foe.HP <= 0 && len(evs) == 0 {
		t.Error("expected strike events")
	}
	if foe.HP > 0 {
		t.Errorf("nat-20 diagonal strike must hurt the rat, hp %d", foe.HP)
	}
	if r.Player.ActionsLeft != before-1 {
		t.Error("diagonal bump spends 1 action")
	}
}

func TestTownDiagonalWalkAndGateBump(t *testing.T) {
	town := NewTown(&Meta{Selected: "Fighter"})
	if town.Pos != (Cell{10, 4}) {
		t.Fatalf("spawn %v, want (10,4)", town.Pos)
	}
	town.Move(UpRight) // (11,3) is open square
	if town.Pos != (Cell{11, 3}) {
		t.Errorf("diagonal walk: pos %v, want (11,3)", town.Pos)
	}
	town.Pos = Cell{8, 11}
	town.Move(DownRight) // corner-bump the gate at (9,12)
	if !town.StartRequested {
		t.Error("diagonal gate bump must request a run")
	}
}
