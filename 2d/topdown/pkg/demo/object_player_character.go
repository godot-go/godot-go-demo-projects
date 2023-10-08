package demo

import (
	. "github.com/godot-go/godot-go/pkg/gdextension"
	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"strings"
	"unsafe"
)

const (
	TileSize = 16
)

func RegisterClassPlayerCharacter() {
	ClassDBRegisterClass(&PlayerCharacter{}, []GDExtensionPropertyInfo{}, nil, func(t GDClass) {
		// virtuals
		ClassDBBindMethodVirtual(t, "V_Input", "_input", []string{"event"}, nil)
		ClassDBBindMethodVirtual(t, "V_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_PhysicsProcess", "_physics_process", nil, nil)
		ClassDBBindMethodVirtual(t, "V_Set", "_set", []string{"name", "value"}, nil)
		ClassDBBindMethodVirtual(t, "V_Get", "_get", []string{"name"}, nil)

		// properties
		ClassDBBindMethod(t, "GetDirection", "get_direction", nil, nil)
		ClassDBBindMethod(t, "SetDirection", "set_direction", []string{"id"}, nil)
		ClassDBAddProperty(t, GDEXTENSION_VARIANT_TYPE_VECTOR2, "entity", "set_direction", "get_direction")

		// signals
		ClassDBAddSignal(t, "moved",
			SignalParam{
				Type: GDEXTENSION_VARIANT_TYPE_VECTOR2,
				Name: "direction"},
		)
	})
}

type PlayerCharacter struct {
	CharacterBody2DImpl
	walkAnimation AnimationPlayer
	direction     Vector2
	speed         float32
	input         Input
}

func (p *PlayerCharacter) GetClassName() string {
	return "PlayerCharacter"
}

func (p *PlayerCharacter) GetParentClassName() string {
	return "CharacterBody2D"
}

func (h *PlayerCharacter) V_Set(name string, value Variant) bool {
	switch name {
	case "direction":
		h.direction = value.ToVector2()
		vDir := NewVariantVector2(h.direction)
		defer vDir.Destroy()
		log.Info("V_Set",
			zap.Any("direction", Stringify(vDir)),
		)
		return true
	}
	return false
}

func (h *PlayerCharacter) V_Get(name string) (Variant, bool) {
	switch name {
	case "direction":
		vDir := NewVariantVector2(h.direction)
		log.Info("V_Get",
			zap.Any("direction", Stringify(vDir)),
		)
		return vDir, true
	}
	return Variant{}, false
}

func (h *PlayerCharacter) GetDirection() Vector2 {
	return h.direction
}

func (h *PlayerCharacter) SetDirection(v Vector2) {
	h.direction = v
}

func (h *PlayerCharacter) V_Input(refInputEvent RefInputEvent) {
	event := refInputEvent.TypedPtr()
	if event == nil {
		log.Warn("PlayerCharacter.V_Input: null refEvent parameter")
		return
	}

	// BUG: godot-go isn't properly wrapping the go struct to the underlying type.
	//      i believe we have to call a gdextension interface cast it to the
	//      underlying type as a workaround
	//
	// switch event.(type) {
	// case InputEventKey:
	//   h.direction = input.GetVector(uiLeft, uiRight, uiUp, uiDown, -1.0)
	// }
	dir := h.input.GetVector(uiLeft, uiRight, uiUp, uiDown, -1.0)
	vDir := NewVariantVector2(dir)
	log.Info("V_Input",
		zap.Any("dir", Stringify(vDir)),
	)
	h.SetDirection(dir)
}

func (h *PlayerCharacter) V_Ready() {
	h.input = GetInputSingleton()
	if h.input == nil {
		log.Panic("unable to get input singleton")
	}
	h.speed = 5.0
	p := NewNodePathWithString(NewStringWithLatin1Chars("sprite/animation_player"))
	str := p.GetConcatenatedSubnames()
	defer str.Destroy()
	log.Info("searching path...", zap.String("names", str.ToUtf8()))
	n := h.GetNode(p)
	pno := n.GetGodotObjectOwner()
	h.walkAnimation = NewAnimationPlayerWithGodotOwnerObject(pno)
	if !h.walkAnimation.HasAnimation(walkRight) {
		log.Panic("unable to find walk-right animation")
	}
	if !h.walkAnimation.HasAnimation(walkLeft) {
		log.Panic("unabel to find walk-left animation")
	}
	if !h.walkAnimation.HasAnimation(walkDown) {
		log.Panic("unable to find walk-down")
	}
	if !h.walkAnimation.HasAnimation(walkUp) {
		log.Panic("unable to find walk-up")
	}
	if !h.walkAnimation.HasAnimation(idleRight) {
		log.Panic("unable to find idle-right")
	}
	if !h.walkAnimation.HasAnimation(idleLeft) {
		log.Panic("unable to find idle-left")
	}
	if !h.walkAnimation.HasAnimation(idleDown) {
		log.Panic("unable to find idle-down")
	}
	if !h.walkAnimation.HasAnimation(idleUp) {
		log.Panic("unable to find idle-up")
	}
}

