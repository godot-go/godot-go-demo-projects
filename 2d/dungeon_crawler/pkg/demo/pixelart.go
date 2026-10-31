package demo

import (
	"bytes"
	"fmt"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// The town reads as authored pixel art: facade and prop sprites are
// laid out off shared 16/24-column grid rules, inked cell-by-cell through
// the polygon path. Roof hue resolves per building kind, and the decay
// ladder edits walls before blitting — condition is a pixel edit, not a
// redraw of geometry.

// sprite is a pixel-art grid: one equal-length row per string, '.' or
// ' ' = transparent, any other byte = a palette key.
type sprite []string

// spriteCell is one inked pixel in sprite space.
type spriteCell struct {
	dx, dy int
	k      byte
}

// spriteCells rasterizes a sprite top-to-bottom, left-to-right. Ragged
// sprites are programmer errors and panic loudly.
func spriteCells(s sprite) []spriteCell {
	w := len(s[0])
	for _, row := range s {
		if len(row) != w {
			panic("ragged sprite row")
		}
	}
	cells := make([]spriteCell, 0, w*len(s))
	for dy, row := range s {
		for dx := 0; dx < w; dx++ {
			if k := row[dx]; k != '.' && k != ' ' {
				cells = append(cells, spriteCell{dx, dy, k})
			}
		}
	}
	return cells
}

// storyPalette is the shared color battery: walls, beams, glass, stone,
// foliage, awnings. The roof keys (r/s/t/u) resolve per building below.
var storyPalette = map[byte]rgb{
	'o': {0.08, 0.07, 0.06}, // holes / shadowed voids
	'e': {0.93, 0.87, 0.74}, // plaster light
	'E': {0.80, 0.72, 0.58}, // plaster shade
	'f': {0.30, 0.21, 0.13}, // beam brown
	'F': {0.46, 0.32, 0.19}, // wood mid (boards, tabletops)
	'd': {0.30, 0.19, 0.11}, // door
	'D': {0.19, 0.12, 0.07}, // door dark
	'k': {0.95, 0.80, 0.30}, // brass / gold accent
	'l': {0.14, 0.17, 0.22}, // window glass dark
	'v': {0.55, 0.40, 0.26}, // window frame / grey iron accent
	'm': {0.55, 0.55, 0.53}, // stone mid
	'M': {0.70, 0.70, 0.68}, // stone light
	'n': {0.38, 0.38, 0.37}, // stone dark
	'T': {0.24, 0.42, 0.18}, // leaf mid
	'Y': {0.42, 0.62, 0.28}, // leaf light
	'y': {0.15, 0.28, 0.13}, // leaf shadow
	'P': {0.13, 0.30, 0.22}, // pine mid
	'Q': {0.09, 0.22, 0.16}, // pine dark
	'a': {0.30, 0.55, 0.72}, // water
	'A': {0.80, 0.30, 0.24}, // red accent (awning, cross, banner)
	'B': {0.90, 0.86, 0.78}, // cream accent (awning stripe)
	'x': {0.92, 0.90, 0.84}, // bright accent (sign board)
	'q': {0.50, 0.32, 0.66}, // violet accent (scrolls)
	'i': {0.98, 0.82, 0.40}, // lit glass
}

// buildingPal resolves a building's full ink palette: the story battery
// plus its roof keys shaded from that building's identity hue, so kinds
// stay tellable at a glance by roof color. Cash keeps its keys keyed on
// the story battery.
func buildingPal(b crawler.TownBuilding) map[byte]rgb {
	pal := make(map[byte]rgb, len(storyPalette)+4)
	for k, v := range storyPalette {
		pal[k] = v
	}
	hue := buildingColor(b)
	pal['r'] = clampRGB(shade(hue, 0.55))
	pal['s'] = clampRGB(shade(hue, 0.78))
	pal['t'] = lerpRGB(hue, rgb{1, 1, 1}, 0.25)
	pal['u'] = clampRGB(shade(hue, 0.30))
	return pal
}

func clampRGB(v rgb) rgb {
	c := func(f float32) float32 {
		if f > 1 {
			return 1
		}
		return f
	}
	return rgb{c(v.r), c(v.g), c(v.b)}
}

// signGlyphs are the 3-wide, 2-tall sign motifs stamped onto each
// building's wall board.
var signGlyphs = map[crawler.TownBuilding][2]string{
	crawler.BldgSmithy:       {"Akk", "kAA"},
	crawler.BldgClinic:       {"kA.", ".AA"},
	crawler.BldgLibrary:      {"kTk", "kTk"},
	crawler.BldgArchive:      {"qxq", "xqq"},
	crawler.BldgGuildFighter: {"kk.", ".Ak"},
	crawler.BldgGuildCleric:  {".k.", "kkk"},
	crawler.BldgGuildWizard:  {"All", "llA"},
	crawler.BldgGuildRogue:   {"qk.", ".kq"},
}

// buildingKinds is the facade roster in a stable order.
var buildingKinds = []crawler.TownBuilding{
	crawler.BldgGate,
	crawler.BldgSmithy, crawler.BldgClinic, crawler.BldgLibrary,
	crawler.BldgArchive,
	crawler.BldgGuildFighter, crawler.BldgGuildCleric,
	crawler.BldgGuildWizard, crawler.BldgGuildRogue,
}

// facadeSprite generates one 16x16 cottage front. Rows 0-6 are a
// shingled gable with the ridge seam holding the center; row 7 the
// eaves board; the wall carries a shade cap, two framed windows, a
// sign board with the kind's glyph, a door, and a stone foundation.
func facadeSprite(g [2]string) sprite {
	s := make(sprite, 16)
	for y := 0; y < 16; y++ {
		var sb bytes.Buffer
		switch {
		case y < 7: // gable roof: width grows downward, seam at center
			half := y + 1
			for x := 0; x < 16; x++ {
				if x >= 8-half && x <= 7+half {
					switch {
					case y < 2:
						sb.WriteByte('r')
					case y < 5:
						sb.WriteByte('s')
					default:
						sb.WriteByte('t')
					}
				} else {
					sb.WriteByte('.')
				}
			}
			// punch the ridge seam over the shingles
			row := []byte(sb.String())
			row[7], row[8] = 'u', 'u'
			sb.Reset()
			sb.Write(row)
		case y == 7:
			sb.Write(bytes.Repeat([]byte{'f'}, 16))
		case y == 8:
			sb.WriteByte('f')
			sb.Write(bytes.Repeat([]byte{'E'}, 14))
			sb.WriteByte('f')
		case y < 11: // upper wall: two framed windows
			sb.WriteString("fevllveeeevllvef")
		case y < 12: // plain shade strip between windows and the sign
			sb.WriteByte('f')
			sb.WriteString("Eeeeeeeeeeeeee")
			sb.WriteByte('f')
		case y < 14: // sign board over the door head
			sb.WriteString("fex")
			sb.WriteString(g[y-12])
			sb.WriteString("xeddd")
			sb.WriteString("eeee")
			sb.WriteByte('f')
		case y < 15: // waist with the door continuing through
			sb.WriteByte('f')
			sb.WriteString("eeeeeee")
			sb.WriteString("ddd")
			sb.WriteString("eeee")
			sb.WriteByte('f')
		default: // foundation stones
			sb.WriteByte('f')
			sb.WriteString("nnmMnnmMnnmMnn")
			sb.WriteByte('f')
		}
		s[y] = sb.String()
	}
	return s
}

// buildingSprite holds one generated front per kind; the gate is its own
// wider stone sprite (three tiles).
var buildingSprite = map[crawler.TownBuilding]sprite{}

func init() {
	for b, g := range signGlyphs {
		buildingSprite[b] = facadeSprite(g)
	}
	buildingSprite[crawler.BldgGate] = gateSprite()
	buildFacadeTiers()
}

// gateSprite generates the 24x16 gatehouse: crenellated battlements,
// stone cheeks with a timber lintel, a red banner pinned with a brass
// sword, and a dark arch between flanking pillars.
func gateSprite() sprite {
	s := make(sprite, 16)
	for y := 0; y < 16; y++ {
		var sb bytes.Buffer
		switch {
		case y < 2: // battlement crown: merlons every third stone
			for x := 0; x < 24; x++ {
				if x%3 == 0 {
					sb.WriteByte('M')
				} else {
					sb.WriteByte('.')
				}
			}
		case y < 3:
			sb.Write(bytes.Repeat([]byte{'n'}, 24))
		default: // stone body (y 3..15)
			for x := 0; x < 24; x++ {
				switch {
				case x == 0 || x == 23:
					sb.WriteByte('n')
				case y >= 10 && y <= 14 && x >= 4 && x <= 19:
					// arch opening: widens with depth, outline pillars
					if x == 4 || x == 19 {
						sb.WriteByte('n')
					} else {
						sb.WriteByte('o')
					}
				case y == 15:
					sb.WriteByte('n')
				case y == 3 || y == 8: // two timber bands
					sb.WriteByte('f')
				case y >= 4 && y <= 7 && x >= 10 && x <= 13: // banner
					if x == 11 || x == 12 {
						sb.WriteByte('k')
					} else {
						sb.WriteByte('A')
					}
				default:
					if (x+y)%4 == 0 {
						sb.WriteByte('m')
					} else if x%5 == 2 {
						sb.WriteByte('M')
					} else {
						sb.WriteByte('n')
					}
				}
			}
		}
		s[y] = sb.String()
	}
	return s
}

// --- town props ---

const (
	propTree = iota
	propPine
	propWell
	propStall
	propCrate
	propBarrel
	propLamp
)

var propSprite = map[int]sprite{
	propTree:   treeSprite(),
	propPine:   pineSprite(),
	propWell:   wellSprite(),
	propStall:  stallSprite(),
	propCrate:  crateSprite(),
	propBarrel: barrelSprite(),
	propLamp:   lampSprite(),
}

// treeSprite: a leafy round canopy with a lit crown and a wood trunk.
func treeSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		for x := 0; x < 8; x++ {
			dx, dy := float32(x)-3.5, float32(y)-3.0
			d := dx*dx + dy*dy
			switch {
			case d <= 9.9 && y < 6:
				row[x] = 'T'
				if dy < -1.0 && d <= 6.0 {
					row[x] = 'Y'
				}
				if d > 7.0 {
					row[x] = 'y'
				}
			case x >= 3 && x <= 4 && y >= 5:
				row[x] = 'f'
			}
		}
		s[y] = string(row)
	}
	return s
}

