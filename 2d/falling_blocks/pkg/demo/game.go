package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"godot-go-demo-projects/2d/falling_blocks/pkg/falling_blocks"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// autoDrop (FALLING_BLOCKS_AUTODROP=1) is a development mode that hard-drops every
// piece on each gravity tick, enabling headless play-through testing.
var (
	autoDrop = os.Getenv("FALLING_BLOCKS_AUTODROP") != ""
	// autoGarbage force-sends one garbage row per lock (wire-path testing).
	autoGarbage = os.Getenv("FALLING_BLOCKS_AUTOGARBAGE") != ""
	// autoRematch makes the host restart after a match ends (headless soak).
	autoRematch = os.Getenv("FALLING_BLOCKS_AUTOREMATCH") != ""
)

const (
	boardX   = float32(20)
	boardY   = float32(20)
	cellSize = float32(32)

	// Next-piece preview mini-grid, drawn below the HUD "Next:" label
	// (which sits at x=380, y=116..140 in Match.tscn).
	previewX    = float32(380)
	previewY    = float32(150)
	previewCell = float32(20)
)

// kindColors indexes fallingblocks.Kind for board rendering.
var kindColors = [7][3]float32{
	{0.2, 0.9, 0.9},   // I cyan
	{0.95, 0.85, 0.2}, // O yellow
	{0.7, 0.3, 0.9},   // T purple
	{0.3, 0.9, 0.3},   // S green
	{0.95, 0.3, 0.3},  // Z red
	{0.3, 0.45, 0.95}, // J blue
	{0.95, 0.6, 0.2},  // L orange
}

// RegisterClassFallingBlocksGame registers the match root class.
func RegisterClassFallingBlocksGame() {
	ClassDBRegisterClass(NewFallingBlocksGameFromOwnerObject, []GDExtensionPropertyInfo{}, nil, func(t *FallingBlocksGame) {
		ClassDBBindMethodVirtual(t, "V_FallingBlocksGame_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_FallingBlocksGame_Process", "_process", nil, nil)
		ClassDBBindMethodVirtual(t, "V_FallingBlocksGame_Draw", "_draw", nil, nil)
		ClassDBBindMethod(t, "OnGravityTimeout", "on_gravity_timeout", nil, nil)
		ClassDBBindMethod(t, "OnQuit", "on_quit", nil, nil)
		ClassDBBindMethod(t, "OnRematch", "on_rematch", nil, nil)
		ClassDBBindMethod(t, "OnPlayAgain", "on_play_again", nil, nil)
	})
}

