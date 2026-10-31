# Dungeon Crawler (godot-go)

A small Pathfinder-2e flavored dungeon-crawl roguelite written in Go with
[godot-go](https://github.com/godot-go/godot-go). Explore 5 procedurally generated
floors under a 3-action turn economy, earn XP and renown, and spend renown on
town upgrades between runs.

## Rules

- Each turn you get 3 action points: Stride (move 1 tile), Strike (bump an
  enemy), or a class skill. Unspent actions are lost.
- Attacks are d20 + proficiency + ability vs AC; a second Strike in the same
  turn takes −5, a third −10. Natural 20 doubles damage dice, natural 1 misses.
- Four classes (Fighter, Cleric, Wizard, Rogue) each bring a combat skill and
  world skill training; checks against doors, traps and hidden doors use the
  same d20 system.
- Kills grant XP; crossing PF2e cumulative thresholds levels you mid-run.
- Death or slaying the floor-5 boss ends the run and pays renown, which is the
  only thing that persists, spent on buildings in the walkable town hub.

## Layout

```
2d/dungeon_crawler/
├── main.go             # GDExtension entry point (GodotGoDemo2DDungeonCrawlerInit)
├── pkg/crawler/        # pure-Go rules: turns, checks, dungeon gen, town, save
├── pkg/demo/           # DungeonCrawler node: input, rendering, HUD
└── project/            # Godot project: Main.tscn, input map, gdextension
```

The rules package has no Godot imports and is unit tested; the node layer only
reads input and paints engine snapshots.

## Build & run

```sh
make build   # compiles the Go module into project/lib/*.so
make run     # runs the game
make editor  # opens the Godot editor
```

Requires `godot` (4.2+) on `PATH`, a Go toolchain, and cgo/clang, same as the other
demos in this repo.

## Controls

The game opens in town. Walk into a building to open it; walk into the gate to
descend.

| Key            | Action                                  |
| -------------- | --------------------------------------- |
| WASD or arrows | Move · bump enemy = Strike · bump feature = interact |
| Space          | Class combat skill                      |
| E              | Search for hidden doors                 |
| Q              | Quaff a potion (free, no action cost)   |
| Enter          | End turn · confirm panel                |
| Esc            | Close panel                             |

## Tests

```sh
cd 2d/dungeon_crawler
go test ./pkg/crawler/
```
