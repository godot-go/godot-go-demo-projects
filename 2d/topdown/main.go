package main

import "C"
import (
	"godot-go-demo-projects/2d/topdown/pkg/demo"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

//export GodotGoDemo2DTopDownInit
func GodotGoDemo2DTopDownInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Debug("ExampleLibraryInit called")
	initObj := core.NewInitObject(
		(ffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(ffi.GDExtensionClassLibraryPtr)(p_library),
		(*ffi.GDExtensionInitialization)(unsafe.Pointer(r_initialization)),
	)

	initObj.RegisterSceneInitializer(func() {
		demo.PlayerCharacterGDExtensionInit()
		demo.RegisterClassPlayerCharacter()
	})

	initObj.RegisterSceneTerminator(func() {
		demo.PlayerCharacterGDExtensionTerminate()
	})

	return initObj.Init()
}

func main() {
	// log.Trace("this application is meant to be run as a plugin to godot")
}