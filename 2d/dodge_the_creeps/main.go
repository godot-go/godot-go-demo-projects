package main

import "C"
import (
	"godot-go-demo-projects/2d/dodgethecreep/pkg/demo"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdextension"
	"github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
)

//export GodotGoDemo2DDodgeTheCreepsInit
func GodotGoDemo2DDodgeTheCreepsInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Debug("GodotGoDemo2DDodgeTheCreepsInit called")
	initObj := gdextension.NewInitObject(
		(gdextensionffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(gdextensionffi.GDExtensionClassLibraryPtr)(p_library),
		(*gdextensionffi.GDExtensionInitialization)(unsafe.Pointer(r_initialization)),
	)

	initObj.RegisterSceneInitializer(func() {
		demo.RegisterClassHUD()
	})

	initObj.RegisterSceneTerminator(func() {
	})

	return initObj.Init()
}

func main() {
	// log.Trace("this application is meant to be run as a plugin to godot")
}