// pineSprite: a stacked cone of boughs over a short trunk.
func pineSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		for x := 0; x < 8; x++ {
			w := float32(8-y) * 0.42
			if y == 7 {
				w = 1.0
			}
			if 3.5-w <= float32(x) && float32(x) <= 3.5+w {
				if y == 7 {
					row[x] = 'f'
				} else if y%2 == 1 {
					row[x] = 'Q'
				} else {
					row[x] = 'P'
				}
			}
		}
		s[y] = string(row)
	}
	return s
}

// wellSprite: a stone ring roofed by small posts, water dark within.
func wellSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		for x := 0; x < 8; x++ {
			dx, dy := float32(x)-3.5, float32(y)-3.0
			d := dx*dx + dy*dy
			switch {
			case y == 0:
				if x >= 2 && x <= 5 {
					row[x] = 'f'
				}
			case d <= 5.0 && y < 6:
				if d <= 2.6 {
					row[x] = 'a'
				} else {
					row[x] = 'm'
				}
			case y >= 6:
				row[x] = 'n'
			}
		}
		s[y] = string(row)
	}
	return s
}

// stallSprite: a striped awning over a wood counter with wares.
func stallSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		for x := 0; x < 8; x++ {
			switch {
			case y < 2:
				if x%2 == 0 {
					row[x] = 'A'
				} else {
					row[x] = 'B'
				}
			case y == 2:
				row[x] = 'f'
			case y == 3:
				if x%2 == 0 {
					row[x] = 'k'
				} else if x%3 == 0 {
					row[x] = 'T'
				} else {
					row[x] = 'F'
				}
			case y < 6:
				row[x] = 'F'
			case x == 0 || x == 7:
				row[x] = 'f'
			}
		}
		s[y] = string(row)
	}
	return s
}

