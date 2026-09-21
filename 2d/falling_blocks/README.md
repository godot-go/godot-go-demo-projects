# Falling Blocks Multiplayer (godot-go)

A falling-blocks clone with a multiplayer lobby, written in Go via
[godot-go](https://github.com/godot-go/godot-go). Players can play solo, host
a LAN session, or join one. In a versus match every player runs their own
board locally; line clears send garbage to the opponents — last player
standing wins, and the host can rematch.

## Layout

```
2d/falling_blocks/
├── main.go            # GDExtension entry point (GodotGoDemo2DFallingBlocksInit)
├── pkg/falling_blocks/ # pure-Go game core: board, pieces, 7-bag, scoring,
│                      #   garbage — fully unit tested, no Godot dependency
├── pkg/demo/          # Godot node classes (title, lobby, match, session)
│   └── README.md      # godot-go binding findings (multiplayer on v0.3.40)
└── project/           # Godot project: Main/Lobby/Match/Session scenes,
                       #   gdextension, project settings (input map + autoload)
```

## Build & run

```sh
make build   # compiles the Go module into project/lib/*.so
make run     # runs the game (windowed)
make editor  # opens the Godot editor
```

Requires `godot` (4.2+) on `PATH`, a Go toolchain and cgo/clang, same as the
other demos in this repo.

## Controls

| Key                  | Action            |
| -------------------- | ----------------- |
| A/D or Left/Right    | Move piece        |
| W/X/Up               | Rotate clockwise  |
| Z                    | Rotate counter-cw |
| S/Down               | Soft drop (+1/cell) |
| Space                | Hard drop (+2/cell) |
| P                    | Pause (solo only) |

## Multiplayer (LAN)

1. **Host**: one player clicks *Host game* (default port 7777). The lobby
   shows the player list; share your LAN address (e.g. `192.168.1.20`).
2. **Join**: other players enter that address + port and click *Join game*.
3. Everyone presses *Ready*, then the host presses *Start match* (2–4
   players supported; a full lobby refuses new connections).
4. Clearing lines sends garbage: double = 1, triple = 2, four-line clear = 4, plus
   1 per combo. Garbage lands under your stack when you lock your next piece.
   Top out and you're out; the last player standing wins. The host can start a
   rematch from the results screen.

Sessions are direct client-to-host over ENet (Godot high-level multiplayer);
the host relays events and arbitrates eliminations. No dedicated server.

### Two-terminal LAN test (headless, no window needed)

```sh
# terminal 1 — host
LOG_LEVEL=info FALLING_BLOCKS_AUTOHOST=7777 FALLING_BLOCKS_AUTOPLAY=1 FALLING_BLOCKS_AUTODROP=1 \
  godot --headless --path project/

# terminal 2 (after ~5s) — client
LOG_LEVEL=info FALLING_BLOCKS_AUTOJOIN=127.0.0.1:7777 FALLING_BLOCKS_AUTOPLAY=1 FALLING_BLOCKS_AUTODROP=1 \
  godot --headless --path project/
```

Automation env vars (development only):

| Var                   | Effect                                            |
| --------------------- | ------------------------------------------------- |
| `FALLING_BLOCKS_AUTOHOST`     | port; auto-host on the title screen               |
| `FALLING_BLOCKS_AUTOJOIN`     | `addr:port`; auto-join on the title screen        |
| `FALLING_BLOCKS_AUTOPLAY`     | auto ready/start in the lobby                     |
| `FALLING_BLOCKS_AUTODROP`     | hard-drop every piece on each gravity tick        |
| `FALLING_BLOCKS_AUTOGARBAGE`  | also send 1 garbage row per lock (wire test)      |
| `FALLING_BLOCKS_AUTOREMATCH`  | host automatically rematches when a match ends    |

## Tests

```sh
go test ./pkg/falling_blocks/   # game core: rules, scoring, kicks, garbage
```

## Credits

`project/art/House In a Forest Loop.ogg` Copyright &copy; 2012 [HorrorPen](https://opengameart.org/users/horrorpen), [CC-BY 3.0: Attribution](http://creativecommons.org/licenses/by/3.0/). Source: https://opengameart.org/content/loop-house-in-a-forest
