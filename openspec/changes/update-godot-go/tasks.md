## 1. Bump the dependency in both modules

- [x] 1.1 Run `go get github.com/godot-go/godot-go@v0.3.25` in `2d/dodge_the_creeps`
- [x] 1.2 Run `go get github.com/godot-go/godot-go@v0.3.25` in `2d/topdown`
- [x] 1.3 Run `go mod tidy` in `2d/dodge_the_creeps`
- [x] 1.4 Run `go mod tidy` in `2d/topdown`
- [x] 1.5 Run `go mod tidy` at the workspace root to refresh `go.work.sum` (via `go work sync`; no go.mod at root)
- [x] 1.6 Confirm both `go.mod` files require `github.com/godot-go/godot-go v0.3.25`

## 2. Fix API compatibility issues

- [x] 2.1 Build `2d/dodge_the_creeps` (`make build` / `go build`) and fix any godot-go API compile errors in `main.go` and `pkg/demo/hud.go`
- [x] 2.2 Build `2d/topdown` (`make build` / `go build`) and fix any godot-go API compile errors in `main.go` and `pkg/demo/object_player_character.go`

## 3. Verify

- [x] 3.1 Run `go build ./...` from the workspace root with no errors (run per module; no go.mod at workspace root)
- [x] 3.2 Run `make` in `2d/dodge_the_creeps` and confirm the shared library builds
- [x] 3.3 Run `make` in `2d/topdown` and confirm the shared library builds
- [x] 3.4 If Godot is available, run `make run` in each project to confirm the demos load (verified: both demos run with v0.3.25; `.godot` imports were generated via `make ci_gen_project_files` first)
