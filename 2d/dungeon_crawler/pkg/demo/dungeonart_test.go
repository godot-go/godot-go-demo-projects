package demo

import (
	"fmt"
	"testing"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// cellKey resolves one art cell of a run list ('.' for uncovered).
func cellKey(runs []artRun, dx, dy int) byte {
	for _, r := range runs {
		if dx >= r.dx && dx < r.dx+r.w && dy >= r.dy && dy < r.dy+r.h {
			return r.k
		}
	}
	return '.'
}

func runArea(runs []artRun) int {
	area := 0
	for _, r := range runs {
		area += r.w * r.h
	}
	return area
}

func TestDungeonWallRunsCoverTile(t *testing.T) {
	// Every wall tile is fully covered at 7x7 art cells with palette
	// keys least-costly to fill.
	for _, cell := range []crawler.Cell{{X: 2, Y: 3}, {X: 7, Y: 5}, {X: 13, Y: 9}} {
		runs := dungeonWallRuns(cell)
		if area := runArea(runs); area != 49 {
			t.Errorf("wall runs at (%d,%d) cover %d/49 cells", cell.X, cell.Y, area)
		}
		for _, r := range runs {
			if r.dx < 0 || r.dy < 0 || r.dx+r.w > 7 || r.dy+r.h > 7 {
				t.Errorf("wall run at (%d,%d) outside the tile: %+v", cell.X, cell.Y, r)
			}
			if _, ok := dungeonPalette[r.k]; !ok {
				t.Errorf("wall run key %q unmapped at (%d,%d)", r.k, cell.X, cell.Y)
				break
			}
		}
	}
}

func TestDungeonWallMasonryBond(t *testing.T) {
	// A top-lit facing row, two running-bond brick courses separated by
	// a full mortar course, and a mortar base row.
	runs := dungeonWallRuns(crawler.Cell{X: 0, Y: 0})
	for x := 0; x < 7; x++ {
		if got := cellKey(runs, x, 0); got != 'V' {
			t.Errorf("top edge cell (%d,0) = %q, want V facing", x, got)
		}
		if got := cellKey(runs, x, 3); got != 'w' {
			t.Errorf("course joint cell (%d,3) = %q, want w mortar", x, got)
		}
		if got := cellKey(runs, x, 6); got != 'w' {
			t.Errorf("base joint cell (%d,6) = %q, want w mortar", x, got)
		}
	}
	var bricksInSecondCourse, verticalJoints int
	for y := 4; y < 6; y++ {
		for x := 0; x < 7; x++ {
			if cellKey(runs, x, y) == 'W' {
				bricksInSecondCourse++
			}
		}
	}
	for y := 1; y < 3; y++ {
		if cellKey(runs, 3, y) == 'w' {
			verticalJoints++
		}
	}
	if bricksInSecondCourse == 0 || verticalJoints == 0 {
		t.Errorf("running bond broken: %d bricks, %d vertical joints", bricksInSecondCourse, verticalJoints)
	}
}

func TestDungeonFloorRunsCoverTile(t *testing.T) {
	for _, cell := range []crawler.Cell{{X: 3, Y: 2}, {X: 9, Y: 7}} {
		runs := dungeonFloorRuns(cell)
		if area := runArea(runs); area != 49 {
			t.Errorf("floor runs at (%d,%d) cover %d/49 cells", cell.X, cell.Y, area)
		}
		for _, r := range runs {
			if _, ok := dungeonPalette[r.k]; !ok {
				t.Errorf("floor run key %q unmapped at (%d,%d)", r.k, cell.X, cell.Y)
				break
			}
		}
	}
}

func TestDungeonFloorSlabJoints(t *testing.T) {
	// Slabs share their top and left edges: dark joints there, slab tone
	// inside; interior variation only comes from the hash speckles.
	runs := dungeonFloorRuns(crawler.Cell{X: 5, Y: 5})
	for x := 0; x < 7; x++ {
		if got := cellKey(runs, x, 0); got != 'z' {
			t.Errorf("top edge (%d,0) = %q, want z joint", x, got)
		}
	}
	for y := 0; y < 7; y++ {
		if got := cellKey(runs, 0, y); got != 'z' {
			t.Errorf("left edge (0,%d) = %q, want z joint", y, got)
		}
	}
	slabs, speckles := 0, 0
	for y := 1; y <= 6; y++ {
		for x := 1; x <= 6; x++ {
			switch cellKey(runs, x, y) {
			case 'S':
				slabs++
			case 'Z':
				speckles++
			}
		}
	}
	if slabs == 0 || speckles > 18 {
		t.Errorf("slab interior too flat or speckle noise: %d slabs, %d speckles", slabs, speckles)
	}
}

func TestDungeonArtDeterministic(t *testing.T) {
	for _, cell := range []crawler.Cell{{X: 1, Y: 1}, {X: 6, Y: 4}, {X: 10, Y: 11}} {
		wall := fmt.Sprint(dungeonWallRuns(cell))
		if again := fmt.Sprint(dungeonWallRuns(cell)); wall != again {
			t.Fatalf("wall runs at (%d,%d) unstable", cell.X, cell.Y)
		}
		floor := fmt.Sprint(dungeonFloorRuns(cell))
		if again := fmt.Sprint(dungeonFloorRuns(cell)); floor != again {
			t.Fatalf("floor runs at (%d,%d) unstable", cell.X, cell.Y)
		}
	}
}

func TestDungeonArtVaries(t *testing.T) {
	// Scatter and decals must differ across space — walls and floors
	// cannot read as one wallpaper tile repeated.
	walls := map[string]bool{}
	floors := map[string]bool{}
	for y := 0; y < 12; y++ {
		for x := 0; x < 12; x++ {
			walls[fmt.Sprint(dungeonWallRuns(crawler.Cell{X: x, Y: y}))] = true
			floors[fmt.Sprint(dungeonFloorRuns(crawler.Cell{X: x, Y: y}))] = true
		}
	}
	if len(walls) < 8 {
		t.Errorf("walls too uniform: %d distinct looks across 144 tiles", len(walls))
	}
	if len(floors) < 15 {
		t.Errorf("floors too uniform: %d distinct looks across 144 tiles", len(floors))
	}
}
