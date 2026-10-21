// Package snake holds the rules for the snake demo: grid movement on a fixed
// tick, growth by one segment per pellet, and a run won at a pellet target.
//
// The package knows nothing about Godot. A caller drives it by calling Tick
// once per timer timeout and Turn when the player asks for a direction.
package snake

import "math/rand"

// Cell is a board position, with (0,0) at the top-left.
type Cell struct{ X, Y int }

// Direction is one of the four grid directions.
type Direction int

const (
	Right Direction = iota
	Down
	Left
	Up
)

// Opposite returns the direction 180 degrees away.
func (d Direction) Opposite() Direction { return (d + 2) % 4 }

// delta is the one-cell offset for a direction.
func (d Direction) delta() Cell {
	switch d {
	case Right:
		return Cell{1, 0}
	case Down:
		return Cell{0, 1}
	case Left:
		return Cell{-1, 0}
	default:
		return Cell{0, -1}
	}
}

// State is the run state.
type State int

const (
	Playing State = iota
	Won
	Lost
)

// LossReason explains why a run ended in Lost.
type LossReason int

const (
	NoLoss LossReason = iota
	HitWall
	HitSelf
)

const (
	// DefaultWidth and DefaultHeight size the playfield in cells.
	DefaultWidth  = 20
	DefaultHeight = 15
	// DefaultStartLength is the initial body length.
	DefaultStartLength = 3
	// DefaultGoal is the number of pellets needed to win.
	DefaultGoal = 20
)

// Game is a single snake run. NewGame is the only supported constructor.
type Game struct {
	Width, Height int
	Goal          int
	State         State
	Loss          LossReason
	Consumed      int
	// Pellet is the current pellet cell. X or Y is negative when no pellet is
	// available (the board is full); check HasPellet.
	Pellet Cell

	body       []Cell
	occupied   []bool
	dir        Direction
	pending    Direction
	hasPending bool
	paused     bool
	rnd        *rand.Rand
}

// NewGame starts a run on a width x height board. The snake begins centered,
// length DefaultStartLength, headed right. A nil rnd falls back to a fixed seed
// so tests stay deterministic.
func NewGame(width, height int, rnd *rand.Rand) *Game {
	if rnd == nil {
		rnd = rand.New(rand.NewSource(1))
	}
	g := &Game{Width: width, Height: height, Goal: DefaultGoal, rnd: rnd}
	g.reset()
	return g
}

// NewDefaultGame starts a run on the default board.
func NewDefaultGame(rnd *rand.Rand) *Game {
	return NewGame(DefaultWidth, DefaultHeight, rnd)
}

// reset rebuilds the starting position and clears run state.
func (g *Game) reset() {
	if g.occupied == nil || len(g.occupied) != g.Width*g.Height {
		g.occupied = make([]bool, g.Width*g.Height)
	} else {
		for i := range g.occupied {
			g.occupied[i] = false
		}
	}

	length := DefaultStartLength
	if length > g.Width {
		length = g.Width
	}
	headX := g.Width / 2
	if headX < length-1 {
		headX = length - 1
	}
	headY := g.Height / 2

	g.body = make([]Cell, length)
	for i := 0; i < length; i++ {
		c := Cell{X: headX - i, Y: headY}
		g.body[i] = c
		g.occupied[g.index(c)] = true
	}

	g.State = Playing
	g.Loss = NoLoss
	g.Consumed = 0
	g.dir = Right
	g.pending = Right
	g.hasPending = false
	g.paused = false
	g.spawnPellet()
}

// Restart begins a fresh run on the same board.
func (g *Game) Restart() { g.reset() }

// Turn requests a direction change. The change applies at the next tick. A
// reversal of the current direction is rejected, as is any request made while a
// turn is already queued, so at most one turn applies per tick. It reports
// whether the request was accepted.
func (g *Game) Turn(d Direction) bool {
	if g.State != Playing || g.paused {
		return false
	}
	if d == g.dir || d == g.dir.Opposite() || g.hasPending {
		return false
	}
	g.pending = d
	g.hasPending = true
	return true
}

// Tick advances the snake one cell. It does nothing unless the run is playing
// and unpaused.
func (g *Game) Tick() {
	if g.State != Playing || g.paused {
		return
	}

	if g.hasPending {
		if g.pending != g.dir.Opposite() {
			g.dir = g.pending
		}
		g.hasPending = false
	}

	d := g.dir.delta()
	next := Cell{X: g.body[0].X + d.X, Y: g.body[0].Y + d.Y}
	if next.X < 0 || next.X >= g.Width || next.Y < 0 || next.Y >= g.Height {
		g.State = Lost
		g.Loss = HitWall
		return
	}
	if g.occupied[g.index(next)] {
		g.State = Lost
		g.Loss = HitSelf
		return
	}

	grow := next == g.Pellet
	if !grow {
		tail := g.body[len(g.body)-1]
		g.occupied[g.index(tail)] = false
		g.body = g.body[:len(g.body)-1]
	}

	g.occupied[g.index(next)] = true
	g.body = append(g.body, Cell{})
	copy(g.body[1:], g.body[:len(g.body)-1])
	g.body[0] = next

	if !grow {
		return
	}
	g.Consumed++
	if g.Consumed >= g.Goal {
		g.State = Won
		g.Pellet = Cell{X: -1, Y: -1}
		return
	}
	g.spawnPellet()
}

// Pause stops or resumes ticks.
func (g *Game) Pause(on bool) { g.paused = on }

// TogglePause flips the paused state and returns the new value.
func (g *Game) TogglePause() bool {
	g.paused = !g.paused
	return g.paused
}

// Paused reports whether ticks are currently ignored.
func (g *Game) Paused() bool { return g.paused }

// Direction reports the direction the snake will move on the next tick.
func (g *Game) Direction() Direction { return g.dir }

// Head returns the head cell.
func (g *Game) Head() Cell { return g.body[0] }

// Body returns the body cells, head first, tail last. Callers must not mutate
// the returned slice.
func (g *Game) Body() []Cell { return g.body }

// HasPellet reports whether a pellet is present on the board.
func (g *Game) HasPellet() bool { return g.Pellet.X >= 0 && g.Pellet.Y >= 0 }

func (g *Game) index(c Cell) int { return c.Y*g.Width + c.X }

// spawnPellet places the pellet on a uniformly chosen free cell. A full board
// leaves no pellet.
func (g *Game) spawnPellet() {
	free := make([]Cell, 0, len(g.occupied))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			c := Cell{X: x, Y: y}
			if !g.occupied[g.index(c)] {
				free = append(free, c)
			}
		}
	}
	if len(free) == 0 {
		g.Pellet = Cell{X: -1, Y: -1}
		return
	}
	g.Pellet = free[g.rnd.Intn(len(free))]
}
