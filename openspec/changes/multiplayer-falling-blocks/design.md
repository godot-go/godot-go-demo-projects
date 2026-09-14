## Context

See `proposal.md - Why`. The demo must follow the existing godot-go demo layout (`2d/dodge_the_creeps`: exported `...Init` c-shared entry point, `pkg/` Go classes registered via `core.NewInitObject`, a `project/` Godot folder with `.gdextension`, and a Makefile producing `project/lib/libgodotgo-2dfalling-blocks-*`). godot-go already exposes `ENetMultiplayerPeer`, `MultiplayerAPI` (`SetMultiplayerPeer`, `GetMultiplayerPeer`, `GetPeers`), and `Node.Multiplayer()`, so networking uses Godot's built-in high-level multiplayer rather than a custom socket layer.

## Goals / Non-Goals

**Goals:**
- Network-free core game logic that is unit-testable with plain `go test` (no Godot runtime).
- Direct client-to-client session over LAN (host is one of the players); no dedicated server, no relay, no NAT traversal.
- Small, RPC-light wire protocol: events, not board replication.

**Non-Goals:**
- Online accounts, matchmaking servers, STUN/relay, internet play through NAT.
- Ranked mechanics, SRS fairness rulesouts beyond basic kicks, hold/swap piece, T-spin detection.
- Spectator clients joining mid-match.

## Decisions

### Pure-Go game core in `pkg/falling_blocks` (board, pieces, scoring) separate from Godot nodes
The playfield is a plain Go struct (`Board`, `Piece`, 7-bag `Randomizer`, `ClearAndScore`) with an injected `*math/rand.Rand`. A thin `BoardView` Node2D in `pkg/demo` renders a `Board` and forwards input. Alternative considered: put logic directly in the node's `_Process` (Godot-binding-heavy, untestable without a running engine). This split is what lets gameplay specs be validated by Go tests.

### Deterministic visuals, event-based networking (no lockstep, no frame sync)
Each player simulates only their own board; the network carries *effects* (garbage queued, score/lines deltas, eliminated) — never inputs or full boards. Alternative considered: full state replication from a host authority — simpler conceptually but doubles traffic, adds latency to player input, and undermines the responsiveness of a falling-block game. Divergence risk is nil because boards are private by design.

### Host as relay + arbiter
All clients connect only to the host via `ENetMultiplayerPeer.CreateServer/CreateClient`. Client RPCs are sent with `Multiplayer.Rpc` to the host, which relays broadcast events (garbage targeting, roster, match end/rematch) to everyone, since clients hold no direct connection to each other. The host additionally decides elimination ordering and winner (receiving `GameOver` reports). Alternative: peer-to-peer mesh — more connection management for no benefit in a 2-4 player demo.

### Lobby, match, and results as separate scenes with `autoload` session singleton
`Main.tscn` (title) → `Lobby.tscn` → `Match.tscn` → results overlay inside Match. An autoloaded `Session` Go node owns the current `MultiplayerPeer` config and roster across scene changes (Godot's `MultiplayerAPI` survives free/reload of the root when re-set, so `session.reload_blocks` is not relied on; instead the peer is created per scene and `Session` stores only the address/port/names). Alternative: single-scene UI switching — more node bookkeeping, less idiomatic for Godot demos.

### Roster via host-authoritative `Players` dict replicated on change
Host keeps `{peer_id: {name, ready, alive, score, lines, garbage}}` and pushes the whole table with each change event (rare, tiny payload) rather than per-field RPCs; clients render from the snapshot. Avoids lost-update races at trivial bandwidth.

### Garbage queue applied at lock time
Incoming garbage increments `Game.Pending` on the recipient's local game and is injected via `Board.AddGarbageLines` after the piece locks but before line-clear resolution, matching versus-mode convention and making the "applies on next lock" spec scenario a unit test.

### Level-scaled gravity curve
`GravityInterval(level) = 1.0 × 0.8^(level−1)` seconds with a 0.05s floor (classic-flavored, resolving the former open question); the gravity timer is re-armed with the new interval whenever the level increases.

### Ghost preview and auto-shift as pure-Go helpers
`Game.GhostCells` projects the hard-drop landing position, and a `ShiftRepeater` (170ms initial delay, 60ms repeat; soft-drop repeats every 40ms) decides per-frame horizontal moves — both plain Go in `pkg/falling_blocks`, driven by the frame loop in `pkg/demo`, so landing geometry and input timing stay unit-testable without Godot.

### Solo mode as a synthetic one-entry roster
`Session.Solo` builds a solo-flagged roster with a single alive player and enters `Match.tscn` unchanged; pause is gated to solo matches (online play never pauses) so versus timing stays honest.

### Dev-automation env vars for headless LAN tests
`FALLING_BLOCKS_AUTOHOST`, `_AUTOJOIN`, `_AUTOPLAY`, `_AUTODROP`, `_AUTOGARBAGE`, and `_AUTOREMATCH` drive the title/lobby/match screens from the environment for scripted multi-process testing; they are read once at startup, are unused in normal play, and are documented in the demo README.

## Risks / Trade-offs

- [godot-go RPC binding gaps] → If a needed `MultiplayerAPI`/`SceneMultiplayer` method is missing from generated bindings, fall back to `Node.Call`/`CallDeferred` by name on the multiplayer object, or extend via `godot.InternalGetClassMethod`; validate early (Task 4 spike).
- [Headless CI cannot run ENet scenes] → Keep networking code thin and funnel all rules through `pkg/falling_blocks`; unit-test rules, and cover networking with scripted multi-process headless runs driven by the `FALLING_BLOCKS_AUTO*` env vars documented in the demo README.
- [Disconnect mid-match strands a game] → `Multiplayer.PeerDisconnected` handler in `Match` implements the spec's elimination/fallback rule; host ignores events from unknown peers.
- [Gravity timing differences across machines] → Accepted: because boards are private and only garbage crosses the wire, timing drift cannot desync a shared state; worst case is a slightly late garbage arrival.
- [Hard drop + immediate relay may arrive out of order vs. new piece spawn] → Garbage RPCs carry a sequence number; recipient applies at next lock, which is already the deferral point.

## Migration Plan

New, self-contained module: no migration. Rollback = remove `2d/falling_blocks/` and its `go.work` entry.

## Open Questions

(none — the gravity curve is fixed under Decisions, and the session size was resolved at a 4-player host-side cap enforced per the multiplayer-lobby spec.)
