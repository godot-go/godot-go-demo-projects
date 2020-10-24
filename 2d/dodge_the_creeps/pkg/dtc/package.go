package dtc

import "github.com/godot-go/godot-go/pkg/gdnative"

var (
	rng gdnative.RandomNumberGenerator
)

func InitGlobals() {
	rng = gdnative.NewRandomNumberGenerator()
}

func DestroyGlobals() {
	// rng.Destroy()
}
