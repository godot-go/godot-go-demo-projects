package crawler

import "testing"

func TestDirDelta(t *testing.T) {
	cases := []struct {
		dir  Dir
		want Cell
	}{
		{Up, Cell{0, -1}},
		{Down, Cell{0, 1}},
		{Left, Cell{-1, 0}},
		{Right, Cell{1, 0}},
	}
	for _, tc := range cases {
		if got := tc.dir.Delta(); got != tc.want {
			t.Errorf("Dir(%d).Delta() = %v, want %v", tc.dir, got, tc.want)
		}
	}
}

// buildTestFloor makes a 5x3 room fully solid except a 3x1 strip of floor at
// y=1, x=1..3, and lets the test stamp features onto specific tiles.
func buildTestFloor() *Floor {
	f := NewFloor(5, 3)
	for x := 1; x <= 3; x++ {
		f.SetFloor(Cell{x, 1})
	}
	return f
}

func TestTileWalkability(t *testing.T) {
	f := buildTestFloor()

	if !f.TileWalkable(Cell{2, 1}) {
		t.Error("carved floor tile should be walkable")
	}
	if f.TileWalkable(Cell{0, 0}) {
		t.Error("solid tile should not be walkable")
	}
	if f.TileWalkable(Cell{-1, 1}) {
		t.Error("out-of-bounds west should not be walkable")
	}
	if f.TileWalkable(Cell{5, 1}) {
		t.Error("out-of-bounds east should not be walkable")
	}

	f.SetFeature(Cell{2, 1}, FeatureLockedDoor)
	if f.TileWalkable(Cell{2, 1}) {
		t.Error("locked door should block movement")
	}
	f.SetFeature(Cell{2, 1}, FeatureHiddenDoor)
	if f.TileWalkable(Cell{2, 1}) {
		t.Error("hidden door should block movement")
	}
	f.SetFeature(Cell{2, 1}, FeatureOpenDoor)
	if !f.TileWalkable(Cell{2, 1}) {
		t.Error("open door should be walkable")
	}
	f.SetFeature(Cell{2, 1}, FeatureStairs)
	if !f.TileWalkable(Cell{2, 1}) {
		t.Error("stairs tile should be walkable (bump descends)")
	}
}

func TestEnemyAt(t *testing.T) {
	f := buildTestFloor()
	e := &Enemy{ID: 1, Kind: "rat", Pos: Cell{3, 1}, HP: 8, MaxHP: 8}
	f.Enemies = []*Enemy{e}

	if got := f.EnemyAt(Cell{3, 1}); got != e {
		t.Errorf("EnemyAt(3,1) = %v, want the rat", got)
	}
	if got := f.EnemyAt(Cell{2, 1}); got != nil {
		t.Errorf("EnemyAt(2,1) = %v, want nil", got)
	}

	e.HP = 0
	if got := f.EnemyAt(Cell{3, 1}); got != nil {
		t.Error("dead enemy must not block its tile")
	}
}
