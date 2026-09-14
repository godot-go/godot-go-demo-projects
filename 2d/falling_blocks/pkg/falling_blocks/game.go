package fallingblocks

import (
	"math"
	"math/rand"
	"time"
)

// Game ties a Board, a bag Randomizer, and the scoring/leveling/garbage rules
// together. It has no dependency on Godot so it can be unit tested directly.
type Game struct {
	Board *Board
	rnd   *rand.Rand
	bag   *Randomizer

	Cur    Kind
	CurX   int
	CurY   int
	CurRot int

	Score    int
	Lines    int
	Level    int
	Combo    int
	Pending  int // queued garbage rows, applied at next lock
	Over     bool
	Paused   bool
	GraceSec float64 // lock-delay grace accumulated (presentation may use)

	onLockHooks []func(cleared, garbageSent int)
}

// NewGame creates a running game seeded with the given value.
func NewGame(seed int64) *Game {
	rnd := rand.New(rand.NewSource(seed))
	b := NewBoard(10, 22, 2, rnd)
	g := &Game{Board: b, rnd: rnd, bag: NewRandomizer(rnd), Level: 1}
	g.spawn()
	return g
}

// NewTimedGame seeds from wall clock (for real play).
func NewTimedGame() *Game {
	return NewGame(time.Now().UnixNano())
}

func (g *Game) spawn() {
	g.Cur = g.bag.Next()
	x, y, ok := g.Board.Spawn(g.Cur)
	g.CurX, g.CurY, g.CurRot = x, y, 0
	if !ok {
		g.Over = true
	}
}

// Preview returns the next-next piece (the piece after the current spawn).
func (g *Game) Preview() Kind { return g.bag.Preview() }

// CurrentCells returns absolute cells of the active piece.
func (g *Game) CurrentCells() []Cell {
	return PieceCells(g.Cur, g.CurRot, g.CurX, g.CurY)
}

func (g *Game) canPlace(dx, dy, drot int) bool {
	rot := (g.CurRot + drot + 4) % 4
	return g.Board.Fits(PieceCells(g.Cur, rot, g.CurX+dx, g.CurY+dy))
}

// Move shifts the active piece horizontally; reports success.
func (g *Game) Move(dx int) bool {
	if g.Over || g.Paused || !g.canPlace(dx, 0, 0) {
		return false
	}
	g.CurX += dx
	return true
}

// Rotate turns the active piece (dir +1 cw, -1 ccw) applying horizontal
// wall-kick offsets of up to two cells; reports success.
func (g *Game) Rotate(dir int) bool {
	if g.Over || g.Paused {
		return false
	}
	kicks := []int{0, 1, -1}
	if g.Cur == KindI {
		kicks = []int{0, 1, -1, 2, -2}
	}
	for _, k := range kicks {
		if g.canPlace(k, 0, dir) {
			g.CurX += k
			g.CurRot = (g.CurRot + dir + 4) % 4
			return true
		}
	}
	return false
}

// GravityStep moves the piece down one cell if possible. It reports whether
// the piece fell (false means it is resting and should lock).
func (g *Game) GravityStep() bool {
	if g.Over || g.Paused {
		return false
	}
	if g.canPlace(0, 1, 0) {
		g.CurY++
		return true
	}
	return false
}

// SoftDrop moves the piece down one cell and scores 1 when it entered.
func (g *Game) SoftDrop() {
	if g.GravityStep() {
		g.Score++
	}
}

// HardDrop teleports the piece to the lowest valid row, scores 2 per cell,
// and locks it.
func (g *Game) HardDrop() {
	if g.Over || g.Paused {
		return
	}
	dist := 0
	for g.canPlace(0, dist+1, 0) {
		dist++
	}
	g.CurY += dist
	g.Score += 2 * dist
	g.Lock()
}

// GhostCells returns the absolute cells the active piece would occupy if
// hard-dropped now: the landing position without moving or locking.
func (g *Game) GhostCells() []Cell {
	dist := 0
	for g.canPlace(0, dist+1, 0) {
		dist++
	}
	return PieceCells(g.Cur, g.CurRot, g.CurX, g.CurY+dist)
}

// LockResult describes the outcome of locking the active piece.
type LockResult struct {
	Cleared     int
	GarbageSent int
	Over        bool
}

// LockPieceAtBottom locks the piece if it is resting; used by gravity lock.
func (g *Game) LockPieceAtBottom() LockResult {
	if g.canPlace(0, 1, 0) {
		return LockResult{}
	}
	return g.Lock()
}

// Lock stamps the active piece, injects pending garbage, clears lines, scores,
// and spawns the next piece.
func (g *Game) Lock() LockResult {
	if g.Over {
		return LockResult{Over: true}
	}
	g.Board.Lock(g.Cur, g.CurRot, g.CurX, g.CurY)

	if g.Pending > 0 {
		g.Board.AddGarbageLines(g.Pending)
		g.Pending = 0
	}

	cleared := g.Board.ClearLines()

	res := LockResult{Cleared: cleared}
	if cleared > 0 {
		base := [5]int{0, 100, 300, 500, 800}
		if cleared > 4 {
			cleared = 4
		}
		res.Cleared = cleared
		attack := [5]int{0, 0, 1, 2, 4}[cleared]
		res.GarbageSent = attack + g.Combo
		g.Combo++
		g.Score += base[cleared] * g.Level
		g.Lines += cleared
		g.Level = 1 + g.Lines/10
	} else {
		g.Combo = 0
	}

	if g.Board.LockedInBuffer() {
		g.Over = true
		res.Over = true
	} else {
		g.spawn()
		res.Over = g.Over
	}
	for _, h := range g.onLockHooks {
		h(res.Cleared, res.GarbageSent)
	}
	return res
}

// AddGarbage queues garbage rows to be injected at the next lock.
func (g *Game) AddGarbage(n int) {
	if n < 0 {
		n = 0
	}
	g.Pending += n
}

// OnLock registers a hook invoked after every successful lock with the number
// of cleared lines and garbage sent.
func (g *Game) OnLock(fn func(cleared, garbageSent int)) {
	g.onLockHooks = append(g.onLockHooks, fn)
}

// GravityInterval returns the seconds per gravity step at the given level
// (classic-flavored curve, clamped).
func GravityInterval(level int) float64 {
	if level < 1 {
		level = 1
	}
	v := 1.0 * math.Pow(0.8, float64(level-1))
	if v < 0.05 {
		v = 0.05
	}
	return v
}

// GravityInterval is the active game's current interval.
func (g *Game) GravityInterval() float64 { return GravityInterval(g.Level) }
