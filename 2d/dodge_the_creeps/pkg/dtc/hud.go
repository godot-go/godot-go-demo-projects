package dtc

import (
	"strconv"

	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
)

type HUD struct {
	gdnative.CanvasLayerImpl
	gdnative.UserDataIdentifiableImpl

	messageLabel gdnative.Label
	messageTimer gdnative.Timer
	startButton gdnative.Button
	timer gdnative.SceneTreeTimer
	scoreLabel gdnative.Label
}

func (p *HUD) ClassName() string {
	return "HUD"
}

func (p *HUD) BaseClass() string {
	return "CanvasLayer"
}

func (p *HUD) Init() {
}

func (p *HUD) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	// methods
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("show_message", "ShowMessage")
	e.RegisterMethod("show_game_over", "ShowGameOver")
	e.RegisterMethod("update_score", "UpdateScore")

	// signals
	e.RegisterSignal("start_game")
}

func (p *HUD) Ready() {
	p.messageLabel = gdnative.NewLabelWithOwner(p.FindNode(messageLabel, true, true).GetOwnerObject())
	p.messageTimer = gdnative.NewTimerWithOwner(p.FindNode(messageTimer, true, true).GetOwnerObject())
	p.startButton = gdnative.NewButtonWithOwner(p.FindNode(startButton, true, true).GetOwnerObject())
	p.scoreLabel = gdnative.NewLabelWithOwner(p.FindNode(scoreLabel, true, true).GetOwnerObject())

	binds := gdnative.NewArray()

	gstrName := p.GetName()

	log.Info("ready info", gdnative.StringField("name", gstrName.AsGoString()))

	onStartButtonPressed := gdnative.NewStringFromGoString("OnStartButtonPressed")
	defer onStartButtonPressed.Destroy()
	p.startButton.Connect(pressed, p, onStartButtonPressed, binds, 0)

	onMessageTimerTimeout := gdnative.NewStringFromGoString("OnMessageTimerTimeout")
	defer onMessageTimerTimeout.Destroy()
	p.messageTimer.Connect(timeout, p, onMessageTimerTimeout, binds, 0)
}

func (p *HUD) ShowMessage(text gdnative.String) {
	p.messageLabel.SetText(text)
	p.messageLabel.Show()
	p.messageTimer.Start(-1)
}

func (p *HUD) ShowGameOver() {
	p.ShowMessage(gameOver)

	// yield($messageTimer, "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	p.messageTimer.Connect(timeout, p, showGameOverYieldMessageTimerTimeoutMethod, binds, int32(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldMessageTimerTimeout() {
	text := gdnative.NewStringFromGoString("Dodge the\nCreeps")
	defer text.Destroy()
	p.messageLabel.SetText(text)
	p.messageLabel.Show()

	// yield(get_tree().create_timer(1), "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	p.timer = p.GetTree().CreateTimer(1, true)
	p.timer.Connect(timeout, p, showGameOverYieldSceneTreeTimerTimeout, binds, int32(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldSceneTreeTimerTimeout() {
	p.startButton.Show()
}

func (p *HUD) UpdateScore(score int32) {
	text := gdnative.NewStringFromGoString(strconv.Itoa(int(score)))
	defer text.Destroy()
	p.scoreLabel.SetText(text)
}

func (p *HUD) OnStartButtonPressed() {
	p.startButton.Hide()
	p.EmitSignal(startGame)
}

func (p *HUD) OnMessageTimerTimeout() {
	p.messageLabel.Hide()
}

func (p *HUD) Free() {
}

func NewHUD() HUD {
	log.Debug("NewHUD")

	inst := gdnative.CreateCustomClassInstance("HUD", "CanvasLayer").(*HUD)
	return *inst
}

var (
	messageLabel gdnative.String
	messageTimer gdnative.String
	startButton gdnative.String
	gameOver gdnative.String
	pressed gdnative.String
	timeout gdnative.String
	showGameOverYieldMessageTimerTimeoutMethod gdnative.String
	showGameOverYieldSceneTreeTimerTimeout gdnative.String
	scoreLabel gdnative.String
	startGame gdnative.String
)

func HUDNativescriptInit() {
	messageLabel = gdnative.NewStringFromGoString("MessageLabel")
	messageTimer = gdnative.NewStringFromGoString("MessageTimer")
	startButton = gdnative.NewStringFromGoString("StartButton")
	gameOver = gdnative.NewStringFromGoString("Game Over")
	pressed = gdnative.NewStringFromGoString("pressed")
	timeout = gdnative.NewStringFromGoString("timeout")
	showGameOverYieldMessageTimerTimeoutMethod = gdnative.NewStringFromGoString("ShowGameOverYieldMessageTimerTimeout")
	showGameOverYieldSceneTreeTimerTimeout = gdnative.NewStringFromGoString("ShowGameOverYieldSceneTreeTimerTimeout")
	scoreLabel = gdnative.NewStringFromGoString("ScoreLabel")
	startGame = gdnative.NewStringFromGoString("start_game")
}

func HUDNativescriptTerminate() {
	messageLabel.Destroy()
	messageTimer.Destroy()
	startButton.Destroy()
	gameOver.Destroy()
	pressed.Destroy()
	timeout.Destroy()
	showGameOverYieldMessageTimerTimeoutMethod.Destroy()
	showGameOverYieldSceneTreeTimerTimeout.Destroy()
	scoreLabel.Destroy()
	startGame.Destroy()
}
