package demo

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func NewHUDFromOwnerObject(owner *GodotObject) GDClass {
	obj := &HUD{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

func RegisterClassHUD() {
	ClassDBRegisterClass(NewHUDFromOwnerObject, []GDExtensionPropertyInfo{}, nil, func(t *HUD) {
		// virtuals
		ClassDBBindMethodVirtual(t, "V_HUD_OnStartButtonPressed", "_on_StartButton_pressed", nil, nil)
		ClassDBBindMethodVirtual(t, "V_HUD_OnMessageTimerTimeout", "_on_MessageTimer_timeout", nil, nil)
		ClassDBBindMethodVirtual(t, "V_HUD_ExitTree", "_exit_tree", nil, nil)

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
	messageTimerConnected   bool
	sceneTreeTimerRef       RefSceneTreeTimer
	sceneTreeTimerConnected bool
}

func UnregisterClassHUD() {
	ClassDBUnregisterClass[*HUD]()
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
	err := messageTimer.Connect(gdsnTimeout, callable, uint32(OBJECT_CONNECT_FLAGS_CONNECT_ONE_SHOT))
	if err != OK {
		log.Panic("message timer connect failure", zap.Any("error", err))
	}
	c.messageTimerConnected = true
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
	if c.sceneTreeTimerConnected {
		c.sceneTreeTimerRef.Unref()
		c.sceneTreeTimerConnected = false
	}
	gdsnTimeout := NewStringNameWithUtf8Chars("timeout")
	defer gdsnTimeout.Destroy()
	gdnsCallableMethodName := NewStringNameWithUtf8Chars("show_game_over_await_scene_tree_timer_timeout")
	defer gdnsCallableMethodName.Destroy()
	callable := NewCallableWithObjectStringName(c, gdnsCallableMethodName)
	defer callable.Destroy()
	c.sceneTreeTimerRef = sceneTreeTimerRef
	sceneTreeTimer := sceneTreeTimerRef.Ptr()
	err := sceneTreeTimer.Connect(gdsnTimeout, callable, uint32(OBJECT_CONNECT_FLAGS_CONNECT_ONE_SHOT))
	if err != OK {
		log.Panic("message timer connect failure", zap.Any("error", err))
	}
	c.sceneTreeTimerConnected = true
}

func (c *HUD) ShowGameOverAwaitSceneTreeTimerTimeout() {
	// release the scene tree timer ref held since the message timer timeout
	if c.sceneTreeTimerConnected {
		c.sceneTreeTimerRef.Unref()
		c.sceneTreeTimerConnected = false
	}
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

func (c *HUD) V_HUD_OnStartButtonPressed() {
	// $StartButton.hide()
	startButton := c.getStartButton()
	startButton.Hide()

	// start_game.emit()
	gdsnStartGame := NewStringNameWithUtf8Chars("start_game")
	defer gdsnStartGame.Destroy()
	c.EmitSignal(gdsnStartGame)
}

func (c *HUD) V_HUD_OnMessageTimerTimeout() {
	c.messageTimerConnected = false
	// $MessageLabel.hide()
	messageLabel := c.getMessageLabel()
	messageLabel.Hide()
}

func (c *HUD) Cleanup() {
	log.Info("HUD.Cleanup, disconnecting remaining signal connections",
		zap.Bool("messageTimerConnected", c.messageTimerConnected),
		zap.Bool("sceneTreeTimerConnected", c.sceneTreeTimerConnected))
	if c.messageTimerConnected {
		messageTimer := c.getMessageTimer()
		timeoutSN := NewStringNameWithUtf8Chars("timeout")
		msgTimerCallableSN := NewStringNameWithUtf8Chars("show_game_over_await_message_timer_timeout")
		callable := NewCallableWithObjectStringName(c, msgTimerCallableSN)
		if messageTimer.IsConnected(timeoutSN, callable) {
			messageTimer.Disconnect(timeoutSN, callable)
		}
		callable.Destroy()
		msgTimerCallableSN.Destroy()
		timeoutSN.Destroy()
		c.messageTimerConnected = false
	}
	if c.sceneTreeTimerConnected {
		sceneTreeTimer := c.sceneTreeTimerRef.Ptr()
		timeoutSN := NewStringNameWithUtf8Chars("timeout")
		stTimerCallableSN := NewStringNameWithUtf8Chars("show_game_over_await_scene_tree_timer_timeout")
		callable := NewCallableWithObjectStringName(c, stTimerCallableSN)
		if sceneTreeTimer.IsConnected(timeoutSN, callable) {
			sceneTreeTimer.Disconnect(timeoutSN, callable)
		}
		callable.Destroy()
		stTimerCallableSN.Destroy()
		timeoutSN.Destroy()
		c.sceneTreeTimerRef.Unref()
		c.sceneTreeTimerConnected = false
	}
}

func (c *HUD) V_HUD_ExitTree() {
	c.Cleanup()
}
