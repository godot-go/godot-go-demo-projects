package demo

import (
	. "github.com/godot-go/godot-go/pkg/gdextension"
	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func RegisterClassHUD() {
	ClassDBRegisterClass(&HUD{}, []GDExtensionPropertyInfo{}, nil, func(t GDClass) {
		// virtuals
		ClassDBBindMethodVirtual(t, "V_OnStartButtonPressed", "_on_StartButton_pressed", nil, nil)
		ClassDBBindMethodVirtual(t, "V_OnMessageTimerTimeout", "_on_MessageTimer_timeout", nil, nil)

		// properties
		ClassDBBindMethod(t, "ShowMessage", "show_message", []string{"text"}, nil)
		ClassDBBindMethod(t, "ShowGameOver", "show_game_over", nil, nil)
		ClassDBBindMethod(t, "ShowGameOverAwaitMessageTimerTimeout", "show_game_over_await_message_timer_timeout", nil, nil)
		ClassDBBindMethod(t, "ShowGameOverAwaitSceneTreeTimerTimeout", "show_game_over_await_scene_tree_timer_timeout", nil, nil)
		ClassDBBindMethod(t, "UpdateScore", "update_score", []string{"score"}, nil)

		// signals
		ClassDBAddSignal(t, "start_game")
	})
}

type HUD struct {
	CanvasLayerImpl
}

func (c *HUD) GetClassName() string {
	return "HUD"
}

func (c *HUD) GetParentClassName() string {
	return "CanvasLayer"
}

func (c *HUD) getScoreLabel() Label {
	gds := NewStringWithLatin1Chars("ScoreLabel")
	defer gds.Destroy()
	path := NewNodePathWithString(gds)
	defer path.Destroy()
	return ObjectCastTo(c.GetNode(path), "Label").(Label)
}

func (c *HUD) getMessageLabel() Label {
	gds := NewStringWithLatin1Chars("MessageLabel")
	defer gds.Destroy()
	path := NewNodePathWithString(gds)
	defer path.Destroy()
	return ObjectCastTo(c.GetNode(path), "Label").(Label)
}

func (c *HUD) getMessageTimer() Timer {
	gds := NewStringWithLatin1Chars("MessageTimer")
	defer gds.Destroy()
	path := NewNodePathWithString(gds)
	defer path.Destroy()
	return ObjectCastTo(c.GetNode(path), "Timer").(Timer)
}

func (c *HUD) getStartButton() Button {
	gds := NewStringWithLatin1Chars("StartButton")
	defer gds.Destroy()
	path := NewNodePathWithString(gds)
	defer path.Destroy()
	return ObjectCastTo(c.GetNode(path), "Button").(Button)
}

func (c *HUD) ShowMessage(text Variant) {
	// $MessageLabel.text = text
	messageLabel := c.getMessageLabel()
	gdsText := text.ToString()
	defer gdsText.Destroy()
	messageLabel.SetText(gdsText)

	// $MessageLabel.show()
	messageLabel.Show()

	// $MessageTimer.start()
	messageTimer := c.getMessageTimer()
	messageTimer.Start(-1)
}

func (c *HUD) ShowGameOver() {
	// show_message("Game Over")
	gameOverMessage := NewVariantGoString("Game Over")
	defer gameOverMessage.Destroy()
	c.ShowMessage(gameOverMessage)

	// await $MessageTimer.timeout
	messageTimer := c.getMessageTimer()
	gdsnTimeout := NewStringNameWithUtf8Chars("timeout")
	defer gdsnTimeout.Destroy()
	gdnsCallableMethodName := NewStringNameWithUtf8Chars("show_game_over_await_message_timer_timeout")
	defer gdnsCallableMethodName.Destroy()
	callable := NewCallableWithObjectStringName(c, gdnsCallableMethodName)
	defer callable.Destroy()
	err := messageTimer.Connect(gdsnTimeout, callable, OBJECT_CONNECT_FLAGS_CONNECT_ONE_SHOT)
	if err != OK {
		log.Panic("message timer connect failure", zap.Any("error", err))
	}
}

func (c *HUD) ShowGameOverAwaitMessageTimerTimeout() {
	// $MessageLabel.text = "Dodge the\nCreeps"
	messageLabel := c.getMessageLabel()
	gdsText := NewStringWithUtf8Chars("Dodge the\nCreeps")
	defer gdsText.Destroy()
	messageLabel.SetText(gdsText)

	// $MessageLabel.show()
	messageLabel.Show()

	// await get_tree().create_timer(1).timeout
	tree := c.GetTree()
	sceneTreeTimerRef := tree.CreateTimer(1, true, false, false)
	gdsnTimeout := NewStringNameWithUtf8Chars("timeout")
	defer gdsnTimeout.Destroy()
	gdnsCallableMethodName := NewStringNameWithUtf8Chars("show_game_over_await_scene_tree_timer_timeout")
	defer gdnsCallableMethodName.Destroy()
	callable := NewCallableWithObjectStringName(c, gdnsCallableMethodName)
	defer callable.Destroy()
	sceneTreeTimer := sceneTreeTimerRef.TypedPtr()
	err := sceneTreeTimer.Connect(gdsnTimeout, callable, OBJECT_CONNECT_FLAGS_CONNECT_ONE_SHOT)
	if err != OK {
		log.Panic("message timer connect failure", zap.Any("error", err))
	}
}

func (c *HUD) ShowGameOverAwaitSceneTreeTimerTimeout() {
	// $StartButton.show()
	startButton := c.getStartButton()
	startButton.Show()
}

func (c *HUD) UpdateScore(score Variant) {
	// $ScoreLabel.text = str(score)
	scoreLabel := c.getScoreLabel()
	gdsScore := score.ToString()
	defer gdsScore.Destroy()
	scoreLabel.SetText(gdsScore)
}

func (c *HUD) V_OnStartButtonPressed() {
	// $StartButton.hide()
	startButton := c.getStartButton()
	startButton.Hide()

	// start_game.emit()
	gdsnStartGame := NewStringNameWithUtf8Chars("start_game")
	defer gdsnStartGame.Destroy()
	c.EmitSignal(gdsnStartGame)
}

func (c *HUD) V_OnMessageTimerTimeout() {
	// $MessageLabel.hide()
	messageLabel := c.getMessageLabel()
	messageLabel.Hide()
}
