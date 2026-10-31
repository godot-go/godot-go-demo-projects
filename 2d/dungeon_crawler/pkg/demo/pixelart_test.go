package demo

import (
	"testing"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// wharfSprite is a tiny test sprite: 3x3 with a transparent corner, so
// rasterization rules are checkable by hand.
var wharfSprite = sprite{
	".ab",
	"acc",
	"b..",
}

func TestSpriteCellsSkipsTransparent(t *testing.T) {
	cells := spriteCells(wharfSprite)
	if len(cells) != 6 { // 9 minus three dots
		t.Fatalf("got %d cells: %+v", len(cells), cells)
	}
	for _, c := range cells {
		if c.dx == 0 && c.dy == 0 {
			t.Errorf("transparent (0,0) was rasterized")
		}
	}
}

func TestSpriteCellsDeterministic(t *testing.T) {
	a := spriteCells(wharfSprite)
	b := spriteCells(wharfSprite)
	if len(a) != len(b) {
		t.Fatal("rasterization not stable")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("cell %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestSpriteCellsIndexOrder(t *testing.T) {
	// Rows top-to-bottom, columns left-to-right — per-pixel shading in the
	// art strings lands where the artist drew it.
	cells := spriteCells(wharfSprite)
	want := []spriteCell{
		{1, 0, 'a'}, {2, 0, 'b'},
		{0, 1, 'a'}, {1, 1, 'c'}, {2, 1, 'c'},
		{0, 2, 'b'},
	}
	for i, c := range want {
		if cells[i] != c {
			t.Errorf("cell %d = %+v, want %+v", i, cells[i], c)
		}
	}
}

func TestSpriteCellsRejectsRagged(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("ragged sprite did not panic")
		}
	}()
	spriteCells(sprite{"ab", "abc"})
}

func TestBuildingSpriteShapes(t *testing.T) {
	// Every metered facade sprite must be 16x16 at 4px-per-art-pixel =
	// one 2x2 block, dense enough to read as a building, with every inked
	// key mapped in its palette. The gate is shape-checked separately.
	for _, b := range buildingKinds {
		s := buildingSprite[b]
		wantW := 16
		if b == crawler.BldgGate {
			wantW = 24
		}
		if len(s) != 16 {
			t.Errorf("%s sprite has %d rows", townBuildingName(b), len(s))
			continue
		}
		for y, row := range s {
			if len(row) != wantW {
				t.Errorf("%s row %d is %d wide", townBuildingName(b), y, len(row))
			}
		}
		if n := len(spriteCells(s)); n < 90 {
			t.Errorf("%s sprite only %d/256 cells inked", townBuildingName(b), n)
		}
		for _, cell := range spriteCells(s) {
			if _, ok := buildingPal(b)[cell.k]; !ok {
				t.Errorf("%s key %q unmapped in its palette", townBuildingName(b), cell.k)
				break
			}
		}
	}
}

func TestGateSpriteShape(t *testing.T) {
	s := buildingSprite[crawler.BldgGate]
	if len(s) != 16 || len(s[0]) != 24 {
		t.Fatalf("gate sprite %dx%d, want 24x16 (three tiles wide, two tall)", len(s[0]), len(s))
	}
}

func TestBuildingRoofsCarryHue(t *testing.T) {
	// The roof keys must stay in the building's identity-hue family: the
	// dominant channel of each roof shade matches the identity color's, so
	// kinds remain tellable at a glance.
	for _, b := range buildingKinds {
		hue := buildingColor(b)
		pal := buildingPal(b)
		for _, k := range []byte{'r', 's', 't'} { // roof dark/mid/light
			got := pal[k]
			if dominantChan(hue) != dominantChanTol(got, 0.06) {
				t.Errorf("%s roof key %q (%+v) drifts from hue %+v", townBuildingName(b), k, got, hue)
			}
		}
	}
}

func TestPropSpriteShapes(t *testing.T) {
	for _, k := range []int{propTree, propPine, propWell, propStall, propCrate, propBarrel, propLamp} {
		s := propSprite[k]
		if len(s) != 8 || len(s[0]) != 8 {
			t.Errorf("prop %d sprite %dx%d, want 8x8", k, len(s[0]), len(s))
		}
		for _, cell := range spriteCells(s) {
			if _, ok := storyPalette[cell.k]; !ok {
				t.Errorf("prop %d key %q unmapped", k, cell.k)
				break
			}
		}
	}
}

func townOf(t *testing.T) *crawler.Town {
	t.Helper()
	return crawler.NewTown(crawler.DefaultMeta())
}

func newTownForProps() *crawler.Town { return crawler.NewTown(crawler.DefaultMeta()) }

func TestTownPropsStayOffWalkables(t *testing.T) {
	tt := newTownForProps()
	for _, prop := range townProps(tt) {
		cell := crawler.Cell{X: prop.x, Y: prop.y}
		if tt.BuildingAt(cell) != crawler.BldgNone {
			t.Errorf("prop %d at (%d,%d) sits on a building", prop.kind, prop.x, prop.y)
		}
		if isTrail(cell) {
			t.Errorf("prop %d at (%d,%d) squats the trail", prop.kind, prop.x, prop.y)
		}
		spawn := crawler.Cell{X: 10, Y: 4}
		if cell == spawn || chebyshev(cell, spawn) <= 1 {
			t.Errorf("prop %d at (%d,%d) crowds the spawn", prop.kind, prop.x, prop.y)
		}
	}
}

func TestTownPropsDeterministic(t *testing.T) {
	tt := newTownForProps()
	a := townProps(tt)
	b := townProps(tt)
	if len(a) != len(b) {
		t.Fatalf("prop pass unstable: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("prop %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
	if len(a) < 18 {
		t.Errorf("only %d props; the lawn should feel full", len(a))
	}
	var trees, pines int
	for _, p := range a {
		switch p.kind {
		case propTree:
			trees++
		case propPine:
			pines++
		}
	}
	if trees < 8 || pines < 5 {
		t.Errorf("greenery too thin: %d trees, %d pines", trees, pines)
	}
	// The well must hold its plaza cell.
	var well bool
	for _, p := range a {
		if p.kind == propWell && p.x == 8 && p.y == 9 {
			well = true
		}
	}
	if !well {
		t.Errorf("well missing from (8,9)")
	}
}

func TestTownPropsNoOverlap(t *testing.T) {
	tt := newTownForProps()
	seen := map[crawler.Cell]bool{}
	for _, prop := range townProps(tt) {
		c := crawler.Cell{X: prop.x, Y: prop.y}
		if seen[c] {
			t.Errorf("two props at (%d,%d)", prop.x, prop.y)
		}
		seen[c] = true
	}
}
