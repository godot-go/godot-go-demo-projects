---
name: update-godot-go
description: Use when asked to upgrade, update, or bump the godot-go dependency (github.com/godot-go/godot-go) in the godot-go-demo-projects repo, e.g. "upgrade godot-go to the latest version".
---

# Updating godot-go

## Overview

Each demo is a separate Go module in a `go.work` workspace, pinning
`github.com/godot-go/godot-go` to a released tag. An upgrade means: bump
**every** module to the same tag, verify each demo builds and boots, then
commit with the repo's convention.

## Workflow

1. **Find the target version from the module proxy** (local git tags in
   `../godot-go` may be stale):

   ```
   go list -m -versions github.com/godot-go/godot-go | tr ' ' '\n' | tail -1
   ```

2. **Discover every module from the `use` block in `go.work`** — never
   hardcode the list; new demos get added over time.

3. **Bump each module** (run inside each module directory):

   ```
   go get github.com/godot-go/godot-go@vX.Y.Z
   ```

   `go.mod`/`go.sum` change. Indirect bumps (e.g. `zap`) ride along and are
   expected — accept them.

4. **Verify every demo**:
   - `make build` — the cgo c-shared build must exit 0.
   - Headless smoke: `LOG_LEVEL=debug godot --headless --path project/ --quit`
     — expect the extension's `...Init called` line (it is debug-level, so
     `LOG_LEVEL=debug` is required) and no panic. The gitignored
     `project/.godot/` cache must be present; without it the extension
     silently never loads.

5. **Commit** with the repo's message style: `upgrade godot-go to vX.Y.Z`
   (append migration notes if the API changed).

## If the build breaks

API migrations live in the godot-go source at `../godot-go`. Diff the tags:

```
git -C ../godot-go log --oneline vOLD..vNEW
```

Past example: v0.3.40 required qualified virtual names
`V_<Class>_<Method>` (flat `V_` panics) and a `go 1.27.0` directive bump.

## Common mistakes

- Bumping only one module — all `go.work` modules must move together.
- Trusting stale local tags — query the module proxy.
- Editing `go.mod` by hand instead of `go get` — `go.sum` drifts.
- Skipping the headless smoke test — a clean build doesn't prove the
  extension registers with the engine.