// crateSprite: a strapped packing box.
func crateSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		if y >= 1 && y <= 6 {
			for x := 1; x < 7; x++ {
				row[x] = 'F'
			}
			row[1], row[6] = 'f', 'f' // side rails
			row[3], row[4] = 'd', 'd' // iron strapping
			row[2], row[5] = 'f', 'f'
		}
		s[y] = string(row)
	}
	return s
}

// barrelSprite: corded staves with metal hoops.
func barrelSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		for x := 0; x < 8; x++ {
			dx := float32(x) - 3.5
			half := float32(3.0)
			if y == 0 || y == 7 {
				half = 2.2
			}
			if dx >= -half && dx <= half {
				switch {
				case y == 1 || y == 5:
					row[x] = 'v'
				case x <= 1 || x >= 6:
					row[x] = 'f'
				default:
					row[x] = 'F'
				}
			}
		}
		s[y] = string(row)
	}
	return s
}

// lampSprite: a wrought lamp post with a gold head.
func lampSprite() sprite {
	s := make(sprite, 8)
	for y := 0; y < 8; y++ {
		row := bytes.Repeat([]byte{'.'}, 8)
		switch {
		case y <= 1:
			row[3], row[4] = 'k', 'k'
		case y < 7:
			row[3] = 'f'
		default:
			row[2], row[3], row[4] = 'n', 'n', 'n'
		}
		s[y] = string(row)
	}
	return s
}

