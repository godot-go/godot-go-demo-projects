package snake

import (
	"math/rand"
	"testing"
)

func newGame(w, h int) *Game {
	return NewGame(w, h, rand.New(rand.NewSource(1)))
}

func placePelletAhead(g *Game) {
	d := g.dir.delta()
	head := g.Head()
	g.Pellet = Cell{X: head.X + d.X, Y: head.Y + d.Y}
}

func runUntilStopped(g *Game, limit int) {
	for i := 0; i < limit && g.State == Playing; i++ {
		g.Tick()
	}
}

func isBody(g *Game, c Cell) bool {
	for _, b := range g.Body() {
		if b == c {
			return true
		}
	}
	return false
}

func TestTickAdvancesOneCell(t *testing.T) {
	g := newGame(10, 10)
	g.Pellet = Cell{X: 0, Y: 0} // keep the pellet off the path so no growth occurs
	head := g.Head()
	tail := g.Body()[len(g.Body())-1]

	g.Tick()

	if got, want := g.Head(), (Cell{X: head.X + 1, Y: head.Y}); got != want {
		t.Fatalf("head = %v, want %v", got, want)
	}
	if got := len(g.Body()); got != DefaultStartLength {
		t.Fatalf("body length = %d, want %d", got, DefaultStartLength)
	}
	if isBody(g, tail) {
		t.Fatalf("tail %v should have vacated", tail)
	}
}

func TestNoMovementWithoutTick(t *testing.T) {
	g := newGame(10, 10)
	before := append([]Cell(nil), g.Body()...)

	g.Turn(Down)

	if g.Head() != before[0] {
		t.Fatalf("head moved before a tick: %v, want %v", g.Head(), before[0])
	}
	if g.Direction() != Right {
		t.Fatalf("direction changed before a tick: %v", g.Direction())
	}
}

func TestQueuedTurnAppliesNextTick(t *testing.T) {
	g := newGame(10, 10)
	head := g.Head()

	if !g.Turn(Down) {
		t.Fatal("Turn(Down) was rejected")
	}
	if g.Direction() != Right {
		t.Fatalf("direction = %v before tick, want Right", g.Direction())
	}

	g.Tick()

	if g.Direction() != Down {
		t.Fatalf("direction after tick = %v, want Down", g.Direction())
	}
	if got, want := g.Head(), (Cell{X: head.X, Y: head.Y + 1}); got != want {
		t.Fatalf("head = %v, want %v", got, want)
	}
}

func TestReversalIgnored(t *testing.T) {
	g := newGame(10, 10)

	if g.Turn(Left) {
		t.Fatal("Turn(Left) from Right was accepted")
	}
	g.Tick()

	if g.Direction() != Right {
		t.Fatalf("direction = %v, want Right", g.Direction())
	}
	if got, want := g.Head(), (Cell{X: 6, Y: 5}); got != want {
		t.Fatalf("head = %v, want %v", got, want)
	}
}

func TestOnlyOneTurnPerTick(t *testing.T) {
	g := newGame(10, 10)

	if !g.Turn(Down) {
		t.Fatal("first Turn(Down) was rejected")
	}
	if g.Turn(Up) {
		t.Fatal("second turn in the same tick was accepted")
	}

	g.Tick()

	if g.Direction() != Down {
		t.Fatalf("direction = %v, want Down", g.Direction())
	}
}

func TestGrowthOnConsumption(t *testing.T) {
	g := newGame(10, 10)
	placePelletAhead(g)
	tail := g.Body()[len(g.Body())-1]

	g.Tick()

	if got := len(g.Body()); got != DefaultStartLength+1 {
		t.Fatalf("body length = %d, want %d", got, DefaultStartLength+1)
	}
	if g.Consumed != 1 {
		t.Fatalf("consumed = %d, want 1", g.Consumed)
	}
	if !isBody(g, tail) {
		t.Fatalf("tail %v should be retained after growth", tail)
	}
}

func TestPelletCounterAdvancesAndRespawns(t *testing.T) {
	g := newGame(10, 10)
	eaten := Cell{X: g.Head().X + 1, Y: g.Head().Y}
	g.Pellet = eaten

	g.Tick()

	if g.Consumed != 1 {
		t.Fatalf("consumed = %d, want 1", g.Consumed)
	}
	if !g.HasPellet() {
		t.Fatal("no pellet respawned")
	}
	if g.Pellet == eaten {
		t.Fatal("pellet respawned on the cell just consumed")
	}
	if isBody(g, g.Pellet) {
		t.Fatalf("pellet %v overlaps the body", g.Pellet)
	}
}

