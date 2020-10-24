package dtc

import (
	"github.com/godot-go/godot-go/pkg/gdnative"
)

type Main struct {
	gdnative.NodeImpl
	gdnative.UserDataIdentifiableImpl

	score int64

	mob gdnative.PackedScene
	hud HUD
}

func (p *Main) ClassName() string {
	return "Main"
}

func (p *Main) BaseClass() string {
	return "Node"
}

func (p *Main) Init() {
}

func (p *Main) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	// methods
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("game_over", "GameOver")
	e.RegisterMethod("new_game", "NewGame")
	e.RegisterMethod("_on_MobTimer_timeout", "OnMobTimerTimeout")
	e.RegisterMethod("_on_ScoreTimer_timeout", "OnScoreTimerTimeout")
	e.RegisterMethod("_on_StartTimer_timeout", "OnStartTimerTimeout")

	// signals
	e.RegisterSignal("start_game")
}

func (p *Main) Ready() {
	p.hud = NewHUDWithOwner(p.FindNode("HUD", true, true).GetOwnerObject())
	rng.Randomize()
}

func (p *Main) GameOver() {
	gdnative.NewTimerWithOwner(p.FindNode("ScoreTimer", true, true).GetOwnerObject()).Stop()
	gdnative.NewTimerWithOwner(p.FindNode("MobTimer", true, true).GetOwnerObject()).Stop()
	p.hud.ShowGameOver()
	gdnative.NewAudioStreamPlayerWithOwner(p.FindNode("Music", true, true).GetOwnerObject()).Stop()
	gdnative.NewAudioStreamPlayerWithOwner(p.FindNode("DeathSound", true, true).GetOwnerObject()).Play(0.0)
}


func (p *Main) NewGame() {
	p.score = 0
	pos := gdnative.NewPosition2DWithOwner(p.FindNode("StartPosition", true, true).GetOwnerObject()).GetPosition()
	player := NewPlayerWithOwner(p.FindNode("Player", true, true).GetOwnerObject())
	player.Start(pos)
	gdnative.NewTimerWithOwner(p.FindNode("StartTimer", true, true).GetOwnerObject()).Start(-1)
	p.hud.UpdateScore(0)
	p.hud.ShowMessage("Get Ready")
	gdnative.NewAudioStreamPlayerWithOwner(p.FindNode("Music", true, true).GetOwnerObject()).Play(0.0)
}


// func (p *Main) OnMobTimerTimeout() {
// 	$MobPath/MobSpawnLocation.offset = randi()
// 	var mob = Mob.instance()
// 	add_child(mob)
// 	var direction = $MobPath/MobSpawnLocation.rotation + TAU / 4
// 	mob.position = $MobPath/MobSpawnLocation.position
// 	direction += rand_range(-TAU / 8, TAU / 8)
// 	mob.rotation = direction
// 	mob.linear_velocity = Vector2(rand_range(mob.min_speed, mob.max_speed), 0).rotated(direction)
// 	# warning-ignore:return_value_discarded
// 	$HUD.connect("start_game", mob, "_on_start_game")
// }


// func (p *Main) OnScoreTimerTimeout() {
// 	score += 1
// 	$HUD.update_score(score)
// }

// func (p *Main) OnStartTimerTimeout() {
// 	$MobTimer.start()
// 	$ScoreTimer.start()
// }