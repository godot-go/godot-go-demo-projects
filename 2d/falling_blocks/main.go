package main

import "C"
import (
	"godot-go-demo-projects/2d/falling_blocks/pkg/demo"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

//export GodotGoDemo2DFallingBlocksInit
func GodotGoDemo2DFallingBlocksInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Debug("GodotGoDemo2DFallingBlocksInit called")
	initObj := core.NewInitObject(
		(ffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(ffi.GDExtensionClassLibraryPtr)(p_library),
		(*ffi.GDExtensionInitialization)(unsafe.Pointer(r_initialization)),
	)

	initObj.RegisterSceneInitializer(func() {
		demo.RegisterClassFallingBlocksTitle()
		demo.RegisterClassFallingBlocksLobby()
		demo.RegisterClassFallingBlocksGame()
		demo.RegisterClassFallingBlocksSession()
	})

	return initObj.Init()
}

func main() {
	// this application is meant to be run as a plugin to godot
}
