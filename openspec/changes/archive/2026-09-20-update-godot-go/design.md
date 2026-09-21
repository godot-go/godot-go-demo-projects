## Context

Both demo modules (`2d/dodge_the_creeps`, `2d/topdown`) depend on `github.com/godot-go/godot-go v0.3.17` and are part of a single Go workspace (`go.work`). Each module builds a `c-shared` library via `make` (`-buildmode=c-shared -tags tools`) that Godot loads through a `.gdextension` file. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**
- Bring both modules onto `github.com/godot-go/godot-go v0.3.25`
- Keep both projects compiling and runnable (`make && make run`)
- Regenerate all module sums consistently

**Non-Goals:**
- No changes to demo behavior, scenes, or `.gdextension` configs
- No upgrade of the base `go 1.21.4` directive or other dependencies unless a new godot-go version requires it
- No godot-go API feature adoption — only compatibility fixes

## Decisions

1. **Upgrade via `go get`, then `go mod tidy`**
   Run `go get github.com/godot-go/godot-go@v0.3.25` in each module directory, then `go mod tidy` in each module and at the workspace root. Rationale: keeps `go.sum` and `go.work.sum` authoritative and consistent. Alternative (hand-editing `go.mod`) is error-prone and leaves sums stale.

2. **Compile-error-driven code adaptation**
   After bumping, build each project and fix only what fails to compile (e.g., renames, signature changes, moved packages in godot-go between `v0.3.17` and `v0.3.25`). Rationale: minimizes churn in the demo sources. Alternative (rewriting to new idioms proactively) risks touching working code for no functional gain.

3. **Verify in the workspace, not per-module in isolation**
   Run `go build ./...` from the workspace root and `make` per project. Rationale: the shared `go.work` means sums and module graph must stay consistent across both modules.

## Risks / Trade-offs

- godot-go `v0.3.17` → `v0.3.25` may introduce API breaks (renamed or moved symbols used by `main.go` / `pkg/demo`) → Mitigation: resolve compile errors incrementally, referencing `../godot-go` source when needed
- New transitive dependencies or versions could be pulled in → Mitigation: review `go.mod`/`go.sum` diff after `go mod tidy`; keep changes minimal
- One project may build while the other fails → Mitigation: build and fix both, verify `go.work.sum` is regenerated at the root
- Godot may not be installed in this environment, blocking runtime verification → Mitigation: rely on compile-time verification (`go build`, `make build`) and note that `make run` verification requires a local Godot binary
