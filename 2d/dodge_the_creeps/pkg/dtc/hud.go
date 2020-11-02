package dtc

import (
	"strconv"

	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
)

type HUD struct {
	gdnative.CanvasLayerImpl
	gdnative.UserDataIdentifiableImpl

	deathCount int
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

func (p *HUD) ShowMessage(text gdnative.String) {
	p.showMessage(text.AsGoString())
}

// ShowMessage can only be called from Go because of the native string argument
func (p *HUD) showMessage(text string) {
	messageLabel := gdnative.NewLabelWithOwner(p.GetNode(gdnative.NewNodePath("MessageLabel")).GetOwnerObject())
	messageLabel.SetText(text)
	messageLabel.Show()

	messageTimer := gdnative.NewTimerWithOwner(p.GetNode(gdnative.NewNodePath("MessageTimer")).GetOwnerObject())
	messageTimer.Start(-1)
}

func (p *HUD) ShowGameOver() {
	p.showMessage("Game Over")

	// yield($messageTimer, "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	messageTimer := gdnative.NewTimerWithOwner(p.GetNode(gdnative.NewNodePath("MessageTimer")).GetOwnerObject())
	messageTimer.Connect("timeout", p, "show_game_over_yield_message_timer_timeout", binds, int64(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldMessageTimerTimeout() {
	messageLabel := gdnative.NewLabelWithOwner(p.GetNode(gdnative.NewNodePath("MessageLabel")).GetOwnerObject())
	messageLabel.SetText("Dodge the\nCreeps")
	messageLabel.Show()

	// yield(get_tree().create_timer(1), "timeout")
	binds := gdnative.NewArray()
	defer binds.Destroy()
	timer := p.GetTree().CreateTimer(1, true)
	timer.Connect("timeout", p, "show_game_over_yield_scene_tree_timer_timeout", binds, int64(gdnative.OBJECT_CONNECT_ONESHOT))
}

func (p *HUD) ShowGameOverYieldSceneTreeTimerTimeout() {
	log.Debug("ShowGameOverYieldSceneTreeTimerTimeout")
	startButton := gdnative.NewButtonWithOwner(p.GetNode(gdnative.NewNodePath("StartButton")).GetOwnerObject())
	startButton.Show()
}

func (p *HUD) UpdateScore(score int64) {
	scoreLabel := gdnative.NewLabelWithOwner(p.GetNode(gdnative.NewNodePath("ScoreLabel")).GetOwnerObject())
	scoreLabel.SetText(strconv.Itoa(int(score)))
}

func (p *HUD) OnStartButtonPressed() {
	startButton := gdnative.NewButtonWithOwner(p.GetNode(gdnative.NewNodePath("StartButton")).GetOwnerObject())
	startButton.Hide()
	p.EmitSignal("start_game")
}

func (p *HUD) OnMessageTimerTimeout() {
	messageLabel := gdnative.NewLabelWithOwner(p.GetNode(gdnative.NewNodePath("MessageLabel")).GetOwnerObject())
	messageLabel.Hide()
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

func init() {
	gdnative.RegisterInitCallback(func() {
		gdnative.RegisterClass(&HUD{})
	})
}
