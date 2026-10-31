# Proposal

## Why

Every demo in this repository is a single-system minigame. We want a demo that proves
the architecture — pure-Go rules package under a thin Godot shell — scales to a
rules-heavy game: a small dungeon-crawl roguelite built on a Pathfinder 2e flavored
rules engine (3-action economy, d20 + proficiency vs DC checks, XP levels) with a
persistent town meta-layer.

## What Changes

- New game module `2d/dungeon_crawler/` following the established layout (`pkg/<rules>`
  pure Go, `pkg/demo` Godot bindings, `project/` Godot scene + `.gdextension`, `Makefile`),
  registered in `go.work`.
- A PF2e-lite rules engine: shared 3-action turn economy (Stride / Strike / class skill,
  multiple attack penalty −5/−10), bump combat resolved with d20 checks, ranged attacks
  (reach 3, auto-target nearest), PF2e cumulative XP level thresholds, conditions-light.
- Four classes — Fighter, Cleric, Wizard, Rogue — each contributing one combat skill and
  one world skill; world skills resolve locked doors, traps, and hidden doors as DC checks.
- A committed-run dungeon: procedurally generated floors 1–5 (BSP rooms and corridors),
  fog of war with explored memory, stairs down, potion loot, and a floor-5 boss whose
  death wins the run; death ends the run.
- A walkable town hub that is the only persisted state: renown payout at run end is
  spent bumping buildings to buy upgrades (Smithy, Clinic, Guilds ×4, Library, Archive).
- Non-goals (explicitly out): ancestries, feat trees, tactical grid positioning, gear
  carryover between runs, shops/currencies beyond renown, dungeon persistence between
  runs (dips), multiplayer.

## Capabilities

### New Capabilities

- `dungeon-crawler-combat`: the PF2e-lite turn and resolution rules — action economy,
  Stride/Strike/skill spending, multiple attack penalty, d20 proficiency checks against
  DCs and AC, class skills, ranged targeting, XP awarding and leveling.
- `dungeon-crawler-run`: everything about one attempt — dungeon generation and layout,
  fog of war, interactable features (stairs, locked doors, traps, hidden doors, potions),
  monsters and their AI turn behavior, win/loss conditions, and end-of-run renown payout.
- `dungeon-crawler-town`: the persistent hub — walkable town, bump-to-interact buildings,
  renown economy, upgrade effects read at run start, class unlock/selection, and the
  save/load boundary that is the game's only persisted state.

### Modified Capabilities

None. No existing demo's requirements change.

## Impact

- New Go module `2d/dungeon_crawler` added to `go.work`.
- New build artifact `project/lib/libgodotgo-2ddungeoncrawler-linux-amd64.so` plus a
  `.gdextension` manifest and Godot project files under `2d/dungeon_crawler/project/`.
- Reuses the tileset assets pattern from `2d/topdown`; no new external dependencies
  beyond the existing godot-go.
- Existing demos, specs, and build flows are untouched.
