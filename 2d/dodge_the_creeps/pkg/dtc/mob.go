package dtc

import (
	"math/rand"

	"github.com/godot-go/godot-go/pkg/gdnative"
)

var (
	mobTypes = []string{"walk", "swim", "fly"}
)

type Mob struct {
	gdnative.RigidBody2DImpl
	gdnative.UserDataIdentifiableImpl

	minSpeed gdnative.Variant
	maxSpeed gdnative.Variant
}

func (p *Mob) ClassName() string {
	return "Mob"
}

func (p *Mob) BaseClass() string {
	return "RigidBody2D"
}

func (p *Mob) Init() {
	p.minSpeed = gdnative.NewVariantReal(150.0)
	p.maxSpeed = gdnative.NewVariantReal(250.0)
}

func (p *Mob) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	// methods
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("_on_VisibilityNotifier2D_screen_exited", "OnVisibilityNotifier2DScreenExited")
	e.RegisterMethod("_on_start_game", "OnStartGame")

	// properties
	e.RegisterProperty("min_speed", "SetMinSpeed", "GetMinSpeed", p.minSpeed)
	e.RegisterProperty("max_speed", "SetMaxSpeed", "GetMaxSpeed", p.maxSpeed)
}

func (p *Mob) Ready() {
	animatedSprite := gdnative.NewAnimatedSpriteWithOwner(p.FindNode("AnimatedSprite", true, true).GetOwnerObject())
	animation := mobTypes[rand.Intn(len(mobTypes))]
	animatedSprite.SetAnimation(animation)
}

func (p *Mob) OnVisibilityNotifier2DScreenExited() {
	p.QueueFree()
}

func (p *Mob) OnStartGame() {
	p.QueueFree()
}

func (p *Mob) GetMinSpeed() gdnative.Variant {
	return p.minSpeed
}

func (p *Mob) SetMinSpeed(v gdnative.Variant) {
	newSpeed := v.AsReal()

	p.minSpeed.Destroy()
	p.minSpeed = gdnative.NewVariantReal(newSpeed)
}

func (p *Mob) GetMaxSpeed() gdnative.Variant {
	return p.maxSpeed
}

func (p *Mob) SetMaxSpeed(v gdnative.Variant) {
	newSpeed := v.AsReal()

	p.maxSpeed.Destroy()
	p.maxSpeed = gdnative.NewVariantReal(newSpeed)
}

func NewMobWithOwner(owner *gdnative.GodotObject) Mob {
	inst := gdnative.GetCustomClassInstanceWithOwner(owner).(*Mob)
	return *inst
}

func init() {
	gdnative.RegisterInitCallback(func() {
		gdnative.RegisterClass(&Mob{})
	})
}
