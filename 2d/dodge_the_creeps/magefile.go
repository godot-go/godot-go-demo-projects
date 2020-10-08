//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"fmt"
	"os"
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
	godotBin   string
	ci         bool
	targetOS   string
	targetArch string
)

func init() {
	var (
		ok  bool
	)

	if targetOS, ok = os.LookupEnv("TARGET_OS"); !ok {
		targetOS = runtime.GOOS
	}

	if targetArch, ok = os.LookupEnv("TARGET_ARCH"); !ok {
		targetArch = runtime.GOARCH
	}

	envCI, _ := os.LookupEnv("CI")
	ci = envCI == "true"
}

func initGodotBin() {
	var (
		err error
	)

	godotBin, _ = os.LookupEnv("GODOT_BIN")

	if godotBin, err = which(godotBin); err == nil {
		fmt.Printf("GODOT_BIN = %s\n", godotBin)
		return
	}

	if !ci {
		if godotBin, err = which("godot3"); err == nil {
			fmt.Printf("GODOT_BIN = %s\n", godotBin)
			return
		}

		if godotBin, err = which("godot"); err == nil {
			fmt.Printf("GODOT_BIN = %s\n", godotBin)
			return
		}
	}

	panic(err)
}

func envWithPlatform(platform BuildPlatform) map[string]string {
	envs := map[string]string{
		"GOOS":                   targetOS,
		"GOARCH":                 targetArch,
		"CGO_ENABLED":            "1",
	}

	return envs
}

func Clean() error {
	return nil
}

func Build() error {
	mg.Deps(initGodotBin)

	appPath := filepath.Join(".")
	outputPath := filepath.Join("libs")

	return buildGodotPlugin(
		appPath,
		outputPath,
		BuildPlatform{
			OS:   targetOS,
			Arch: targetArch,
		},
	)
}

func Run() error {
	mg.Deps(Build)

	appPath := filepath.Join(".")

	return runPlugin(appPath)
}

func runPlugin(appPath string) error {
	return sh.RunWith(
		map[string]string{
			"GOTRACEBACK": "crash",
			"GODEBUG": "asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1,tracebackancestors=3",
			"LOG_LEVEL": "debug",
			"TEST_USE_GINKGO_WRITER": "1",
		},
		godotBin, "--verbose",
		"--path", appPath)
}

func buildGodotPlugin(appPath string, outputPath string, platform BuildPlatform) error {
	return sh.RunWith(envWithPlatform(platform), mg.GoCmd(), "build",
		"-tags", "tools", "-buildmode=c-shared", "-x", "-trimpath",
		"-o", filepath.Join(outputPath, platform.godotPluginCSharedName(appPath)),
		filepath.Join(appPath, "main.go"),
	)
}

func (p BuildPlatform) godotPluginCSharedName(appPath string) string {
	// NOTE: these files needs to line up with CI as well as the naming convention
	//       expected by the test godot project
	switch(p.OS) {
		case "windows":
			return fmt.Sprintf("libgodotgo-dodge-the-creeps-windows-4.0-%s.dll", p.Arch)
		case "darwin":
			return fmt.Sprintf("libgodotgo-dodge-the-creeps-darwin-10.6-%s.dylib", p.Arch)
		case "linux":
			return fmt.Sprintf("libgodotgo-dodge-the-creeps-linux-%s.so", p.Arch)
		default:
			panic(fmt.Errorf("unsupported build platform: %s", p.OS))
	}
}

func which(filename string) (string, error) {
	if len(filename) == 0 {
		return "", fmt.Errorf("no filename specified")
	}
	return sh.Output("which", filename)
}

var Default = Build
