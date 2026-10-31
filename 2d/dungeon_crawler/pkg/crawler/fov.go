package crawler

// ComputeFOV returns the tiles visible from origin within radius (Euclidean),
// using permissive Bresenham line-of-sight: a wall between origin and target
// blocks sight, except a purely diagonal step through a corner gap is allowed
// when at least one side tile is open. Solid tiles themselves can be seen.
func ComputeFOV(f *Floor, origin Cell, radius int) map[Cell]bool {
	vis := map[Cell]bool{origin: true}
	for y := max(0, origin.Y-radius); y <= min(f.H-1, origin.Y+radius); y++ {
		for x := max(0, origin.X-radius); x <= min(f.W-1, origin.X+radius); x++ {
			c := Cell{x, y}
			dx, dy := c.X-origin.X, c.Y-origin.Y
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			if hasLineOfSight(f, origin, c) {
				vis[c] = true
			}
		}
	}
	return vis
}

// hasLineOfSight walks the Bresenham line from a to b; intermediate solid
// tiles block. The endpoint may be solid (walls are visible).
func hasLineOfSight(f *Floor, a, b Cell) bool {
	if a == b {
		return true
	}
	x0, y0 := a.X, a.Y
	dx, dy := abs(b.X-x0), -abs(b.Y-y0)
	sx, sy := step(x0, b.X), step(y0, b.Y)
	err := dx + dy

	for {
		e2 := 2 * err
		prevX, prevY := x0, y0
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
		if x0 == b.X && y0 == b.Y {
			return true
		}
		if f.IsSolid(Cell{x0, y0}) {
			// Permissive corner: a diagonal step is survivable when either
			// adjacent side tile is open (walking along a wall).
			diagonal := x0 != prevX && y0 != prevY
			if !diagonal {
				return false
			}
			sideOpen := !f.IsSolid(Cell{x0, prevY}) || !f.IsSolid(Cell{prevX, y0})
			if !sideOpen {
				return false
			}
		}
	}
}

// MarkExplored records the visible set in the floor's memory.
func (f *Floor) MarkExplored(visible map[Cell]bool) {
	for c := range visible {
		f.Explored[f.index(c)] = true
	}
}

// ExploredAt reports whether a tile has ever been seen.
func (f *Floor) ExploredAt(c Cell) bool {
	return f.InBounds(c) && f.Explored[f.index(c)]
}