// blit inks one sprite at a pixel origin through the polygon path. px is
// the world size of one art pixel. Horizontal runs of identical pixels
// merge into one fill, so flat art costs few draws, not one per cell.
func blit(c *DungeonCrawler, s sprite, gx, gy float32, pal map[byte]rgb, px float32) {
	w := len(s[0])
	p := pen{c, false}
	for dy, row := range s {
		for run := 0; run < w; {
			k := row[run]
			if k == '.' || k == ' ' {
				run++
				continue
			}
			end := run + 1
			for end < w && row[end] == k {
				end++
			}
			p.fill(gx+float32(run)*px, gy+float32(dy)*px, float32(end-run)*px, px, pal[k])
			run = end
		}
	}
}

// artPx is the world size of one art pixel: 16-column facades span one
// 2x2 town block (32px tiles), 8-column props one tile.
const artPx = 4

// tetier snaps a kept-ness fraction onto the decay ladder: 0 ruin, 1
// patched, 2 kept.
func tetier(cond float64) int {
	switch {
	case cond < 0.34:
		return 0
	case cond < 0.85:
		return 1
	}
	return 2
}

// editSprite deep-copies a sprite into writable rows.
func editSprite(s sprite) [][]byte {
	out := make([][]byte, len(s))
	for y, row := range s {
		out[y] = []byte(row)
	}
	return out
}

// facadeTiers caches each building's three decay-tier sprites (ruin /
// patched / kept) and palettes, edited once at init from the base art.
type btier struct {
	b    crawler.TownBuilding
	tier int
}

var (
	facadeTiers = map[btier]sprite{}
	facadePals  = map[btier]map[byte]rgb{}
)

// tierFade dims the whole palette as the building slips into ruin, so
// derelict facades read as one washed-out photograph of themselves.
var tierFade = [3]float32{0.60, 0.86, 1.0}

// facadeTier returns the cached sprite + palette for one building/tier.
func facadeTier(b crawler.TownBuilding, tier int) (sprite, map[byte]rgb) {
	return facadeTiers[btier{b, tier}], facadePals[btier{b, tier}]
}