func NewFallingBlocksGameFromOwnerObject(owner *GodotObject) GDClass {
	obj := &FallingBlocksGame{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// UnregisterClassFallingBlocksGame unregisters the match root class.
func UnregisterClassFallingBlocksGame() {
	ClassDBUnregisterClass[*FallingBlocksGame]()
}

// FallingBlocksGame is the root of Match.tscn: renders the board, drives gravity,
// reads input, and exchanges versus state through the Session.
type FallingBlocksGame struct {
	Node2DImpl

	session *session
	g       *fallingblocks.Game

	gravity   Timer
	scoreLbl  Label
	linesLbl  Label
	levelLbl  Label
	nextLbl   Label
	oppLbl    [3]Label
	message   Label
	panel     Panel
	winnerLbl Label

	rematchBtn Button
	againBtn   Button

	token   int64
	running bool
	dirty   bool

	lastPoll     time.Time
	lastRoster   Roster
	selfID       int32
	isHost       bool
	shifter      fallingblocks.ShiftRepeater
	softRepeat   time.Time
	lastLevel    int
	resultLogged bool
	rematchFired bool
}

func (c *FallingBlocksGame) GetClassName() string { return "FallingBlocksGame" }
func (c *FallingBlocksGame) GetParentClassName() string {
	return "Node2D"
}

func (c *FallingBlocksGame) V_FallingBlocksGame_Ready() {
	c.session = sessionNode(c)
	if c.session == nil {
		return // editor preview
	}
	c.selfID = c.session.selfId()
	c.isHost = c.session.isHostPeer()

	c.gravity = findTimer(c, "Gravity")
	c.scoreLbl = findLabel(c, "HUD/ScoreLabel")
	c.linesLbl = findLabel(c, "HUD/LinesLabel")
	c.levelLbl = findLabel(c, "HUD/LevelLabel")
	c.nextLbl = findLabel(c, "HUD/NextLabel")
	for i := 0; i < 3; i++ {
		c.oppLbl[i] = findLabel(c, fmt.Sprintf("HUD/OppBox/OppLabel%d", i))
	}
	c.message = findLabel(c, "HUD/MessageLabel")
	c.panel = findPanel(c, "HUD/ResultPanel")
	c.winnerLbl = findLabel(c, "HUD/ResultPanel/WinnerLabel")
	c.rematchBtn = findButton(c, "HUD/ResultPanel/RematchButton")
	c.againBtn = findButton(c, "HUD/ResultPanel/PlayAgainButton")

	bindSignal(c.gravity, c, "timeout", "on_gravity_timeout")
	bindPressed(c.rematchBtn, c, "on_rematch")
	bindPressed(c.againBtn, c, "on_play_again")
	bindPressed(findButton(c, "HUD/QuitButton"), c, "on_quit")

	c.startNewGame()
}

func (c *FallingBlocksGame) startNewGame() {
	_ = json.Unmarshal([]byte(c.session.rosterJSON()), &c.lastRoster)
	c.token = c.lastRoster.Token

	c.g = fallingblocks.NewTimedGame()
	c.g.OnLock(c.onLock)
	c.running = true
	c.lastLevel = c.g.Level
	c.gravity.SetWaitTime(c.g.GravityInterval())
	c.gravity.Start(-1)
	c.setPanelVisible(false)
	c.rematchFired = false
	showNode(c.message, false)
	c.dirty = true
	c.refreshStats()
}

func (c *FallingBlocksGame) onLock(cleared, garbageSent int) {
	c.dirty = true
	c.refreshStats()
	if c.lastRoster.Solo || c.lastRoster.Phase != PhaseMatch {
		return
	}
	if cleared > 0 && garbageSent > 0 {
		c.session.reportGarbage(garbageSent)
	}
	if autoGarbage {
		c.session.reportGarbage(1)
	}
	c.session.reportStatus(c.g.Score, c.g.Lines, c.g.Pending)
}

func (c *FallingBlocksGame) V_FallingBlocksGame_Process() {
	if c.g == nil {
		return
	}
	c.pollSession()
	c.handleInput()
	if c.g.Over && c.running {
		c.onLocalGameOver()
	}
	if c.dirty {
		c.dirty = false
		c.QueueRedraw()
	}
}

func (c *FallingBlocksGame) pollSession() {
	if time.Since(c.lastPoll) < 100*time.Millisecond {
		return
	}
	c.lastPoll = time.Now()

	if c.running {
		if n := c.session.consumeGarbage(); n > 0 {
			c.g.AddGarbage(n)
			c.dirty = true
		}
	}

	var r Roster
	if err := json.Unmarshal([]byte(c.session.rosterJSON()), &r); err != nil {
		return
	}
	c.lastRoster = r
	c.renderOpponents(&r)

	if r.Token != c.token {
		c.startNewGame()
		return
	}
	if r.Phase == PhaseOver {
		c.showMatchResult(&r)
		if autoRematch && c.isHost && !c.rematchFired {
			c.rematchFired = true
			c.session.requestRematch()
		}
	}
}

func (c *FallingBlocksGame) renderOpponents(r *Roster) {
	i := 0
	for _, p := range r.Players {
		if p.ID == c.selfID {
			continue
		}
		if i >= 3 {
			break
		}
		status := ""
		if !p.Alive {
			status = " — OUT"
		}
		setLabel(c.oppLbl[i], fmt.Sprintf("%s: %d pts / %d lines / +%d gar%s", p.Name, p.Score, p.Lines, p.Pending, status))
		showNode(c.oppLbl[i], true)
		i++
	}
	for ; i < 3; i++ {
		showNode(c.oppLbl[i], false)
	}
}

func (c *FallingBlocksGame) onLocalGameOver() {
	log.Info("falling-blocks-demo: local game over", zap.Int("score", c.g.Score), zap.Int("lines", c.g.Lines))
	c.running = false
	c.gravity.Stop()
	if !c.lastRoster.Solo && c.lastRoster.Phase == PhaseMatch {
		c.session.reportGameOver()
		showNode(c.message, true)
		setLabel(c.message, "ELIMINATED\nWaiting for other players...")
	} else {
		c.setPanelVisible(true)
		showNode(c.againBtn, c.lastRoster.Solo)
		showNode(c.rematchBtn, false)
		setLabel(c.winnerLbl, fmt.Sprintf("GAME OVER\nscore %d", c.g.Score))
	}
	c.refreshStats()
}

func (c *FallingBlocksGame) showMatchResult(r *Roster) {
	if !c.resultLogged {
		c.resultLogged = true
		log.Info("falling-blocks-demo: match result", zap.Int32("winner", r.Winner), zap.Bool("is_host", c.isHost))
	}
	c.running = false
	c.gravity.Stop()
	c.setPanelVisible(true)
	showNode(c.message, false)
	name := "Nobody"
	for _, p := range r.Players {
		if p.ID == r.Winner {
			name = p.Name
		}
	}
	if r.Winner < 0 {
		setLabel(c.winnerLbl, "DRAW")
	} else {
		setLabel(c.winnerLbl, name+" WINS!")
	}
	showNode(c.rematchBtn, c.isHost)
	showNode(c.againBtn, false)
}

func (c *FallingBlocksGame) setPanelVisible(v bool) {
	showNode(c.panel, v)
}

func (c *FallingBlocksGame) refreshStats() {
	setLabel(c.scoreLbl, fmt.Sprintf("Score: %d", c.g.Score))
	setLabel(c.linesLbl, fmt.Sprintf("Lines: %d", c.g.Lines))
	setLabel(c.levelLbl, fmt.Sprintf("Level: %d", c.g.Level))
	setLabel(c.nextLbl, "Next: "+fallingblocks.KindName(c.g.Preview()))
}

func (c *FallingBlocksGame) handleInput() {
	if !c.running || c.g.Paused {
		if actionJustPressed("pause_game") && c.lastRoster.Solo {
			c.togglePause()
		}
		return
	}
	now := time.Now()

	dir := 0
	if actionPressed("move_left") {
		dir = -1
	} else if actionPressed("move_right") {
		dir = 1
	}
	justPressed := (dir == -1 && actionJustPressed("move_left")) ||
		(dir == 1 && actionJustPressed("move_right"))
	if c.shifter.Step(now, dir, justPressed) {
		if c.g.Move(dir) {
			c.dirty = true
		}
	}

	if actionJustPressed("rotate_cw") {
		if c.g.Rotate(1) {
			c.dirty = true
		}
	}
	if actionJustPressed("rotate_ccw") {
		if c.g.Rotate(-1) {
			c.dirty = true
		}
	}

	if actionPressed("soft_drop") {
		if actionJustPressed("soft_drop") || now.After(c.softRepeat) {
			c.g.SoftDrop()
			c.softRepeat = now.Add(40 * time.Millisecond)
			c.dirty = true
		}
	}
	if actionJustPressed("hard_drop") {
		c.g.HardDrop()
		c.dirty = true
	}
	if c.g.Level != c.lastLevel {
		c.lastLevel = c.g.Level
		c.gravity.SetWaitTime(c.g.GravityInterval())
	}
}

func (c *FallingBlocksGame) togglePause() {
	c.g.Paused = !c.g.Paused
	if c.g.Paused {
		c.gravity.Stop()
		showNode(c.message, true)
		setLabel(c.message, "PAUSED\npress P to resume")
	} else {
		showNode(c.message, false)
		c.gravity.Start(-1)
	}
}

func (c *FallingBlocksGame) OnGravityTimeout() {
	if !c.running || c.g == nil || c.g.Paused || c.g.Over {
		return
	}
	if autoDrop {
		c.g.HardDrop()
		c.dirty = true
		return
	}
	if !c.g.GravityStep() {
		c.g.Lock()
	}
	c.dirty = true
}

func (c *FallingBlocksGame) OnQuit() {
	c.session.leave()
	changeScene(c, "res://Main.tscn")
}

func (c *FallingBlocksGame) OnRematch() {
	c.session.requestRematch()
}

func (c *FallingBlocksGame) OnPlayAgain() {
	c.startNewGame()
}

// drawFilledRect paints a solid axis-aligned rectangle. It uses
// draw_colored_polygon rather than draw_rect because godot-go v0.3.43
// mis-marshals draw_rect's filled/width arguments against Godot 4.8,
// emitting a spurious "width has no effect when filled" warning on every
// call (which floods stdout and stalls the main thread).
func drawFilledRect(c *FallingBlocksGame, x, y, w, h float32, col Color) {
	pts := NewPackedVector2Array()
	defer pts.Destroy()
	pts.Append(NewVector2WithFloat32Float32(x, y))
	pts.Append(NewVector2WithFloat32Float32(x+w, y))
	pts.Append(NewVector2WithFloat32Float32(x+w, y+h))
	pts.Append(NewVector2WithFloat32Float32(x, y+h))
	uvs := NewPackedVector2Array()
	defer uvs.Destroy()
	c.DrawColoredPolygon(pts, col, uvs, RefTexture2D(nil))
}

func (c *FallingBlocksGame) V_FallingBlocksGame_Draw() {
	if c.g == nil {
		return
	}
	b := c.g.Board
	bg := NewColorWithFloat32Float32Float32(0.08, 0.08, 0.1)
	drawFilledRect(c, boardX, boardY, float32(b.W)*cellSize, float32(b.H-b.Buffer)*cellSize, bg)

	for y := b.Buffer; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if b.At(x, y) {
				c.fillRect(x, y)
			}
		}
	}
	c.drawGhost()
	for _, cell := range c.g.CurrentCells() {
		if cell.Y >= b.Buffer {
			c.fillRect(cell.X, cell.Y, kindColors[c.g.Cur][:]...)
		}
	}
	c.drawPreview()
}

