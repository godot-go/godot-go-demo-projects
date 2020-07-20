package main

import (
	_ "godot-go-demo-projects/2d/topdown/pkg/export"
	"godot-go-demo-projects/2d/topdown/pkg/demo"
	"github.com/pcting/godot-go/pkg/gdnative"
	"github.com/pcting/godot-go/pkg/log"
)

func init() {
	gdnative.RegisterInitCallback(initNativescript)
	gdnative.RegisterInitCallback(demo.PlayerCharacterNativescriptInit)
	gdnative.RegisterTerminateCallbacks(demo.PlayerCharacterNativescriptTerminate)
}

func initNativescript() {
	log.SetLevel(log.TraceLevel)
	log.Trace("initNativescript called")

	os := gdnative.GetSingletonOS()

	if os.IsDebugBuild() {
		log.Info("running Godot debug build!")
	}

	gdnative.RegisterClass(&demo.PlayerCharacter{}, demo.PlayerCharacterCreateFunc)
}

func main() {
	log.Trace("this application is meant to be run as a plugin to godot")
}