// buildFacadeTiers runs after the base sprites exist: it edits the grid
// per tier — ruin punches roof holes, boards one window and shatters the
// other, and cracks the plaster; patched lights a pane; kept adds the
// awning, flowerboxes, a lantern, and a stone chimney — and shades the
// palette by tierFade.
func buildFacadeTiers() {
	for _, b := range buildingKinds {
		base := buildingSprite[b]
		for tier := 0; tier < 3; tier++ {
			g := editSprite(base)
			switch tier {
			case 0:
				roofHole(g, base, b)
				boardWindow(g, 3, 9, 'f')  // left pane boarded
				boardWindow(g, 11, 9, 'o') // right pane shattered dark
				wallCracks(g, b)
				if b == crawler.BldgGate {
					gateHoles(g, base)
				}
			case 1:
				boardWindow(g, 3, 9, 'i') // lamplight moved in
			default:
				for x := 1; x <= 14; x++ { // awning stripes under the eaves
					if (x-1)/2%2 == 0 {
						g[7][x] = 'A'
					} else {
						g[7][x] = 'B'
					}
				}
				g[11][3], g[11][4] = 'A', 'Y' // flowerboxes below the windows
				g[11][11], g[11][12] = 'A', 'Y'
				g[13][12] = 'k' // lantern beside the door head
				for _, at := range [][2]int{{12, 2}, {13, 2}, {12, 3}, {13, 3}, {12, 4}, {13, 4}} {
					g[at[1]][at[0]] = 'm' // stone chimney stack
				}
				g[3][13], g[4][12], g[4][13] = 'n', 'n', 'n'
			}
			facadeTiers[btier{b, tier}] = func(g [][]byte) sprite {
				rows := make(sprite, len(g))
				for y, row := range g {
					rows[y] = string(row)
				}
				return rows
			}(g)
			pal := buildingPal(b)
			if tierFade[tier] < 1 {
				f := tierFade[tier]
				for k, v := range pal {
					pal[k] = shade(v, f)
				}
			}
			facadePals[btier{b, tier}] = pal
		}
	}
}

// roofHole punches hash-scattered 'o' holes through gable cells.
func roofHole(g [][]byte, base sprite, b crawler.TownBuilding) {
	for y := 0; y <= 6; y++ {
		for x := 0; x < len(base[y]); x++ {
			switch base[y][x] {
			case 'r', 's', 't':
				if tile01(fmt.Sprintf("rh%d,%d,%d", int(b), x, y)) < 0.06 {
					g[y][x] = 'o'
				}
			}
		}
	}
}

// boardWindow replaces the two pane pixels of a facade window run — cols
// x0/x0+1 across rows y/y+1 — with one ink byte (board, void, or light).
func boardWindow(g [][]byte, x0, y0 int, k byte) {
	for y := y0; y < y0+2; y++ {
		g[y][x0], g[y][x0+1] = k, k
	}
}

// wallCracks streaks three fixed plaster cells dark.
func wallCracks(g [][]byte, b crawler.TownBuilding) {
	g[11][6], g[14][5], g[14][13] = 'o', 'o', 'o'
}

// gateHoles pits the gate's stone face when the town decayed.
func gateHoles(g [][]byte, base sprite) {
	for y := 3; y <= 15; y++ {
		for x := 0; x < len(base[y]); x++ {
			switch base[y][x] {
			case 'm', 'M', 'n':
				if tile01(fmt.Sprintf("gh%d,%d,%d", x, y, 0)) < 0.05 {
					g[y][x] = 'o'
				}
			}
		}
	}
}

