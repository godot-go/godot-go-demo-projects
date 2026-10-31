package demo

import (
	"fmt"
	"math"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// The town wears its meta progress on its face: a level-0 building is a
// lean-and-leak ruin, the first renown patches it whole, and a capped
// building earns trim, flowerboxes, chimney smoke, and a lantern. The
// shared ground decays from trampled mud to kept grass with the town's
// full upgrade fraction. Everything repaints per frame off c.now, so
// smoke drifts and lantern pulses stay alive between purchases.

// condition01 maps a building level to 0..1 kept-ness; an unmetered
// building (cap 0) reads as already kept.
func condition01(level, max int) float64 {
	if max <= 0 || level >= max {
		return 1
	}
	if level <= 0 {
		return 0
	}
	return float64(level) / float64(max)
}

// tile01 hashes a key onto [0,1) so scatter details (moss, tufts, pebbles,
// rubble) paste deterministically instead of dancing each frame.
func tile01(key string) float64 {
	h := 0
	for i := 0; i < len(key); i++ {
		h = h*31 + int(key[i])
	}
	return float64(h%997) / 997
}

// townWear reads the meta's whole upgrade ledger onto [0,1].
func townWear(m *crawler.Meta) float64 {
	have, max := m.UpgradeTotal()
	return condition01(have, max)
}

func lerpRGB(a, b rgb, k float64) rgb {
	f := float32(k)
	return rgb{a.r + (b.r-a.r)*f, a.g + (b.g-a.g)*f, a.b + (b.b-a.b)*f}
}

// townGround paints one open tile as lush-or-worn lawn: a base wash
// brightening with the town's whole upgrade fraction, then a per-cell
// dither of grass tones, tufts, clover, and tiny blooms. All scatter is
// hash-pasted, never dancing frame to frame.
func townGround(c *DungeonCrawler, x, y int, gx, gy float32, wear float64) {
	p := pen{c, false}
	ruin := lerpRGB(rgb{0.13, 0.15, 0.11}, rgb{0.24, 0.28, 0.15}, 0.4)
	lush := rgb{0.20, 0.38, 0.15}
	base := lerpRGB(ruin, lush, wear)
	p.fill(gx, gy, townTile, townTile, base)
	const cell = 16 // dither cell: 2x2 per tile
	for cy := 0; cy < 2; cy++ {
		for cx := 0; cx < 2; cx++ {
			h := tile01(fmt.Sprintf("gd%d,%d,%d", x, y, cy*2+cx))
			px, py := gx+float32(cx*cell), gy+float32(cy*cell)
			switch {
			case h < 0.16*wear+0.02:
				p.fill(px, py, cell, cell, lerpRGB(base, rgb{0.27, 0.45, 0.19}, 0.8))
			case h > 1-0.13*(1-wear)-0.03:
				p.fill(px, py, cell, cell, lerpRGB(base, rgb{0.13, 0.20, 0.10}, 0.8))
			}
		}
	}
	if wear > 0.34 {
		h := tile01(fmt.Sprintf("gb%d,%d", x, y))
		if h < 0.10 {
			bloom := [3]rgb{{0.92, 0.62, 0.72}, {0.95, 0.85, 0.45}, {0.62, 0.72, 0.95}}[int(h*300)%3]
			p.disc(gx+float32(5+22*float32(tile01("gq"+fmt.Sprint(x, ",", y)))), gy+float32(5+22*float32(tile01("gr"+fmt.Sprint(x, ",", y)))), 1.6, bloom)
		}
	}
}

// townPathElement tramples the wander trail: a warm dirt ribbon with
// hash-pasted ruts and pebble scatter as the town keeps itself.
func townPathElement(c *DungeonCrawler, x, y int, gx, gy float32, wear float64) {
	p := pen{c, false}
	base := lerpRGB(rgb{0.24, 0.18, 0.12}, rgb{0.52, 0.38, 0.26}, wear)
	p.fill(gx, gy, townTile, townTile, base)
	const cell = 16
	for cy := 0; cy < 2; cy++ {
		for cx := 0; cx < 2; cx++ {
			h := tile01(fmt.Sprintf("pd%d,%d,%d", x, y, cy*2+cx))
			px, py := gx+float32(cx*cell), gy+float32(cy*cell)
			switch {
			case h < 0.14:
				p.fill(px, py, cell, cell, lerpRGB(base, rgb{0.34, 0.24, 0.15}, 0.7))
			case h > 1-0.11*(1-wear)-0.02:
				p.fill(px, py, cell, cell, lerpRGB(base, rgb{0.60, 0.47, 0.34}, 0.7))
			}
		}
	}
	if wear > 0.5 {
		h := tile01(fmt.Sprintf("pp%d,%d", x, y))
		if h < 0.09 {
			p.disc(gx+float32(5+22*float32(tile01("pq"+fmt.Sprint(x, ",", y)))), gy+float32(5+22*float32(tile01("pr"+fmt.Sprint(x, ",", y)))), 1.9, rgb{0.66, 0.60, 0.52})
		}
	}
}

// plazCell is where the well holds court — the stone circle decorates
// that trail-adjacent tile. townProps pins the same constant below.
var plazCell = crawler.Cell{X: 8, Y: 9}

// townPlaza paints the stone circle under the well: a cobble roundel
// that keeps its joints regardless of wear.
func townPlaza(c *DungeonCrawler, gx, gy float32) {
	p := pen{c, false}
	const cell = 16
	for cy := 0; cy < 2; cy++ {
		for cx := 0; cx < 2; cx++ {
			h := tile01(fmt.Sprintf("pl%d", cy*2+cx))
			px, py := gx+float32(cx*cell), gy+float32(cy*cell)
			switch {
			case h < 0.4:
				p.fill(px, py, cell, cell, rgb{0.47, 0.46, 0.44})
			case h < 0.75:
				p.fill(px, py, cell, cell, rgb{0.55, 0.55, 0.53})
			default:
				p.fill(px, py, cell, cell, rgb{0.38, 0.38, 0.37})
			}
		}
	}
}

// townRect is a building's stamped footprint in tile coordinates.
type townRect struct{ x0, y0, x1, y1 int }

// townBuildingRects collects one bounding box per building kind. Stamped
// blocks are rectangular, so a min/max scan is exact.
func townBuildingRects(t *crawler.Town) map[crawler.TownBuilding]townRect {
	out := map[crawler.TownBuilding]townRect{}
	for y := 0; y < t.H; y++ {
		for x := 0; x < t.W; x++ {
			b := t.BuildingAt(crawler.Cell{X: x, Y: y})
			if b == crawler.BldgNone {
				continue
			}
			r, ok := out[b]
			if !ok {
				r = townRect{x0: x, y0: y, x1: x, y1: y}
			}
			if x < r.x0 {
				r.x0 = x
			}
			if x > r.x1 {
				r.x1 = x
			}
			if y < r.y0 {
				r.y0 = y
			}
			if y > r.y1 {
				r.y1 = y
			}
			out[b] = r
		}
	}
	return out
}

// buildingMax mirrors the rules' upgrade caps for the decay ladder.
func buildingMax(b crawler.TownBuilding) int {
	switch b {
	case crawler.BldgGate:
		return 0
	case crawler.BldgGuildFighter, crawler.BldgGuildCleric, crawler.BldgGuildWizard, crawler.BldgGuildRogue:
		return 2
	}
	return 3
}

// townBuildingName mirrors the rules' save-key names (the crawler's own
// name() is unexported), so facade code can read the meta's Buildings.
func townBuildingName(b crawler.TownBuilding) string {
	switch b {
	case crawler.BldgSmithy:
		return "Smithy"
	case crawler.BldgClinic:
		return "Clinic"
	case crawler.BldgLibrary:
		return "Library"
	case crawler.BldgArchive:
		return "Archive"
	case crawler.BldgGuildFighter:
		return "Guild Fighter"
	case crawler.BldgGuildCleric:
		return "Guild Cleric"
	case crawler.BldgGuildWizard:
		return "Guild Wizard"
	case crawler.BldgGuildRogue:
		return "Guild Rogue"
	case crawler.BldgGate:
		return "Gate"
	}
	return ""
}

// townBox maps a building's stamped tile rect onto the pixel box its
// facade must cover, through the town view's centering offsets.
func townBox(r townRect, ox, oy float32, tile int) (bx, by, bw, bh float32) {
	bx = ox + float32(r.x0*tile)
	by = oy + float32(r.y0*tile)
	bw = (float32(r.x1-r.x0) + 1) * float32(tile)
	bh = (float32(r.y1-r.y0) + 1) * float32(tile)
	return
}

// drawTownFacade blits one building's pixel-art front, wearing its decay
// tier: a ruin (roof holes, boards, cracks), a patched first step, or a
// kept building with awning, flowerboxes, chimney smoke, and a lantern.
// The unmetered gate rides the whole-town wear level instead.
func drawTownFacade(c *DungeonCrawler, b crawler.TownBuilding, r townRect, ox, oy float32, m *crawler.Meta, t float64) {
	p := pen{c, false}
	bx, by, bw, bh := townBox(r, ox, oy, townTile)
	ph := fxPhase(fmt.Sprintf("b%d", int(b)))

	cond := condition01(m.Buildings[townBuildingName(b)], buildingMax(b))
	if b == crawler.BldgGate {
		cond = townWear(m) // the gate is kept when the town is
	}
	tier := tetier(cond)
	s, pal := facadeTier(b, tier)
	blit(c, s, bx, by, pal, artPx)

	if b == crawler.BldgGate {
		if tier == 0 { // guttering torches in the dark arch
			for _, dx := range []float32{0.30, 0.70} {
				tx, ty := bx+bw*dx, by+bh*0.72
				p.glow(tx, ty, 3.6, rgb{0.95, 0.55, 0.2}, 0.4+0.3*pulse01(t, ph, 1.3))
			}
		}
		return
	}

	switch tier {
	case 0: // rubble and weeds strewn around the ruin's base
		for k := 0; k < 3; k++ {
			if tile01(fmt.Sprintf("rb%d,%d", int(b), k)) < 0.75 {
				rx := bx + bw*float32(0.08+0.28*tile01(fmt.Sprintf("rx%d,%d", int(b), k)))
				p.disc(rx, by+bh-float32(2+k), 1.7, rgb{0.30, 0.29, 0.28})
			}
		}
		wxx := bx + bw*float32(0.15+0.7*tile01(fmt.Sprintf("wd%d", int(b))))
		p.fill(wxx, by+bh-4, 2, 5, rgb{0.11, 0.14, 0.08})
		p.fill(wxx+5, by+bh-5, 2, 6, rgb{0.10, 0.13, 0.07})
	case 1: // a warm hearth inside the patched building
		p.glow(bx+16, by+38, 5, rgb{0.92, 0.72, 0.34}, 0.16+0.08*pulse01(t, ph, 0.9))
	case 2: // chimney smoke and a pulsing lantern by the door
		for k := 0; k < 2; k++ {
			age := math.Mod(t*0.16+ph*0.11+float64(k)*0.5, 1)
			p.glow(bx+52+3*float32(math.Sin(t*0.9+ph+float64(k))), by+6-13*float32(age), 2.2+2.2*float32(age), rgb{0.55, 0.55, 0.58}, 0.30*(1-age))
		}
		p.glow(bx+50, by+54, 5, gold, 0.35+0.35*pulse01(t, ph, 0.8))
	}
}