func (h *PlayerCharacter) V_PhysicsProcess(delta float64) {
	dir := h.direction
	h.updateSprite(dir)
	calcV := dir.Multiply_float(float32(delta) * h.speed * TileSize)
	h.MoveAndCollide(calcV, false, 0.785398, true)

	// emit signal on position change for UI to refresh
	vPos := NewVariantVector2(h.GetPosition())
	defer vPos.Destroy()
	h.EmitSignal(moved, vPos)
}

func (h *PlayerCharacter) updateSprite(dqir Vector2) {
	dir := h.direction
	x := dir.MemberGetx()
	y := dir.MemberGety()

	a := h.walkAnimation
	ca := a.GetCurrentAnimation()
	pca := &ca

	if x > 0 {
		if !pca.Equal_StringName(walkRight) {
			a.Play(walkRight, -1, 1.0, false)
		}
	} else if x < 0 {
		if !pca.Equal_StringName(walkLeft) {
			a.Play(walkLeft, -1, 1.0, true)
		}
	} else if y > 0 {
		if !pca.Equal_StringName(walkDown) {
			a.Play(walkDown, -1, 1.0, false)
		}
	} else if y < 0 {
		if !pca.Equal_StringName(walkUp) {
			a.Play(walkUp, -1, 1.0, false)
		}
	} else {
		// switch to idle animation if the character isn't moving
		name := pca.ToUtf8()

		if name != "" {
			tokens := strings.Split(name, "-")

			if len(tokens) != 2 {
				log.Panic("unable to parse animation name", zap.String("name", name))
			}

			var animationName StringName
			switch tokens[1] {
			case "up":
				animationName = idleUp
			case "down":
				animationName = idleDown
			case "left":
				animationName = idleLeft
			case "right":
				animationName = idleRight
			default:
				// log.WithField("name", name).Warn("unhandled animation name")
			}

			if !pca.Equal_StringName(animationName) {
				log.Info("switch animation",
					zap.String("name", animationName.ToUtf8()),
				)
				a.Play(animationName, -1, 1.0, false)
			}
		}
	}
}

func (p *PlayerCharacter) Free() {
	// log.WithFields(gdnative.WithObject(p.GetGodotObjectOwner())).Trace("free PlayerCharacter")
	if p.input != nil{
		p.input.Destroy()
	}

	p.walkAnimation = nil

	if p != nil {
		Free(unsafe.Pointer(p))
		p = nil
	}
}

func NewPlayerCharacter() GDClass {
	return CreateGDClassInstance("PlayerCharacter")
}

var (
	moved           StringName
	velocity        StringName
	velocityVariant Variant

	uiRight StringName
	uiLeft  StringName
	uiUp    StringName
	uiDown  StringName

	walkRight StringName
	walkLeft  StringName
	walkUp    StringName
	walkDown  StringName

	idleRight StringName
	idleLeft  StringName
	idleUp    StringName
	idleDown  StringName
)

func PlayerCharacterNativescriptInit() {
	moved = NewStringNameWithLatin1Chars("moved")
	velocity = NewStringNameWithLatin1Chars("velocity")
	velocityVariant = NewVariantStringName(velocity)

	uiRight = NewStringNameWithLatin1Chars("ui_right")
	uiLeft = NewStringNameWithLatin1Chars("ui_left")
	uiUp = NewStringNameWithLatin1Chars("ui_up")
	uiDown = NewStringNameWithLatin1Chars("ui_down")

	walkRight = NewStringNameWithLatin1Chars("walk-right")
	walkLeft = NewStringNameWithLatin1Chars("walk-left")
	walkUp = NewStringNameWithLatin1Chars("walk-up")
	walkDown = NewStringNameWithLatin1Chars("walk-down")

	idleRight = NewStringNameWithLatin1Chars("idle-right")
	idleLeft = NewStringNameWithLatin1Chars("idle-left")
	idleUp = NewStringNameWithLatin1Chars("idle-up")
	idleDown = NewStringNameWithLatin1Chars("idle-down")
}

func PlayerCharacterNativescriptTerminate() {
	moved.Destroy()
	velocity.Destroy()
	velocityVariant.Destroy()

	uiRight.Destroy()
	uiLeft.Destroy()
	uiUp.Destroy()
	uiDown.Destroy()

	walkRight.Destroy()
	walkLeft.Destroy()
	walkUp.Destroy()
	walkDown.Destroy()

	idleRight.Destroy()
	idleLeft.Destroy()
	idleUp.Destroy()
	idleDown.Destroy()
}
