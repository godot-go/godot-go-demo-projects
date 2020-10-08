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
	e.RegisterMethod("show_game_over_yield_message_timer_timeout", "ShowGameOverYieldMessageTimerTimeout")
	e.RegisterMethod("show_game_over_yield_scene_tree_timer_timeout", "ShowGameOverYieldSceneTreeTimerTimeout")
	e.RegisterMethod("update_score", "UpdateScore")
	e.RegisterMethod("_on_StartButton_pressed", "OnStartButtonPressed")
	e.RegisterMethod("_on_MessageTimer_timeout", "OnMessageTimerTimeout")

	// signals
	e.RegisterSignal("start_game")
}

func (p *HUD) Ready() {
	strMessageLabel := gdnative.NewStringFromGoString("MessageLabel")
	defer strMessageLabel.Destroy()
	strMessageTimer := gdnative.NewStringFromGoString("MessageTimer")
	defer strMessageTimer.Destroy()
	strStartButton := gdnative.NewStringFromGoString("StartButton")
	defer strStartButton.Destroy()
	strScoreLabel := gdnative.NewStringFromGoString("ScoreLabel")
	defer strScoreLabel.Destroy()

	p.messageLabel = gdnative.NewLabelWithOwner(p.FindNode(strMessageLabel, true, true).GetOwnerObject())
	p.messageTimer = gdnative.NewTimerWithOwner(p.FindNode(strMessageTimer, true, true).GetOwnerObject())
	p.startButton = gdnative.NewButtonWithOwner(p.FindNode(strStartButton, true, true).GetOwnerObject())
	p.scoreLabel = gdnative.NewLabelWithOwner(p.FindNode(strScoreLabel, true, true).GetOwnerObject())
}

func (p *HUD) ShowMessage(text gdnative.String) {
	p.messageLabel.SetText(text)
	p.messageLabel.Show()
	p.messageTimer.Start(-1)
}

func (p *HUD) ShowGameOver() {
	strGameOver := gdnative.NewStringFromGoString("Game Over")
	defer strGameOver.Destroy()
	p.ShowMessage(strGameOver)

	// yield($messageTimer, "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	method := gdnative.NewStringFromGoString("show_game_over_yield_message_timer_timeout")
	defer method.Destroy()
	strTimeout := gdnative.NewStringFromGoString("timeout")
	defer strTimeout.Destroy()
	p.messageTimer.Connect(strTimeout, p, method, binds, int64(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldMessageTimerTimeout() {
	text := gdnative.NewStringFromGoString("Dodge the\nCreeps")
	defer text.Destroy()
	p.messageLabel.SetText(text)
	p.messageLabel.Show()

	// yield(get_tree().create_timer(1), "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	method := gdnative.NewStringFromGoString("show_game_over_yield_scene_tree_timer_timeout")
	defer method.Destroy()
	strTimeout := gdnative.NewStringFromGoString("timeout")
	strTimeout.Destroy()
	p.timer = p.GetTree().CreateTimer(1, true)
	p.timer.Connect(strTimeout, p, method, binds, int64(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldSceneTreeTimerTimeout() {
	p.startButton.Show()
}

func (p *HUD) UpdateScore(score int64) {
	text := gdnative.NewStringFromGoString(strconv.Itoa(int(score)))
	defer text.Destroy()
	p.scoreLabel.SetText(text)
}

func (p *HUD) OnStartButtonPressed() {
	p.startButton.Hide()
	startGame := gdnative.NewStringFromGoString("start_game")
	defer startGame.Destroy()
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
