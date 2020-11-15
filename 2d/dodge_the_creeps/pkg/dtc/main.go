package dtc

import (
	"math"
	"math/rand"

	"github.com/godot-go/godot-go/pkg/gdnative"
)

type Main struct {
	gdnative.NodeImpl
	gdnative.UserDataIdentifiableImpl

	score int64
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
}

func (p *Main) Ready() {
	rng.Randomize()

	binds := gdnative.NewArray()
	defer binds.Destroy()

	hud := NewHUDWithOwner(p.FindNode("HUD", true, true).GetOwnerObject())
	hud.Connect("start_game", p, "new_game", binds, 0)

	player := NewPlayerWithOwner(p.FindNode("Player", true, true).GetOwnerObject())
	player.Connect("hit", p, "game_over", binds, 0)
}

func (p *Main) GameOver() {
	gdnative.NewTimerWithOwner(p.FindNode("ScoreTimer", true, true).GetOwnerObject()).Stop()
	gdnative.NewTimerWithOwner(p.FindNode("MobTimer", true, true).GetOwnerObject()).Stop()

	hud := NewHUDWithOwner(p.FindNode("HUD", true, true).GetOwnerObject())
	hud.ShowGameOver()

	gdnative.NewAudioStreamPlayerWithOwner(p.FindNode("Music", true, true).GetOwnerObject()).Stop()
	gdnative.NewAudioStreamPlayerWithOwner(p.FindNode("DeathSound", true, true).GetOwnerObject()).Play(0.0)
}

func (p *Main) NewGame() {
	p.score = 0
	pos := gdnative.NewPosition2DWithOwner(p.FindNode("StartPosition", true, true).GetOwnerObject()).GetPosition()
	player := NewPlayerWithOwner(p.FindNode("Player", true, true).GetOwnerObject())
	player.Start(pos)
	gdnative.NewTimerWithOwner(p.FindNode("StartTimer", true, true).GetOwnerObject()).Start(-1)
	hud := NewHUDWithOwner(p.FindNode("HUD", true, true).GetOwnerObject())
	hud.UpdateScore(0)
	hud.showMessage("Get Ready")
	gdnative.NewAudioStreamPlayerWithOwner(p.FindNode("Music", true, true).GetOwnerObject()).Play(0.0)
}

func (p *Main) OnMobTimerTimeout() {
	mobSpawnLocationNodePath := gdnative.NewNodePath("MobPath/MobSpawnLocation")
	mobSpawnLocation := gdnative.NewPathFollow2DWithOwner(p.GetNode(mobSpawnLocationNodePath).GetOwnerObject())
	mobSpawnLocation.SetOffset(float32(rand.Int()))

	// properties
	resLoader := gdnative.GetSingletonResourceLoader()
	res := resLoader.Load("res://Mob.tscn", "", false)
	mobScene := gdnative.NewPackedSceneWithOwner(res.GetOwnerObject())

	mobInst := mobScene.Instance(int64(gdnative.PACKED_SCENE_GEN_EDIT_STATE_DISABLED))

	mob := NewMobWithOwner(mobInst.GetOwnerObject())

	p.AddChild(&mob, false)

	tau := float32(math.Pi * 2)

	direction := mobSpawnLocation.GetRotation() + tau/4

	mob.SetPosition(mobSpawnLocation.GetPosition())

	direction += randRange(-tau/8, tau/8)

	mob.SetRotation(direction)

	mobMinSpeed := mob.GetMinSpeed()
	mobMaxSpeed := mob.GetMaxSpeed()

	x := randRange(float32(mobMinSpeed.AsReal()), float32(mobMaxSpeed.AsReal()))
	linearVelocity := gdnative.NewVector2(x, 0)
	linearVelocity = linearVelocity.Rotated(direction)

	mob.SetLinearVelocity(linearVelocity)

	hud := NewHUDWithOwner(p.FindNode("HUD", true, true).GetOwnerObject())
	binds := gdnative.NewArray()
	defer binds.Destroy()
	hud.Connect("start_game", &mob, "_on_start_game", binds, 0)
}

func randRange(min, max float32) float32 {
	diff := max - min

	return (rand.Float32() * diff) + min
}

func (p *Main) OnScoreTimerTimeout() {
	p.score++

	hud := NewHUDWithOwner(p.FindNode("HUD", true, true).GetOwnerObject())
	hud.UpdateScore(p.score)
}

func (p *Main) OnStartTimerTimeout() {
	gdnative.NewTimerWithOwner(p.FindNode("MobTimer", true, true).GetOwnerObject()).Start(-1)
	gdnative.NewTimerWithOwner(p.FindNode("ScoreTimer", true, true).GetOwnerObject()).Start(-1)
}

func NewMainWithOwner(owner *gdnative.GodotObject) Main {
	inst := gdnative.GetCustomClassInstanceWithOwner(owner).(*Main)
	return *inst
}

func init() {
	gdnative.RegisterInitCallback(func() {
		gdnative.RegisterClass(&Main{})
	})
}
