//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type BuildPlatform struct {
	OS   string
	Arch string
}

var (
	godotBin     = "godot"
)

func envWithPlatform(platform BuildPlatform) map[string]string {
	return map[string]string{
		"GOOS":              platform.OS,
		"GOARCH":            platform.Arch,
		"CGO_LDFLAGS_ALLOW": "pkg-config",
		"CGO_ENABLED":       "1",
	}
}

func buildGodotPlugin(appPath string, platform BuildPlatform) error {
	return sh.RunWith(envWithPlatform(platform), mg.GoCmd(), "build", "-x", "-work",
		"-buildmode=c-shared", "-ldflags=\"-d=checkptr -compressdwarf=false\"", "-gcflags=\"all=-N -l\"",
		"-o", filepath.Join(appPath, "project", "libs", platform.godotPluginCSharedName(appPath)),
		filepath.Join(appPath, "main.go"),
	)
}

func runPlugin(appPath string) error {
	return sh.RunWith(map[string]string{"asyncpremptoff": "1", "cgocheck": "2"}, godotBin, "--verbose", "-v", "-d", "--path", filepath.Join(appPath, "project"))
}

func debugPlugin(appPath string) error {
	return sh.RunWith(map[string]string{"asyncpremptoff": "1", "cgocheck": "2"}, "gdb", "--args", godotBin, "-v", "-d", "--path", filepath.Join(appPath, "project"))
}

func (p *BuildPlatform) godotPluginCSharedName(appPath string) string {
	return fmt.Sprintf("libgodotgo-2d-topdown-%s-%s.so",  p.OS, p.Arch)
}

func Build() error {
	return buildGodotPlugin(
		".",
		BuildPlatform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	)
}

func RunGodot() error {
	mg.Deps(Build)

	return runPlugin(".",)
}

func DebugGodot() error {
	return debugPlugin(".")
}

var Default = Build
