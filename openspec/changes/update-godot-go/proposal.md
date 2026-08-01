## Why

Both demo projects (`2d/dodge_the_creeps` and `2d/topdown`) pin `github.com/godot-go/godot-go` at `v0.3.17`, while the latest release is `v0.3.25`. Upgrading keeps the demos compatible with the current godot-go API surface, picks up bug fixes, and ensures the examples remain a working reference for integrating godot-go.

## What Changes

- Bump `github.com/godot-go/godot-go` from `v0.3.17` to `v0.3.25` in `2d/dodge_the_creeps/go.mod` and `2d/topdown/go.mod`
- Regenerate `go.sum` for both modules and `go.work.sum` (via `go mod tidy`)
- Adapt Go sources (`main.go`, `pkg/demo/*.go`) where the godot-go API changes require it
- Verify both projects build and run (`make` and `make run`)

## Capabilities

### New Capabilities

None. This is a dependency/tooling upgrade; the demo projects' behavior does not change, so no new spec capability is introduced.

### Modified Capabilities

None. There are no existing specs (`openspec/specs/` is empty), and no spec-level behavior changes are involved.

## Impact

- **Dependencies**: `github.com/godot-go/godot-go` `v0.3.17` → `v0.3.25` in both `go.mod` files, with `go.sum`/`go.work.sum` regenerated
- **Code**: `2d/dodge_the_creeps/{main.go,pkg/demo/hud.go}` and `2d/topdown/{main.go,pkg/demo/object_player_character.go}` may need API-compatibility fixes to keep compiling
- **Workspace**: `go.work` itself is unchanged; only module sums are refreshed