func TestSpawnAvoidsBody(t *testing.T) {
	g := newGame(200, 3)
	g.Goal = 1000

	for i := 0; i < 50; i++ {
		placePelletAhead(g)
		g.Tick()
		if g.State != Playing {
			t.Fatalf("run ended early at iteration %d: state=%v", i, g.State)
		}
		if !g.HasPellet() {
			t.Fatalf("no pellet at iteration %d", i)
		}
		if g.Pellet.X < 0 || g.Pellet.X >= g.Width || g.Pellet.Y < 0 || g.Pellet.Y >= g.Height {
			t.Fatalf("pellet %v outside the board", g.Pellet)
		}
		if isBody(g, g.Pellet) {
			t.Fatalf("pellet %v overlaps the body at iteration %d", g.Pellet, i)
		}
	}
}

func TestWallContactLoses(t *testing.T) {
	cases := []struct {
		name  string
		setup func(g *Game)
	}{
		{"right", func(g *Game) {}},
		{"left", func(g *Game) { g.Turn(Left) }},
		{"up", func(g *Game) { g.Turn(Up) }},
		{"down", func(g *Game) { g.Turn(Down) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newGame(6, 6)
			tc.setup(g)
			runUntilStopped(g, 20)
			if g.State != Lost {
				t.Fatalf("state = %v, want Lost", g.State)
			}
			if g.Loss != HitWall {
				t.Fatalf("loss = %v, want HitWall", g.Loss)
			}
		})
	}
}

func TestSelfContactLoses(t *testing.T) {
	g := newGame(8, 8)

	// Grow to length 4 first.
	placePelletAhead(g)
	g.Tick()
	g.Pellet = Cell{X: 0, Y: 7} // keep later moves from growing

	// Loop down, left, up. The last step enters the tail's current cell, which
	// vacates on the same tick: strict collision still loses.
	g.Turn(Down)
	g.Tick()
	g.Turn(Left)
	g.Tick()
	g.Turn(Up)
	g.Tick()

	if g.State != Lost {
		t.Fatalf("state = %v, want Lost", g.State)
	}
	if g.Loss != HitSelf {
		t.Fatalf("loss = %v, want HitSelf", g.Loss)
	}
}

func TestWinAtPelletTargetHaltsMovement(t *testing.T) {
	g := newGame(200, 3)

	for i := 0; i < DefaultGoal+5 && g.State == Playing; i++ {
		placePelletAhead(g)
		g.Tick()
	}

	if g.State != Won {
		t.Fatalf("state = %v, want Won", g.State)
	}
	if g.Consumed != DefaultGoal {
		t.Fatalf("consumed = %d, want %d", g.Consumed, DefaultGoal)
	}
	if g.HasPellet() {
		t.Fatalf("pellet %v should be cleared after winning", g.Pellet)
	}

	head := g.Head()
	g.Tick()
	if g.Head() != head {
		t.Fatalf("head moved after winning: %v -> %v", head, g.Head())
	}
}

func TestPauseStopsTicks(t *testing.T) {
	g := newGame(10, 10)
	head := g.Head()

	g.Pause(true)
	g.Tick()

	if g.Head() != head {
		t.Fatalf("head moved while paused: %v -> %v", head, g.Head())
	}

	g.Pause(false)
	g.Tick()

	if g.Head() == head {
		t.Fatal("head did not move after resuming")
	}
}

func TestRestartResetsRun(t *testing.T) {
	g := newGame(10, 10)
	placePelletAhead(g)
	g.Tick()
	if g.Consumed != 1 {
		t.Fatalf("setup: consumed = %d, want 1", g.Consumed)
	}
	runUntilStopped(g, 20)
	if g.State != Lost {
		t.Fatalf("setup: state = %v, want Lost", g.State)
	}

	g.Restart()

	if g.State != Playing || g.Consumed != 0 {
		t.Fatalf("state=%v consumed=%d, want Playing/0", g.State, g.Consumed)
	}
	if got := len(g.Body()); got != DefaultStartLength {
		t.Fatalf("body length = %d, want %d", got, DefaultStartLength)
	}
	if g.Direction() != Right {
		t.Fatalf("direction = %v, want Right", g.Direction())
	}
	if got, want := g.Head(), (Cell{X: 5, Y: 5}); got != want {
		t.Fatalf("head = %v, want %v", got, want)
	}
	if !g.HasPellet() {
		t.Fatal("restart left no pellet")
	}
}