// drawTownProps blits the lawn's dressing ahead of the facades: trees,
// pines, the plaza well, the market clutter, and the gate lamps, with
// lamp and water glows breathing off the shared clock.
func drawTownProps(c *DungeonCrawler, t *crawler.Town, ox, oy float32, m *crawler.Meta, now float64) {
	for _, prop := range townProps(t) {
		gx, gy := ox+float32(prop.x*townTile), oy+float32(prop.y*townTile)
		blit(c, propSprite[prop.kind], gx, gy, storyPalette, artPx)
		p := pen{c, false}
		ph := fxPhase(fmt.Sprintf("p%d,%d", prop.x, prop.y))
		switch prop.kind {
		case propLamp: // a lamp-lit pool on the walk
			p.glow(gx+16, gy+8, 7, gold, 0.22+0.12*pulse01(now, ph, 0.75))
		case propWell: // a faint aqua shimmer over the water
			p.glow(gx+14, gy+14, 6, rgb{0.45, 0.75, 0.9}, 0.10+0.06*pulse01(now, ph, 0.5))
		}
	}
}

// isTrail mirrors the shell's wander trail rule for prop placement.
func isTrail(c crawler.Cell) bool {
	return (c.X == 10 && c.Y >= 5 && c.Y <= 11) || (c.Y == 5 && c.X >= 2 && c.X <= 18)
}

// chebyshev is the roam-safe distance metric for prop clearance.
func chebyshev(a, b crawler.Cell) int {
	dx, dy := a.X-b.X, a.Y-b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// propAnchor pins one prop to a tile.
type propAnchor struct {
	kind int
	x, y int
}

// townProps lays out the lawn's dressing deterministically: the plaza
// well, a market stall and clutter by the Smithy frontage, lamps
// flanking the gate, and a tree/pine scatter across every free lawn tile
// past the trail, the buildings (with a one-tile moat), and the spawn.
func townProps(t *crawler.Town) []propAnchor {
	occupied := map[crawler.Cell]bool{}
	var out []propAnchor
	add := func(kind, x, y int) {
		c := crawler.Cell{X: x, Y: y}
		if occupied[c] || t.BuildingAt(c) != crawler.BldgNone || isTrail(c) {
			return
		}
		occupied[c] = true
		out = append(out, propAnchor{kind, x, y})
	}
	add(propWell, 8, 9)  // plaza centerpiece by the trail bend
	add(propStall, 2, 3) // market clatter before the Smithy
	add(propCrate, 4, 3)
	add(propBarrel, 3, 3)
	add(propLamp, 9, 11) // lamps frame the gate walk
	add(propLamp, 11, 11)
	spawn := crawler.Cell{X: 10, Y: 4}
	for y := 0; y < t.H; y++ {
		for x := 0; x < t.W; x++ {
			c := crawler.Cell{X: x, Y: y}
			if occupied[c] || t.BuildingAt(c) != crawler.BldgNone || isTrail(c) {
				continue
			}
			if chebyshev(c, spawn) <= 1 {
				continue
			}
			adjacent := false
			for _, d := range [8]crawler.Cell{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}} {
				n := crawler.Cell{X: c.X + d.X, Y: c.Y + d.Y}
				if t.BuildingAt(n) != crawler.BldgNone {
					adjacent = true
					break
				}
			}
			if adjacent {
				continue
			}
			h := tile01(fmt.Sprintf("tr%d,%d", x, y))
			switch {
			case h < 0.09:
				add(propPine, x, y)
			case h < 0.215:
				add(propTree, x, y)
			}
		}
	}
	return out
}

// dominantChan reports which channel of a color is largest — 0 red,
// 1 green, 2 blue. Ties go to the earlier channel.
func dominantChan(v rgb) int {
	if v.r >= v.g && v.r >= v.b {
		return 0
	}
	if v.g >= v.b {
		return 1
	}
	return 2
}

// dominantChanTol reports the dominant channel, collapsing near-ties
// within tol so palette shadings do not flunk the hue-family pin.
func dominantChanTol(v rgb, tol float32) int {
	hi := func(a, b float32) bool { return a > b+tol }
	switch {
	case hi(v.r, v.g) && hi(v.r, v.b):
		return 0
	case hi(v.g, v.b) || (v.g >= v.b && !hi(v.b, v.g)):
		return 1
	}
	return 2
}
