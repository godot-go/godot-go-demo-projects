package demo

import (
	"bytes"
	"fmt"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// The dungeon floor wears the same pixel-art ink as the town: each run
// tile is a 7x7 grid of 3px art cells expressed as merged run rectangles,
// walls read as top-lit masonry in a running bond, floors as slabs sharing
// their top and left joints, and hash-scattered moss, cracks, rubble, and
// bones break the pattern. Explored memory dims through the pen, not the
// artwork.

const (
	runGrid = 7 // art cells across a run tile
	runPx   = 3 // world pixels per art cell (7*3 = 21)
)

// artRun is one inked rectangle in a run tile's art space: cheap to fill,
// merged out of the grid so flat regions cost one draw, not forty-nine.
type artRun struct {
	dx, dy, w, h int
	k            byte
}

// dungeonPalette layers the run's ground keys over the story battery.
var dungeonPalette = func() map[byte]rgb {
	pal := make(map[byte]rgb, len(storyPalette)+6)
	for k, v := range storyPalette {
		pal[k] = v
	}
	pal['W'] = rgb{0.27, 0.26, 0.30} // wall brick
	pal['w'] = rgb{0.20, 0.19, 0.23} // mortar
	pal['V'] = rgb{0.35, 0.34, 0.39} // lit facing edge
	pal['S'] = rgb{0.47, 0.44, 0.38} // floor slab
	pal['z'] = rgb{0.37, 0.34, 0.29} // slab joint
	pal['Z'] = rgb{0.53, 0.50, 0.44} // slab light speckle
	return pal
}()

// rleRuns merges a grid into row-ordered ink runs; transparent cells drop
// out. Runs keep grid order, so later runs paint over earlier ones.
func rleRuns(g [][]byte) []artRun {
	runs := make([]artRun, 0, runGrid*runGrid)
	for dy := 0; dy < len(g); dy++ {
		row := g[dy]
		for dx := 0; dx < runGrid; {
			k := row[dx]
			if k == '.' {
				dx++
				continue
			}
			end := dx + 1
			for end < runGrid && row[end] == k {
				end++
			}
			runs = append(runs, artRun{dx, dy, end - dx, 1, k})
			dx = end
		}
	}
	return runs
}

func newGrid() [][]byte {
	g := make([][]byte, runGrid)
	for y := range g {
		g[y] = bytes.Repeat([]byte{'.'}, runGrid)
	}
	return g
}

// dungeonWallRuns masons one wall tile: a lit facing edge over two
// running-bond courses between full mortar rows. Hash scatter grows moss
// on a brick or opens a crack, only inside the brick courses so the bond
// reads clean everywhere.
func dungeonWallRuns(cell crawler.Cell) []artRun {
	g := newGrid()
	for x := 0; x < runGrid; x++ {
		g[0][x] = 'V'
		g[3][x] = 'w'
		g[6][x] = 'w'
	}
	for y := 1; y <= 2; y++ {
		for x := 0; x < runGrid; x++ {
			if x == 3 {
				g[y][x] = 'w'
			} else {
				g[y][x] = 'W'
			}
		}
	}
	for y := 4; y <= 5; y++ {
		for x := 0; x < runGrid; x++ {
			if x == 1 {
				g[y][x] = 'w'
			} else {
				g[y][x] = 'W'
			}
		}
	}
	if tile01(fmt.Sprintf("wm%d,%d", cell.X, cell.Y)) < 0.16 { // moss clumps
		mx := 1 + int(4*h01(fmt.Sprintf("wmx%d,%d", cell.X, cell.Y)))
		my := 1 + int(4*h01(fmt.Sprintf("wmy%d,%d", cell.X, cell.Y)))
		if my > 2 {
			my += 2 // skip the mortar course row
		}
		g[my][mx], g[my][mx+1] = 'T', 'y'
	}
	if tile01(fmt.Sprintf("wc%d,%d", cell.X, cell.Y)) < 0.07 { // hairline crack
		cx := 2 + int(3*h01(fmt.Sprintf("wcx%d,%d", cell.X, cell.Y)))
		cy := 1 + int(4*h01(fmt.Sprintf("wcy%d,%d", cell.X, cell.Y)))
		if cy > 2 {
			cy += 2 // skip the mortar course row
		}
		g[cy][cx], g[cy][cx+1] = 'o', 'o'
	}
	return rleRuns(g)
}

// h01 hashes decor coordinates; FNV-1a mixing keeps short keys like
// "wmx2,3" from clustering the way the 31-multiplier hash does.
func h01(key string) float64 {
	h := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return float64(h%9173) / 9173
}

// dungeonFloorRuns paves one floor tile: slabs sharing their top and left
// joints, a hash-pasted light scuff patch, and rarer rubble or a bone.
func dungeonFloorRuns(cell crawler.Cell) []artRun {
	g := newGrid()
	for x := 0; x < runGrid; x++ {
		g[0][x] = 'z'
	}
	for y := 0; y < runGrid; y++ {
		g[y][0] = 'z'
	}
	for y := 1; y < runGrid; y++ {
		for x := 1; x < runGrid; x++ {
			g[y][x] = 'S'
		}
	}
	salts := []string{"fs", "fr", "fb", "fc"}
	for i, salt := range salts {
		if i == 0 { // the scuff patch is common pastes
			if tile01(fmt.Sprintf("fs%d,%d", cell.X, cell.Y)) < 0.55 {
				sx := 1 + int(4*h01(fmt.Sprintf("fsx%d,%d", cell.X, cell.Y)))
				sy := 1 + int(4*h01(fmt.Sprintf("fsy%d,%d", cell.X, cell.Y)))
				g[sy][sx], g[sy][sx+1] = 'Z', 'Z'
				g[sy+1][sx] = 'Z'
			}
			continue
		}
		if tile01(fmt.Sprintf("%s%d,%d", salt, cell.X, cell.Y)) < 0.05 { // rubble / bone / crack
			switch salt {
			case "fr": // a little stone cluster
				dy := 1 + int(4*h01(fmt.Sprintf("fry%d,%d", cell.X, cell.Y)))
				dx := 1 + int(4*h01(fmt.Sprintf("frx%d,%d", cell.X, cell.Y)))
				g[dy][dx], g[dy][dx+1], g[dy+1][dx] = 'n', 'm', 'm'
			case "fb": // a dropped bone
				dy := 1 + int(4*h01(fmt.Sprintf("fby%d,%d", cell.X, cell.Y)))
				dx := 1 + int(4*h01(fmt.Sprintf("fbx%d,%d", cell.X, cell.Y)))
				g[dy][dx], g[dy][dx+1] = 'x', 'x'
				g[dy+1][dx-1] = 'E'
			default: // a slab crack
				dy := 1 + int(4*h01(fmt.Sprintf("fcy%d,%d", cell.X, cell.Y)))
				dx := 2 + int(3*h01(fmt.Sprintf("fcx%d,%d", cell.X, cell.Y)))
				g[dy][dx], g[dy][dx+1] = 'o', 'o'
			}
			break // at most one curiosity per tile
		}
	}
	return rleRuns(g)
}
