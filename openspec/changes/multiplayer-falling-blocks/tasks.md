# Tasks: multiplayer-falling-blocks

## 1. Module scaffold

- [x] 1.1 Create `2d/falling_blocks/` Go module (`go.mod` as `godot-go-demo-projects/2d/falling_blocks`, go 1.27) and add it to `go.work`; verify `go build ./...` passes empty
- [x] 1.2 Create `project/` Godot project: `project.godot`, `godotgo.gdextension`, `Makefile` (copy dodge_the_creeps patterns, library name `libgodotgo-2dfalling-blocks`), and `main.go` exporting `GodotGoDemo2DFallingBlocksInit`
- [x] 1.3 Verify the c-shared build produces `project/lib/libgodotgo-2dfalling-blocks-linux-amd64.so` and `make run` opens an empty `Main.tscn` title scene in Godot

## 2. Pure-Go falling-blocks core (pkg/falling_blocks)

- [x] 2.1 Implement `Board` (10x20+buffer, cell grid) with collision, spawn placement, and `LockPiece`; unit-test top-out detection (game-over spec)
- [x] 2.2 Implement 7 tetromino definitions, orientations, and seeded 7-bag `Randomizer`; unit-test bag completeness and preview of next piece
- [x] 2.3 Implement move/rotate with wall-kick offsets; unit-test wall rejection, kick success, and rotation-into-occupied rejection
- [x] 2.4 Implement gravity tick, soft drop, and hard drop with per-cell scoring (1/2 pts) and lock-on-drop; unit-test each
- [x] 2.5 Implement line clear, stack shift, scoring table (100/300/500/800 x level) and level-up every 10 lines with speed curve; unit-test double-clear-at-level-2 and partial-row cases
- [x] 2.6 Implement garbage queue (`AddGarbage(n)`, injected below stack at next lock with one random gap, top-out possible); unit-test apply-on-lock and garbage-induced top-out

## 3. Single-player presentation

- [x] 3.1 Build `Match.tscn` UI skeleton: playfield `BoardView` Node2D drawing a `Board` via `_draw`, side panel with score/level/lines/next-piece preview
- [x] 3.2 Wire Go input handling (`_UnhandledInput` or `_Input`) for move/rotate/soft/hard drop and pause toggle; gravity driven from `_Process` interval per level
- [x] 3.3 Implement local pause overlay (freeze gravity/input) and game-over overlay with final score; verify solo playthrough reaches top-out

## 4. Networking layer

- [x] 4.1 Spike: confirm godot-go bindings for `SceneMultiplayer`/`MultiplayerAPI`, `ENetMultiplayerPeer` create server/client, and Go-method RPC registration work at runtime; document gaps and the `Call`-by-name fallback in `pkg/demo/net` README comments
- [x] 4.2 Implement `Session` autoload: host(port)/join(address,port) via `ENetMultiplayerPeer`, peer connect/disconnect signals, player name, roster snapshot state, and error paths (host port busy, join refused)
- [x] 4.3 Implement host relay helpers: client→host RPC send, host→all broadcast of roster/garbagematch events with sender-unknown-peer filtering

## 5. Lobby (multiplayer-lobby spec)

- [x] 5.1 Build `Main.tscn` title screen: name entry, Host button with port field (default 7777), Join button with address+port fields (default localhost:7777), error labels
- [x] 5.2 Build `Lobby.tscn`: roster list (name, HOST tag, ready state) driven by `Session` snapshots; host view shows address/port for sharing
- [x] 5.3 Implement ready toggle for clients, host always ready; roster updates within 1s on join/leave (verify with two local instances)
- [x] 5.4 Implement host start control disabled until >=2 players and all ready; start RPC makes every client `ChangeSceneFileSafe` into `Match.tscn`
- [x] 5.5 Implement leave/disconnect returning to title and roster cleanup on remaining clients

## 6. Versus match (versus-match spec)

- [x] 6.1 Implement synchronized match start: host sends start RPC with seed-per-peer; all clients reset board/score/garbage and begin on signal
- [x] 6.2 Implement garbage send/receive: on lock+clear compute attack (0/1/2/4 + combo) from spec table, RPC to host, host relays to opponents, recipient `AddGarbage` queue; combo counter reset rules
- [x] 6.3 Implement score/lines/garbage-count status RPCs piggybacked on clear/lock events; opponent panel per player (name, score, lines, pending, alive) refreshing within 1s
- [x] 6.4 Implement elimination + win: client reports top-out to host; host marks eliminated, ends match when <=1 alive, broadcasts winner; results overlay on all clients, boards stop accepting input
- [x] 6.5 Implement rematch (host action from results → full reset + synchronized start) and disconnect-mid-match elimination/early-end rules
- [x] 6.6 Two-instance LAN test (host + client) played through a full match including garbage exchange, winner, rematch; three-instance smoke check

## 7. Polish & docs

- [x] 7.1 Add demo `README.md` (screenshots optional) with run instructions incl. `make build && make run` and two-terminal LAN testing; cap session at 4 players host-side
- [x] 7.2 Run `go vet ./...`, `go test ./...` in `2d/falling_blocks`, and `openspec validate multiplayer-falling-blocks --strict` clean

## 8. Post-implementation additions

- [x] 8.1 Implement `Game.GhostCells` landing projection and render it as a dim shadow in `Match`; unit-test floor, on-stack, and resting cases
- [x] 8.2 Implement pure-Go `ShiftRepeater` (170ms DAS, 60ms repeat) wired into match input plus 40ms soft-drop hold repeat; unit-test tap, hold, direction-change, and release-reset
- [x] 8.3 Add solo mode: title `Solo` button → `Session.Solo` synthetic one-entry roster → offline match with Play Again; gate pause to solo matches
- [x] 8.4 Enforce the 4-player host-side cap: refuse new connections once the lobby is full, with a visible error on the joining client
- [x] 8.5 Add `FALLING_BLOCKS_AUTOHOST/AUTOJOIN/AUTOPLAY/AUTODROP/AUTOGARBAGE/AUTOREMATCH` dev hooks and document headless multi-terminal LAN testing in the demo README
- [x] 8.6 Apply the `1.0 × 0.8^(level−1)` gravity curve (0.05s floor) and re-arm the gravity timer on level-up
- [x] 8.7 End the game when a locked cell occupies a hidden buffer row (stack top-out), not only on blocked spawn; unit-test garbage-induced top-out
