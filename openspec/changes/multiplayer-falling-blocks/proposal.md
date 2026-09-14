# Proposal: multiplayer-falling-blocks

## Why

The repo demonstrates godot-go with small single-player samples (`dodge_the_creeps`, `topdown`) but has no example of a real game loop with state-heavy logic, and no example of Godot's high-level multiplayer API from Go. A falling-blocks game with a multiplayer lobby fills both gaps: it is the classic "learn a game framework end to end" demo and gives users a copy-pasteable reference for hosting, joining, and syncing a competitive match.

## What Changes

- Add a new godot-go demo module at `2d/falling_blocks/` (Go module + Godot project + Makefile, following the layout of `2d/dodge_the_creeps`).
- New falling-blocks gameplay core in Go: standard 7 tetrominoes, movement/rotation with wall kicks, gravity with soft/hard drop, line clears, scoring, levels, next-piece preview, and game over (top-out).
- Playability refinements: ghost landing preview of the hard-drop position, held left/right auto-shift (initial delay, then fast repeat), held soft-drop repeat, and a solo (offline) mode reachable from the title screen.
- New multiplayer lobby scene: host or join a game over ENet (default port 7777), visible player list with names, ready-up toggle, host-controlled match start.
- New versus match mode: each player simulates their own board locally; line clears send garbage lines to opponents via multiplayer RPCs; last player standing wins, with results and rematch support.
- Register the new module in `go.work`.

## Capabilities

### New Capabilities

- `falling-blocks-gameplay`: Core single-player falling-blocks rules — piece generation, movement, rotation, drop, clearing, scoring, leveling, game over, ghost preview, held-input auto-repeat, and solo mode.
- `multiplayer-lobby`: Host/join session management, player roster, ready state, session size cap, and match start handshake over Godot high-level multiplayer.
- `versus-match`: Online competitive play — synchronized per-player boards, garbage-line attacks, win/loss determination, and rematch.

### Modified Capabilities

(none — no existing specs in `openspec/specs/`)

## Impact

- New code under `2d/falling_blocks/` only; no existing demos are modified.
- `go.work` gains one module entry.
- Depends on godot-go bindings (already vendored via `../godot-go`) and Godot's built-in `SceneMultiplayer`/`ENetMultiplayerPeer` — no new third-party dependencies.
- Networking requires LAN/reachable UDP between peers (direct host-to-client, no dedicated server component).
