# Design

## Context

See `proposal.md` for motivation and `specs/snake-gameplay/spec.md` for the behavior
contract.

The repo's demos establish the constraints this design follows:

- `falling_blocks` splits rules (`pkg/falling_blocks`, unit tested, no Godot imports)
  from the nodes that bind and draw (`pkg/demo`). That split is what makes the rules
  testable outside a running engine.
- `falling_blocks` drives gameplay from a `Timer` node in the scene, and paints cells
  with `DrawColoredPolygon`. `draw_rect` mis-marshals its filled/width arguments on
  this godot-go / Godot 4.8 combination, producing a warning per call and stalling the
  main thread.
- Each demo is its own module with a `Makefile` producing
  `project/lib/libgodotgo-<name>-<os>-<arch>.so`, plus an `Init` export in `main.go`.
- `go.work` lists the modules; `update-godot-go` discovers modules from that list, so a
  missed entry would silently drop the demo from dependency upgrades.

The engine here is Godot 4.8 with godot-go v0.3.65.

## Goals / Non-Goals

**Goals:**

- Keep every rule decision in a pure package that tests can call directly.
- One tick = one cell of movement, with speed changes expressed only as timer period.
- Zero imported assets; the board, snake, pellet, and text are built from engine
  primitives and scene labels.
- A new module that behaves exactly like the existing three under `make build`, the
  headless smoke test, and the upgrade sweep.

**Non-Goals:**

- Smooth interpolation between cells, sprite art, sound, or juice/particles.
- Menus, persisted scores, or multiple boards.
- Multiplayer or networking of any kind.

## Decisions

### D1: Pure rules package, thin node layer

`pkg/snake` owns the state machine: board dimensions, body cells, direction, pending
direction, pellet position, consumed count, and `Playing/Won/Lost`. `pkg/demo` owns
one class, `SnakeGame`, which reads input, calls `Tick()`, and draws.

Alternatives: putting rules in the node (cannot unit test without loading Godot),
or a generic ECS (far more machinery than a 300-cell board needs).

### D2: Body as an ordered slice plus an occupancy grid

The body is head-first `[]Cell`, and a `[]bool` grid of `W*H` mirrors occupancy for
O(1) collision checks. A normal tick pops the tail and pushes the new head; eating a
pellet skips the pop, which is the entire growth rule. Placement of a new pellet
collects free cells and picks one with an injected `*rand.Rand`, so tests are
deterministic.

Alternatives: a map-based occupancy set (allocation churn on every tick), or scanning
the body slice per collision (fine at length 23, but the grid is simpler to reason
about and removes any tail-cell ambiguity).

### D3: Ticks come from a scene `Timer`

A `Timer` node named `Step` emits `timeout`; the node handler calls `Tick()`. Pause
stops the timer; speed changes set `wait_time` and restart it. No elapsed-time
accumulation in `_process`, so a stalled frame cannot cause a burst of catch-up moves.
The start gate reuses this: while the run waits for its first input the timer stays
stopped, so the same `timeout` handler needs no special case.

Alternative: accumulate delta in `_process` and step when it crosses a threshold. That
is more code, drifts with frame timing, and makes pause a boolean check instead of a
timer call.

### D4: Poll input in `_process`, latch one turn

Each frame reads the four turn actions and the pause/restart actions. The first turn
requested since the last tick is stored as `pending`; later requests before the tick
are dropped. A pending direction opposing the current one is discarded at tick time.

While the run is waiting to start, `_process` polls `Input.is_anything_pressed()` rather
than the action map, so keys outside the bindings and mouse buttons all start the run.
A `waitForRelease` flag is armed whenever the node enters the waiting state; the held
Enter/R that triggered a restart therefore cannot also start the new run. It clears once
nothing is pressed, and the next press starts play.

Alternatives: `_input` with event objects (the topdown demo documents a type-wrapping
bug in that path), or a queue of turns (allows double-turns inside one tick, which the
spec forbids).

### D5: Draw cells as inset polygons; text as scene labels

`_draw` fills the board background, then each body cell and the pellet as an inset
32x32 rectangle via `DrawColoredPolygon`, leaving a 1 px gap that reads as a grid.
HUD text and the win/lost/paused messages are `Label` nodes in `Main.tscn`; the node
layer only sets their text and visibility.

Alternative: drawing text from Go. Label nodes handle font, centering, and wrapping
with no extra code.

### D6: Layout — 20x15 at 32 px under a 80 px HUD strip

The window is 640x560: an 80 px HUD strip at the top, and the 640x480 playfield below
it at `(0, 80)`. The board is exactly the window width, so wall collisions are visible
at the edge.

Alternative: overlay the HUD on the playfield (covers cells the snake can enter) or
shrink the board to 18x13 to fit a 640x480 window (an arbitrary-looking board).

### D7: Speed ramp is a timer period change only

The run starts at 6 cells/second and shortens the period by one step per 5 pellets to
a floor of 12 cells/second. Because the rules package is tick-driven, the ramp never
touches rules code — it only changes `wait_time`.

### D8: Module scaffold copied from `topdown`

`2d/snake` gets its own `go.mod` (`godot-go-demo-projects/2d/snake`), `main.go` with
`//export GodotGoDemo2DSnakeInit`, and a Makefile producing
`libgodotgo-2dsnake-<os>-<arch>.so`. The library name must match the
`project/gdextension` entry or the extension never loads. The module depends only on
godot-go v0.3.65; no logging dependency is needed.

### D9: Collision is strict, including the vacating tail

The spec makes this explicit: the head entering any currently occupied cell loses,
even if that segment is the tail that would move away this tick. This matches the
pop-then-push order being *skipped on growth only*, and removes a special case that
would otherwise need its own test and explanation.

### D10: The start gate lives in the node layer

`pkg/snake` already treats pause as "ticks are ignored", but the start gate decides
when to start the timer, which prompt to show, and how the first press maps to a turn.
Those are presentation concerns, so `SnakeGame` owns a `started` flag: `_ready` leaves
`Step` stopped with the prompt up, the first anything-pressed frame starts the timer,
hides the prompt, and forwards a direction action as the first turn. Restart returns to
the same waiting state. The rules package and its tests are unchanged.

Alternative: teach the rules package about "not started yet". That adds a state with no
rule consequences — the snake still must not move — and a transition only the renderer
cares about.

## Risks / Trade-offs

- [godot-go marshalling quirks on Godot 4.8] -> Use only APIs the existing demos
  already exercise: `DrawColoredPolygon`, timer nodes, `Input` polling, scene labels.
- [Changing `wait_time` mid-run can restart the current interval] -> Apply the ramp
  only on pellet consumption, where a slightly early tick is invisible.
- [Polling input can miss two taps between ticks] -> The spec asks for exactly one
  applied turn per tick, so dropping the second tap is correct, not a bug.
- [A held key after restart starts the new run instantly] -> `waitForRelease` holds the
  gate until the key that triggered the restart is released.
- [A missed `go.work` entry hides the demo from upgrades and CI builds] -> Add it in
  the same commit as the module and verify with `go build` from the workspace root.
- [Extension silently not loading] -> The headless smoke test requires the gitignored
  `project/.godot/` cache and `LOG_LEVEL=debug`; the `SnakeGameInit called` line is the
  proof.

## Migration Plan

New demo only — nothing to migrate or roll back in existing code. Reverting means
deleting `2d/snake` and its `go.work` line.
