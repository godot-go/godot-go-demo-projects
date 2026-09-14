# pkg/demo — godot-go multiplayer binding notes (spike findings, task 4.1)

Findings against godot-go v0.3.40 on Godot 4.7.3 (Linux). All workarounds are
already baked into this package's code.

## What works out of the box
- `Node.Rpc` / `Node.RpcId` / `Node.RpcConfig` are bound and usable.
- `Object.Call(StringName, ...Variant)` handles arbitrary methods/arguments —
  this is the universal fallback and is how UI classes reach the autoloaded
  `FallingBlocksSession` (`session.call(...)` in ui.go).
- `Node.GetMultiplayer()` → `RefMultiplayerAPI` (engine-owned) is fine.
- ENet client unique IDs are large random 32-bit values on 4.5+; host is 1.
  Treat peer IDs as opaque `int32`.

## Gaps found (and the fallback used)
1. **Custom-class instances cannot be re-fetched via `ObjectCastTo`.**
   `GoCallback_GDClassBindingCreate` is a stub returning `nullptr`, so lazily
   binding a Godot-owned custom-class object (e.g. the autoload node) fails
   and `ObjectCastTo` panics. *Fallback:* call session methods through the
   generic `Object.Call` + Variant path (see `session` wrapper in `ui.go`).
   Builtin classes (Label/Button/Timer...) are unaffected — they have
   generated binding callbacks.
2. **`MultiplayerAPI.SetMultiplayerPeer(RefMultiplayerPeer)` segfaults.**
   Generated ptrcall encodes Ref arguments as `&interfaceValue`; the engine
   dereferences it as `Object**` → SIGSEGV. *Fallback:*
   `api.Call("set_multiplayer_peer", NewVariantGodotObject(owner))` in
   `attachPeer` (session.go). Any engine method taking a Ref argument must be
   invoked through `Call` with an OBJECT variant instead.
3. **`MultiplayerAPI.get_remote_sender_id` is not bound.** *Fallback:*
   `api.Call(sn("get_remote_sender_id"))` (Variant return) in
   `senderId()` (session.go).
4. **`SceneMultiplayer` signal names differ from Godot 3.x**: use
   `connected_to_server` (not `connected`), plus `connection_failed`,
   `server_disconnected`, `peer_connected`, `peer_disconnected`.
5. **`rpc_config(method, config)` expects the per-method dictionary itself**
   (`{"rpc_mode": <int>, ...}`) as the second argument — not a nested map
   keyed by method. Sending an RPC whose config was registered incorrectly
   fails at runtime with "Unable to get the RPC configuration".
6. **ENet `connection_failed` can take tens of seconds** on unreachable hosts.
   The session enforces its own 10s join deadline (`GetJoinState`) so the
   title screen reports errors promptly.
