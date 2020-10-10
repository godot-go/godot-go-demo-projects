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
	e.RegisterMethod("show_message", "ShowMessage_")
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
	p.messageLabel = gdnative.NewLabelWithOwner(p.FindNode("MessageLabel", true, true).GetOwnerObject())
	p.messageTimer = gdnative.NewTimerWithOwner(p.FindNode("MessageTimer", true, true).GetOwnerObject())
	p.startButton = gdnative.NewButtonWithOwner(p.FindNode("StartButton", true, true).GetOwnerObject())
	p.scoreLabel = gdnative.NewLabelWithOwner(p.FindNode("ScoreLabel", true, true).GetOwnerObject())
}

func (p *HUD) ShowMessage_(text gdnative.String) {
	p.ShowMessage(text.AsGoString())
}

// ShowMessage can only be called from Go because of the native string argument
func (p *HUD) ShowMessage(text string) {
	p.messageLabel.SetText(text)
	p.messageLabel.Show()
	p.messageTimer.Start(-1)
}

func (p *HUD) ShowGameOver() {
	p.ShowMessage("Game Over")

	// yield($messageTimer, "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	p.messageTimer.Connect("timeout", p, "show_game_over_yield_message_timer_timeout", binds, int64(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldMessageTimerTimeout() {
	p.messageLabel.SetText("Dodge the\nCreeps")
	p.messageLabel.Show()

	// yield(get_tree().create_timer(1), "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	p.timer = p.GetTree().CreateTimer(1, true)
	p.timer.Connect("timeout", p, "show_game_over_yield_scene_tree_timer_timeout", binds, int64(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldSceneTreeTimerTimeout() {
	p.startButton.Show()
}

func (p *HUD) UpdateScore(score int64) {
	p.scoreLabel.SetText(strconv.Itoa(int(score)))
}

func (p *HUD) OnStartButtonPressed() {
	p.startButton.Hide()
	p.EmitSignal("start_game")
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

func NewHUDWithOwner(owner *gdnative.GodotObject) HUD {
	inst := gdnative.GetCustomClassInstanceWithOwner(owner).(*HUD)
	return *inst
}
