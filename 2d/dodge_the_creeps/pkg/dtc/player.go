package dtc

import (
	"math"

	"github.com/godot-go/godot-go/pkg/gdnative"
)


type Player struct {
	gdnative.Area2DImpl
	gdnative.UserDataIdentifiableImpl

	position gdnative.Vector2
	speed int64
	screen_size gdnative.Vector2

	animatedSprite gdnative.AnimatedSprite
	collisionShape2D gdnative.CollisionShape2D
}

func (p *Player) ClassName() string {
	return "Player"
}

func (p *Player) BaseClass() string {
	return "Area2D"
}

func (p *Player) Init() {
}

func (p *Player) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	// methods
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("_process", "Process")
	e.RegisterMethod("start", "Start")
	e.RegisterMethod("_on_Player_body_entered", "OnPlayerBodyEntered")

	// signals
	e.RegisterSignal("hit")

	// properties
	e.RegisterProperty("speed", "SetSpeed", "GetSpeed", gdnative.NewVariantInt(400))
}

func (p *Player) Ready() {
	p.animatedSprite = gdnative.NewAnimatedSpriteWithOwner(p.FindNode("AnimatedSprite", true, true).GetOwnerObject())
	p.collisionShape2D = gdnative.NewCollisionShape2DWithOwner(p.FindNode("CollisionShape2D", true, true).GetOwnerObject())
	rect := p.GetViewportRect()
	p.screen_size = rect.GetSize()
	// p.speed = 400
	p.Hide()
}

func (p *Player) Process(delta float32) {
	var velocity = gdnative.NewVector2(0, 0)
	input := gdnative.GetSingletonInput()

	velocity.SetX(input.GetActionStrength("move_right") - input.GetActionStrength("move_left"))
	velocity.SetY(input.GetActionStrength("move_down") - input.GetActionStrength("move_up"))

	if velocity.Length() > 0 {
		v1 := velocity.Normalized()
		velocity = v1.OperatorMultiplyScalar(float32(p.speed))
		p.animatedSprite.Play("", false)
	} else {
		p.animatedSprite.Stop()
	}
	incrVelocity := velocity.OperatorMultiplyScalar(delta)
	p.position = p.position.OperatorAdd(incrVelocity)
	p.position.SetX(clamp(p.position.GetX(), 0, p.screen_size.GetX()))
	p.position.SetY(clamp(p.position.GetY(), 0, p.screen_size.GetY()))

	velX := velocity.GetX()
	velY := velocity.GetY()

	if velX != 0 {
		p.animatedSprite.SetAnimation("right")
		p.animatedSprite.SetFlipV(false)
		p.animatedSprite.SetFlipH(velX < 0)
	} else if velY != 0 {
		p.animatedSprite.SetAnimation("up")
		p.animatedSprite.SetFlipV(velY > 0)
	}
}

func (p *Player) Start(pos gdnative.Vector2) {
	p.position = pos
	p.Show()
	p.collisionShape2D.SetDisabled(false)
}


func (p *Player) OnPlayerBodyEntered(_body interface{}) {
	p.Hide()
	p.EmitSignal("hit")
	p.collisionShape2D.SetDisabled(true)
}

func (p *Player) GetSpeed() int64 {
	return p.speed
}

func (p *Player) SetSpeed(v int64) {
	p.speed = v
}

func clamp(v, min, max float32) float32 {
	return float32(math.Max(math.Min(float64(v), float64(max)), float64(min)))
}

func NewPlayerWithOwner(owner *gdnative.GodotObject) Player {
	inst := gdnative.GetCustomClassInstanceWithOwner(owner).(*Player)
	return *inst
}