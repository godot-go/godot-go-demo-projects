package dtc

import (
	"math"

	"github.com/godot-go/godot-go/pkg/gdnative"
)


type Player struct {
	gdnative.Area2DImpl
	gdnative.UserDataIdentifiableImpl

	speed gdnative.Variant
	screenSize gdnative.Vector2
}

func (p *Player) ClassName() string {
	return "Player"
}

func (p *Player) BaseClass() string {
	return "Area2D"
}

func (p *Player) Init() {
	p.speed = gdnative.NewVariantReal(400.0)
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
	e.RegisterProperty("speed", "SetSpeed", "GetSpeed", p.speed)
}

func (p *Player) Ready() {
	rect := p.GetViewportRect()

	// directly setting the screen_size is fine since this
	// isn't exposed as a Godot property
	p.screenSize = rect.GetSize()
	p.Hide()
}

func (p *Player) Process(delta float64) {
	input := gdnative.GetSingletonInput()

	x := input.GetActionStrength("move_right") - input.GetActionStrength("move_left")
	y := input.GetActionStrength("move_down") - input.GetActionStrength("move_up")

	var velocity = gdnative.NewVector2(x, y)

	animatedSprite := gdnative.NewAnimatedSpriteWithOwner(p.GetNode(gdnative.NewNodePath("AnimatedSprite")).GetOwnerObject())

	if velocity.Length() > 0 {
		v1 := velocity.Normalized()
		velocity = v1.OperatorMultiplyScalar(float32(p.speed.AsReal()))
		animatedSprite.Play("", false)
	} else {
		animatedSprite.Stop()
	}

	pos := p.GetPosition()
	newPos := pos.OperatorAdd(velocity.OperatorMultiplyScalar(float32(delta)))
	newPos.SetX(clamp(newPos.GetX(), 0, p.screenSize.GetX()))
	newPos.SetY(clamp(newPos.GetY(), 0, p.screenSize.GetY()))

	p.SetPosition(newPos)

	velX := velocity.GetX()
	velY := velocity.GetY()

	if velX != 0 {
		animatedSprite.SetAnimation("right")
		animatedSprite.SetFlipV(false)
		animatedSprite.SetFlipH(velX < 0)
	} else if velY != 0 {
		animatedSprite.SetAnimation("up")
		animatedSprite.SetFlipV(velY > 0)
	}
}

func (p *Player) Start(pos gdnative.Vector2) {
	p.SetPosition(pos)
	p.Show()
	collisionShape2D := gdnative.NewCollisionShape2DWithOwner(p.GetNode(gdnative.NewNodePath("CollisionShape2D")).GetOwnerObject())
	collisionShape2D.SetDisabled(false)
}

func (p *Player) OnPlayerBodyEntered(_body interface{}) {
	p.Hide()
	p.EmitSignal("hit")
	collisionShape2D := gdnative.NewCollisionShape2DWithOwner(p.GetNode(gdnative.NewNodePath("CollisionShape2D")).GetOwnerObject())
	collisionShape2D.SetDeferred("disabled", gdnative.NewVariantBool(true))
}

func (p *Player) GetSpeed() gdnative.Variant {
	return p.speed
}

func (p *Player) SetSpeed(v gdnative.Variant) {
	newSpeed := v.AsReal()

	p.speed.Destroy()
	p.speed = gdnative.NewVariantReal(newSpeed)
}

func clamp(v, min, max float32) float32 {
	return float32(math.Max(math.Min(float64(v), float64(max)), float64(min)))
}

func NewPlayerWithOwner(owner *gdnative.GodotObject) Player {
	inst := gdnative.GetCustomClassInstanceWithOwner(owner).(*Player)
	return *inst
}

func init() {
	gdnative.RegisterInitCallback(func() {
		gdnative.RegisterClass(&Player{})
	})
}