// drawGhost renders the landing shadow: the cells the active piece would
// occupy if hard-dropped now, as dim gray blocks beneath the live piece.
func (c *FallingBlocksGame) drawGhost() {
	col := NewColorWithFloat32Float32Float32(0.28, 0.28, 0.32)
	for _, cell := range c.g.GhostCells() {
		if cell.Y >= c.g.Board.Buffer {
			drawFilledRect(c,
				boardX+float32(cell.X)*cellSize+1, boardY+float32(cell.Y-c.g.Board.Buffer)*cellSize+1,
				cellSize-2, cellSize-2, col)
		}
	}
}

// drawPreview renders the next tetromino as colored mini-blocks below the
// HUD "Next:" label so the player can anticipate the upcoming piece.
func (c *FallingBlocksGame) drawPreview() {
	next := c.g.Preview()
	cells := fallingblocks.RotationCells(next, 0)
	if len(cells) == 0 {
		return
	}
	minX, minY := cells[0].X, cells[0].Y
	for _, cell := range cells[1:] {
		if cell.X < minX {
			minX = cell.X
		}
		if cell.Y < minY {
			minY = cell.Y
		}
	}
	col := NewColorWithFloat32Float32Float32(
		kindColors[next][0], kindColors[next][1], kindColors[next][2])
	for _, cell := range cells {
		drawFilledRect(c,
			previewX+float32(cell.X-minX)*previewCell+1,
			previewY+float32(cell.Y-minY)*previewCell+1,
			previewCell-2, previewCell-2, col)
	}
}

func (c *FallingBlocksGame) fillRect(x, y int, rgb ...float32) {
	col := NewColorWithFloat32Float32Float32(0.55, 0.57, 0.62)
	if len(rgb) >= 3 {
		col = NewColorWithFloat32Float32Float32(rgb[0], rgb[1], rgb[2])
	}
	drawFilledRect(c,
		boardX+float32(x)*cellSize+1, boardY+float32(y-c.g.Board.Buffer)*cellSize+1,
		cellSize-2, cellSize-2, col)
}
