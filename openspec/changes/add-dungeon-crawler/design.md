# Design

## Context

The repository builds each demo as a Go module exposing a godot-go extension, with a
pure-Go rules package under a thin Godot shell (`2d/snake/pkg/snake` + `2d/snake/pkg/demo`
is the reference). See proposal.md for motivation. Constraints that shape this design:

- The shell talks to Godot only through godot-go registered classes; rules code must
  stay Godot-free so it can be unit-tested headless, as `snake` does with `game_test.go`.
- Build produces one `c-shared` library per module (`project/lib/libgodotgo-<name>-<os>-<arch>.so`),
  loaded via a `.gdextension` manifest; `Makefile` targets are uniform across demos.
- `2d/topdown` already ships a usable tileset resource (`world-tileset.tres`) and sprite
  sheet pattern; `2d/snake` shows immediate-mode `_draw` for simple games.

## Goals / Non-Goals

**Goals:**
- A deterministic, seeded rules engine covering all three spec capabilities, fully
  testable without Godot.
- A shell that only renders, reads input, and forwards verbs — no rules in the shell.
- One JSON save file as the entire persistence surface.

**Non-Goals (design-level):**
- No per-tile Godot nodes; the map is TileMap data, not a scene graph.
- No networked or multiplayer state; single local player.
- No generic "game engine" abstractions beyond this game's needs — rules live in one
  cohesive package, not a framework.

## Decisions

### One rules package, `pkg/crawler`, Godot-free
Single Go package holding combat, dungeon, run, town, and meta logic in files
(`combat.go`, `dungeon.go`, `run.go`, `town.go`, `bestiary.go`, `numbers.go`,
`save.go`, plus `_test.go` siblings). Alternative — one package per subsystem — was
rejected: combat and dungeon share grid/actor types, sub-packages would fight import
cycles for no testability gain at this size, and `snake` precedent is one package.

### Verb-in, events-out engine API
The shell calls engine verbs (`Stride(dir)`, `Strike()`, `UseSkill(id)`, `Search()`,
`Descend()`, `EndTurn()`, …); each returns a list of `Event` values (log lines with
structured payloads: roll, damage, reveal, level-up, payout). The shell renders events
into the combat log and repaints from the engine snapshot. Enemies act inside the verb
call that ended the player turn. Alternatives — engine calling Godot callbacks or
observers — rejected: they re-couple rules to Godot and complicate headless tests.

### Seeded randomness injected everywhere
`crawler.NewRun(seed int64)` threads one `math/rand.Rand` through generation and every
check, so any failure reproduces from the seed printed on the end screen. `snake` uses
package-global `rand`; rejected here because a rules engine with combat odds and
procedural maps is untestable at the edges without reproducibility.

### Turn-based shell: repaint on snapshot change, no `_process` game logic
The Godot node repaints after each verb and on visibility-affecting input; a `_process`
tick exists only for cursor/UI blink. This makes the frame loop irrelevant to rules
(speed cannot drift because there is no speed) and mirrors how turn-based roguelikes
shell cleanly over pure state machines. Alternative real-time driver rejected by genre.

### Immediate-mode `_draw` with colored rects; three fog tints (revised during implementation)
The original decision — shipped `world-tileset.tres` + `TileMapLayer` cells — was
implemented against, then reversed: an authored tileset requires editor-created atlas
assets that cannot be validated headless, and godot-go's `Color` is a raw `[16]uint8`
blob with no float accessors, so every color had to be owned Go-side anyway. The shell
instead paints the current mode in `_draw` with `draw_colored_polygon` rects (the snake
demo's precedent): terrain, features, and entities each one quad, fog expressed as
bright / dimmed-explored / skipped-unseen per tile. Same observable states the spec
requires; zero asset dependencies.

### Bresenham per-tile LOS, radius 6 (revised during implementation)
Symmetric shadowcast was implemented against and replaced with a permissive Bresenham
raycast to each tile within radius 6. The corner-peek concern that drove the original
choice is moot for this game: hidden doors render as plain wall until a successful
Search reveals them, so line-of-sight cannot leak secret positions, and rooms are wide
enough that peek artifacts read as forgiving visibility rather than cheating. Bresenham
is trivially pure grid math in `crawler`.

### BSP room generation with typed room contents
Split-recursive BSP into ~7–9 rooms per floor, L-corridor connections, then a content
pass assigns room types (entrance, stairs, encounter, treasure-with-locked-door, trap
hall, shrine) from a per-floor table. Alternative — cellular automata — rejected: caves
fight doors, corridors, and locked-room placement, which the world skills need.

### Save as one JSON owned by `crawler`, path supplied by shell
`crawler.LoadMeta()/SaveMeta()` read/write `{renown, building levels, first-kills,
unlocked classes, settings}` at `user://dungeon_crawler_save.json`; the demo layer
resolves the `user://` path through Godot `ProjectSettings` and passes a filesystem
path. Writes happen on renown gain and purchase. Alternative — GDScript `ConfigFile` —
rejected: meta logic (costs, caps, unlock gating) is spec'd behavior and must be
testable headless.

### Numbers in one table file, PF2e-derived
`numbers.go` holds PF2e cumulative XP thresholds (20/40/60/80 for levels 2–5),
proficiency = level + 2 (trained), monster statline generator by level (AC, HP, attack
bonus, damage die, XP), class ability arrays and weapon dice, DC tier table, renown
payout weights. Tuning happens here; tests pin invariants (e.g. clearing a floor-1
encounter suite reaches level 2 before floor 3) rather than exact balance.

### Single scene, two modes, Go-built panels
`Main.tscn`'s root is the registered `DungeonCrawler` node managing town mode, run
mode, and an end screen (state switch on the root, reusing the same `_draw` renderer).
Building panels and class select are `Label` nodes in the HUD layer, retargeted by text
(the snake HUD pattern scaled up). Alternative — separate scenes switched by
`change_scene` or a GDScript UI — rejected: scene switching across a godot-go boundary
adds churn for two screens, and the repo keeps logic in Go.

## Risks / Trade-offs

- [godot-go TileMap/Control API verbosity makes the shell bloated] → Spike the shell's
  map repaint first, before writing town content; keep shell to repaint + input only.
- [XP/renown curves unbalanced (grind or breeze)] → All curves in `numbers.go`; invariant
  tests on pacing; tune after first full-playable milestone, before content polish.
- [PF2e fidelity creep (traits, reactions, conditions) expands scope] → Proposal
  non-goals bind; anything not specced stays out of this change.
- [Fog leaks entity positions (memory shows stale monsters)] → Engine snapshot emits
  entities only for currently-visible tiles; explored memory carries terrain only.
- [Turn auto-end surprises players mid-Skill] → HUD shows 3 action pips prominently;
  log lines name each consumed action.
- [BSP rooms too homogeneous] → Room typing + feature tables are data; regenerate
  variety is a content-only fix.

## Migration Plan

Additive only: new module `2d/dungeon_crawler`, new `go.work` entry. Rollback is removing
the module directory and its `go.work` line. No existing demo, spec, or CI step changes.

## Open Questions

- Final bestiary species list and names (PF2e-authentic leaning) — table content, does
  not affect specs or structure.
- Exact ability arrays per class and monster generator constants — tuned during
  implementation; pinned by invariant tests, not by spec.
