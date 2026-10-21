package demo

import (
	"fmt"
	"math/rand"
	"time"

	"godot-go-demo-projects/2d/snake/pkg/snake"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

const (
	// The playfield sits below an 80 px HUD strip and is exactly 640 px wide.
	boardX   = float32(0)
	boardY   = float32(80)
	cellSize = float32(32)

	baseSpeed    = float32(6)  // cells per second at the start
	maxSpeed     = float32(12) // cells per second ceiling
	speedStep    = 5           // pellets between speed increases
	minPelletGap = float32(1)  // inset so adjacent cells stay visually separate

	startPrompt = "PRESS ANY KEY TO START"
)

// RegisterClassSnakeGame registers the demo's root node class.
func RegisterClassSnakeGame() {
	ClassDBRegisterClass(NewSnakeGameFromOwnerObject, []PropertyDescriptor{}, nil, func(t *SnakeGame) {
		ClassDBBindMethodVirtual(t, "V_SnakeGame_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_SnakeGame_Process", "_process", nil, nil)
		ClassDBBindMethodVirtual(t, "V_SnakeGame_Draw", "_draw", nil, nil)
		ClassDBBindMethod(t, "OnStepTimeout", "on_step_timeout", nil, nil)
	})
}

func NewSnakeGameFromOwnerObject(owner *GodotObject) GDClass {
	obj := &SnakeGame{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// UnregisterClassSnakeGame unregisters the root node class.
func UnregisterClassSnakeGame() {
	ClassDBUnregisterClass[*SnakeGame]()
}

// SnakeGame is the root of Main.tscn: it drives the rules clock and draws the
// board. All rules live in package snake.
type SnakeGame struct {
	Node2DImpl

	g              *snake.Game
	step           Timer
	progressLbl    Label
	messageLbl     Label
	startLbl       Label
	paused         bool
	started        bool
	waitForRelease bool
	lastConsumed   int
	dirty          bool
}

func (c *SnakeGame) GetClassName() string { return "SnakeGame" }
func (c *SnakeGame) GetParentClassName() string {
	return "Node2D"
}

func (c *SnakeGame) V_SnakeGame_Ready() {
	c.step = findTimer(c, "Step")
	c.progressLbl = findLabel(c, "HUD/ProgressLabel")
	c.messageLbl = findLabel(c, "HUD/MessageLabel")
	c.startLbl = findLabel(c, "HUD/StartLabel")
	bindSignal(c.step, c, "timeout", "on_step_timeout")
	setLabel(c.startLbl, startPrompt)

	c.g = snake.NewDefaultGame(rand.New(rand.NewSource(time.Now().UnixNano())))
	c.lastConsumed = 0
	c.applySpeed()
	c.refreshHUD()
	c.enterWaiting()
	c.dirty = true
}

func (c *SnakeGame) V_SnakeGame_Process() {
	if c.g == nil {
		return
	}
	if c.started {
		c.handleInput()
	} else {
		c.pollStart()
	}
	if c.dirty {
		c.dirty = false
		c.QueueRedraw()
	}
}

// enterWaiting parks the run on the board until the player presses something.
// waitForRelease keeps the keystroke that triggered a restart from immediately
// starting the new run.
func (c *SnakeGame) enterWaiting() {
	c.started = false
	c.waitForRelease = true
	c.step.Stop()
	c.showMessage("")
	showNode(c.startLbl, true)
}

// pollStart watches for the first key or mouse button, which starts the timer.
// A direction press also becomes the run's first turn.
func (c *SnakeGame) pollStart() {
	input := GetInputSingleton()
	if c.waitForRelease {
		if !input.IsAnythingPressed() {
			c.waitForRelease = false
		}
		return
	}
	if !input.IsAnythingPressed() {
		return
	}

	c.started = true
	c.showMessage("")
	showNode(c.startLbl, false)
	c.step.Start(-1)
	switch {
	case actionJustPressed("move_up"):
		c.g.Turn(snake.Up)
	case actionJustPressed("move_down"):
		c.g.Turn(snake.Down)
	case actionJustPressed("move_left"):
		c.g.Turn(snake.Left)
	case actionJustPressed("move_right"):
		c.g.Turn(snake.Right)
	}
	c.dirty = true
}

func (c *SnakeGame) handleInput() {
	if c.g.State != snake.Playing {
		if actionJustPressed("restart") {
			c.restart()
		}
		return
	}
	if actionJustPressed("pause") {
		c.togglePause()
	}
	if c.paused {
		return
	}
	switch {
	case actionJustPressed("move_up"):
		c.g.Turn(snake.Up)
	case actionJustPressed("move_down"):
		c.g.Turn(snake.Down)
	case actionJustPressed("move_left"):
		c.g.Turn(snake.Left)
	case actionJustPressed("move_right"):
		c.g.Turn(snake.Right)
	}
}

func (c *SnakeGame) togglePause() {
	c.paused = c.g.TogglePause()
	if c.paused {
		c.step.Stop()
		c.showMessage("PAUSED\npress P to resume")
	} else {
		c.step.Start(-1)
		c.showMessage("")
	}
	c.dirty = true
}

func (c *SnakeGame) restart() {
	c.g.Restart()
	c.paused = false
	c.lastConsumed = 0
	c.applySpeed()
	c.refreshHUD()
	c.enterWaiting()
	c.dirty = true
}

// OnStepTimeout is the Step timer's timeout handler: one rules tick.
func (c *SnakeGame) OnStepTimeout() {
	if c.g == nil || !c.started || c.paused || c.g.State != snake.Playing {
		return
	}
	c.g.Tick()

	if c.g.Consumed != c.lastConsumed {
		c.lastConsumed = c.g.Consumed
		c.applySpeed()
	}

	switch c.g.State {
	case snake.Won:
		c.step.Stop()
		c.showMessage(fmt.Sprintf("YOU WIN\n%d pellets eaten — press Enter to play again", c.g.Goal))
	case snake.Lost:
		c.step.Stop()
		if c.g.Loss == snake.HitWall {
			c.showMessage("GAME OVER\nhit the wall — press Enter to try again")
		} else {
			c.showMessage("GAME OVER\nbit your own body — press Enter to try again")
		}
	}

	c.refreshHUD()
	c.dirty = true
}

// applySpeed converts the current pellet count into the tick period. The rules
// package never sees speed; it only sees ticks.
func (c *SnakeGame) applySpeed() {
	speed := baseSpeed + float32(c.g.Consumed/speedStep)
	if speed > maxSpeed {
		speed = maxSpeed
	}
	c.step.SetWaitTime(float64(1 / speed))
}

func (c *SnakeGame) refreshHUD() {
	setLabel(c.progressLbl, fmt.Sprintf("Pellets: %d / %d", c.g.Consumed, c.g.Goal))
}

func (c *SnakeGame) showMessage(text string) {
	setLabel(c.messageLbl, text)
	showNode(c.messageLbl, text != "")
}

func (c *SnakeGame) V_SnakeGame_Draw() {
	if c.g == nil {
		return
	}
	w := float32(c.g.Width) * cellSize
	h := float32(c.g.Height) * cellSize

	bg := NewColorWithFloat32Float32Float32(0.05, 0.07, 0.09)
	drawFilledRect(c, boardX, boardY, w, h, bg)

	// Checker pattern makes the grid readable without drawing lines.
	light := NewColorWithFloat32Float32Float32(0.08, 0.11, 0.14)
	for y := 0; y < c.g.Height; y++ {
		for x := 0; x < c.g.Width; x++ {
			if (x+y)%2 != 0 {
				continue
			}
			drawFilledRect(c,
				boardX+float32(x)*cellSize, boardY+float32(y)*cellSize,
				cellSize, cellSize, light)
		}
	}

	if c.g.HasPellet() {
		pellet := NewColorWithFloat32Float32Float32(0.95, 0.35, 0.35)
		pad := float32(8)
		drawFilledRect(c,
			boardX+float32(c.g.Pellet.X)*cellSize+pad,
			boardY+float32(c.g.Pellet.Y)*cellSize+pad,
			cellSize-2*pad, cellSize-2*pad, pellet)
	}

	for i, cell := range c.g.Body() {
		col := NewColorWithFloat32Float32Float32(0.25, 0.78, 0.42)
		if i == 0 {
			col = NewColorWithFloat32Float32Float32(0.55, 0.98, 0.65)
		}
		drawFilledRect(c,
			boardX+float32(cell.X)*cellSize+minPelletGap,
			boardY+float32(cell.Y)*cellSize+minPelletGap,
			cellSize-2*minPelletGap, cellSize-2*minPelletGap, col)
	}
}
