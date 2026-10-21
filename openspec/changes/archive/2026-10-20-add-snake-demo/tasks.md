# Tasks

## 1. Module and workspace scaffold

- [x] 1.1 Create `2d/snake` with `go.mod` (`module godot-go-demo-projects/2d/snake`, `go 1.27.0`, godot-go v0.3.65) and a `Makefile` copied from `topdown` with output `project/lib/libgodotgo-2dsnake-$(GOOS)-$(GOARCH).so`; verify `cd 2d/snake && go mod tidy && go list -m` reports the module path
- [x] 1.2 Add `./2d/snake` to the `use` block in `go.work`; verify `go work edit -json` lists it and `cd 2d/snake && go build ./...` exits 0

## 2. Rules package (`pkg/snake`)

- [x] 2.1 Define `Cell`, direction, run-state, and `Game` types with `NewGame`, `Turn`, `Tick`, `Pause`, and `Restart`; verify `go build ./pkg/snake` exits 0
- [x] 2.2 Implement tick movement, one-turn latching, and reversal rejection; verify table tests cover advance-one-cell, no movement without a tick, a queued turn applying next tick, ignored reversal, and a dropped second turn in the same tick (`go test ./pkg/snake/`)
- [x] 2.3 Implement pellet consumption, growth (tail retained), counter increment, and free-cell respawn with an injected `*rand.Rand`; verify tests cover length growth, counter advance, and a spawned pellet never overlapping the body
- [x] 2.4 Implement loss on wall and self contact (including the vacating tail cell), the win at 20 pellets, and reset semantics for `Restart` / `Pause`; verify tests cover all four walls, body contact, the 20th pellet, no movement while won or paused, and restart returning length 3 headed right at zero pellets

## 3. Godot node layer and project

- [x] 3.1 Add `pkg/demo` with a `SnakeGame` class registering `_ready`, `_process`, and `_draw` plus `GetClassName`/`GetParentClassName`; verify `make build` produces the shared library and `go vet ./...` passes
- [x] 3.2 Add `project/project.godot` (640x560 viewport, input actions for up/down/left/right using WASD and arrows, plus `restart` and `pause`) and `project/Main.tscn` with the `SnakeGame` root, a `Step` timer, and HUD labels; verify the project imports and runs with `godot --headless --path project/ --editor --quit`
- [x] 3.3 Add `project/gdextension` pointing at `res://lib/libgodotgo-2dsnake-<os>-<arch>.so`; verify `LOG_LEVEL=debug godot --headless --path project/ --quit` logs `SnakeGameInit called` with no panic
- [x] 3.4 Wire ticking to the `Step` timer, speed ramp on pellet milestones, pause, and restart; verify a headless run reaches the scene and a manual run shows movement stopping on pause and resetting on restart
- [x] 3.5 Implement `_draw` for the board, body, and pellet via `DrawColoredPolygon`, and HUD updates for progress plus the won/lost messages; verify a manual run shows the 20x15 grid, growth on each pellet, and the distinct win and loss overlays
- [x] 3.6 Add `main.go` exporting `GodotGoDemo2DSnakeInit` and registering the class; verify `make build` and the headless smoke line above both pass

## 4. Documentation and integration

- [x] 4.1 Write `2d/snake/README.md` with layout, build/run commands, controls, and the win goal; verify each documented command runs as written
- [x] 4.2 Mention the snake demo in the root `README.md`; verify the linked path resolves
- [x] 4.3 Run `make build` in all four demos plus the snake headless smoke test; verify every build exits 0 and the smoke log contains `SnakeGameInit called` with no panic

## 5. Start gate

- [x] 5.1 Keep `Step` stopped in `_ready`, draw the board, and show a "press any key to start" prompt; verify a headless boot leaves the timer stopped with the prompt visible
- [x] 5.2 Start the run on the first key or mouse button press, queueing a direction action as the first turn, and ignore held keys after entering the waiting state; verify with a headless harness
- [x] 5.3 Return to the waiting state when restarting after a win or loss; verify restart leaves the timer stopped with the prompt visible
- [x] 5.4 Update `2d/snake/README.md` for the start prompt and re-run rules tests, `make build`, the headless smoke test, and a captured-frame check

## Workflow follow-up

- Archive the change with `openspec archive add-snake-demo` once the demo is verified and any review is complete.
