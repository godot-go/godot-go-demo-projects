# Proposal

## Why

The repo has three demos but no small, complete, single-player game loop. Snake is
the smallest one that is still a real game: one grid, one pellet, one timer, and a
win condition reachable in a minute. It gives godot-go users a compact reference
for scene-driven ticks, action input, and custom 2D drawing without the moving
parts of the existing demos.

## What Changes

- Add a self-contained demo module at `2d/snake`, registered in `go.work`.
- Add a pure-Go rules package (`pkg/snake`) covering tick movement, turning, pellet
  consumption and growth, wall/self collision, and the win threshold — with unit
  tests and no Godot imports.
- Add a Godot node layer (`pkg/demo`) that drives one tick per timer timeout and
  paints the board, snake, and pellet with colored polygons.
- Add the Godot project: 20x15 board at 32 px (640x480 window), input map for
  WASD/arrows, restart, and pause, plus a Main scene with the step timer and HUD.
- Document build, run, and controls in `2d/snake/README.md` and mention the demo in
  the root README.

Game rules: the snake starts at length 3 and grows one segment per pellet; the run
is won at 20 pellets; touching a wall or its own body loses. No wrapping, no
reverse turns, no assets, no audio, no multiplayer.

## Capabilities

### New Capabilities
- `snake-gameplay`: the rules and observable behavior of the snake run — grid
  movement on a fixed tick, growth on pellet consumption, the 20-pellet win, wall
  and self collision loss, and the restart path.

### Modified Capabilities

<!-- None: the change adds a new demo and does not alter existing capabilities. -->

## Impact

- New code only: `2d/snake/` (module, rules package, node layer, project files).
- `go.work` gains a `use` entry, so the module joins workspace-wide tooling and the
  `update-godot-go` upgrade sweep.
- Root `README.md` gains a mention; no existing demo, spec, or dependency changes.
