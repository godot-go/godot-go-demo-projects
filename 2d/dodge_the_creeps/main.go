package main

//go:generate go run cmd/main.go --gdnative --types --classes

import (
	_ "github.com/godot-go/godot-go-demo-projects/2d/dodge_the_creeps/pkg/export"
	_ "github.com/godot-go/godot-go-demo-projects/2d/dodge_the_creeps/pkg/dtc"
)

func main() {
}
