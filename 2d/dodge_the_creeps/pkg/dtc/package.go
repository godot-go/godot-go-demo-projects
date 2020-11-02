package dtc

import "github.com/godot-go/godot-go/pkg/gdnative"

var (
	rng gdnative.RandomNumberGenerator
)

func init() {
	gdnative.RegisterInitCallback(func() {
		rng = gdnative.NewRandomNumberGenerator()
	})
}
