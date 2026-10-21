# Snake (godot-go)

Classic snake written in Go with [godot-go](https://github.com/godot-go/godot-go).
Steer a three-segment snake around a 20x15 board, eat pellets to grow, and win by
eating 20 of them.

## Rules

- Each pellet eaten grows the snake by one segment and counts toward the goal.
- Hitting a wall or any part of your own body ends the run.
- Reach 20 pellets to win.
- The snake starts at 6 cells per second and gains a step every 5 pellets, up to
  12 cells per second.

## Layout

```
2d/snake/
├── main.go        # GDExtension entry point (GodotGoDemo2DSnakeInit)
├── pkg/snake/     # pure-Go rules: ticks, turning, growth, collision, win
├── pkg/demo/      # SnakeGame node: timer wiring, input, drawing
└── project/       # Godot project: Main.tscn, input map, gdextension
```

The rules package has no Godot imports and is unit tested; the node layer only
drives ticks and paints the board.

## Build & run

```sh
make build   # compiles the Go module into project/lib/*.so
make run     # runs the game
make editor  # opens the Godot editor
```

Requires `godot` (4.2+) on `PATH`, a Go toolchain, and cgo/clang, same as the other
demos in this repo.

The first headless editor import on a fresh checkout exits non-zero while the
gitignored `project/.godot/` cache is created; run the command again.

## Controls

The game opens with the board drawn and a start prompt. Press any key or mouse
button to begin, and again after a restart.

| Key              | Action                    |
| ---------------- | ------------------------- |
| Any key / click  | Start a waiting run       |
| WASD or arrows   | Steer                     |
| P                | Pause / resume            |
| Enter or R       | Restart after win or loss |

## Tests

```sh
cd 2d/snake
go test ./pkg/snake/
```
