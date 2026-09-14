package demo

import (
	"encoding/json"
	"sort"
	"time"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// Session phases.
const (
	PhaseTitle = "title"
	PhaseLobby = "lobby"
	PhaseMatch = "match"
	PhaseOver  = "over"
)

// Join states.
const (
	JoinIdle     = 0
	JoinPending  = 1
	JoinOK       = 2
	JoinFailed   = 3
	JoinHostBusy = 4
)

const maxPlayers = 4

// PlayerInfo is one entry of the host-authoritative roster.
type PlayerInfo struct {
	ID      int32  `json:"id"`
	Name    string `json:"name"`
	Host    bool   `json:"host"`
	Ready   bool   `json:"ready"`
	Alive   bool   `json:"alive"`
	Score   int    `json:"score"`
	Lines   int    `json:"lines"`
	Pending int    `json:"pending"`
}

// Roster is replicated as JSON from host to all clients.
type Roster struct {
	Phase   string        `json:"phase"`
	Solo    bool          `json:"solo"`
	Winner  int32         `json:"winner"`
	Token   int64         `json:"token"`
	Players []*PlayerInfo `json:"players"`
}

func (r *Roster) Find(id int32) *PlayerInfo {
	for _, p := range r.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (r *Roster) AliveCount() int {
	n := 0
	for _, p := range r.Players {
		if p.Alive {
			n++
		}
	}
	return n
}

func (r *Roster) cloneJSON() String {
	b, err := json.Marshal(r)
	if err != nil {
		return NewStringWithUtf8Chars("")
	}
	return NewStringWithUtf8Chars(string(b))
}

func (r *Roster) cloneVariant() Variant {
	s := r.cloneJSON()
	defer s.Destroy()
	return NewVariantString(s)
}

// RegisterClassFallingBlocksSession registers the autoloaded session node class.
func RegisterClassFallingBlocksSession() {
	ClassDBRegisterClass(NewFallingBlocksSessionFromOwnerObject, []GDExtensionPropertyInfo{}, nil, func(t *FallingBlocksSession) {
		ClassDBBindMethodVirtual(t, "V_FallingBlocksSession_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_FallingBlocksSession_ExitTree", "_exit_tree", nil, nil)

		// Local/UI API
		ClassDBBindMethod(t, "Host", "host", []string{"port"}, nil)
		ClassDBBindMethod(t, "Join", "join", []string{"address", "port"}, nil)
		ClassDBBindMethod(t, "Leave", "leave", nil, nil)
		ClassDBBindMethod(t, "Solo", "solo", nil, nil)
		ClassDBBindMethod(t, "RequestSetReady", "request_set_ready", []string{"on"}, nil)
		ClassDBBindMethod(t, "RequestStart", "request_start", nil, nil)
		ClassDBBindMethod(t, "RequestRematch", "request_rematch", nil, nil)
		ClassDBBindMethod(t, "ReportGarbage", "report_garbage", []string{"count"}, nil)
		ClassDBBindMethod(t, "ReportStatus", "report_status", []string{"score", "lines", "pending"}, nil)
		ClassDBBindMethod(t, "ReportGameOver", "report_game_over", nil, nil)
		ClassDBBindMethod(t, "SetDisplayName", "set_display_name", []string{"name"}, nil)
		ClassDBBindMethod(t, "GetRoster", "get_roster", nil, nil)
		ClassDBBindMethod(t, "ConsumeGarbage", "consume_garbage", nil, nil)
		ClassDBBindMethod(t, "GetJoinState", "get_join_state", nil, nil)
		ClassDBBindMethod(t, "SelfId", "self_id", nil, nil)
		ClassDBBindMethod(t, "IsHostPeer", "is_host_peer", nil, nil)

		// Client -> host (ANY_PEER; host validates sender)
		ClassDBBindMethod(t, "SrvHello", "srv_hello", []string{"name"}, nil)
		ClassDBBindMethod(t, "SrvSetReady", "srv_set_ready", []string{"on"}, nil)
		ClassDBBindMethod(t, "SrvRequestStart", "srv_request_start", nil, nil)
		ClassDBBindMethod(t, "SrvGarbage", "srv_garbage", []string{"count"}, nil)
		ClassDBBindMethod(t, "SrvStatus", "srv_status", []string{"score", "lines", "pending"}, nil)
		ClassDBBindMethod(t, "SrvGameOver", "srv_game_over", nil, nil)
		ClassDBBindMethod(t, "SrvRematch", "srv_rematch", nil, nil)

		// Host -> all (AUTHORITY mode)
		ClassDBBindMethod(t, "CliRoster", "cli_roster", []string{"data"}, nil)
		ClassDBBindMethod(t, "CliMatchStart", "cli_match_start", nil, nil)
		ClassDBBindMethod(t, "CliMatchOver", "cli_match_over", []string{"winner"}, nil)
		ClassDBBindMethod(t, "CliGarbage", "cli_garbage", []string{"from", "count"}, nil)

		// MultiplayerAPI signal sinks
		ClassDBBindMethod(t, "OnPeerConnected", "on_peer_connected", []string{"id"}, nil)
		ClassDBBindMethod(t, "OnPeerDisconnected", "on_peer_disconnected", []string{"id"}, nil)
		ClassDBBindMethod(t, "OnNetConnected", "on_net_connected", nil, nil)
		ClassDBBindMethod(t, "OnNetConnectionFailed", "on_net_connection_failed", nil, nil)
		ClassDBBindMethod(t, "OnServerDisconnected", "on_server_disconnected", nil, nil)
	})
}

func NewFallingBlocksSessionFromOwnerObject(owner *GodotObject) GDClass {
	obj := &FallingBlocksSession{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// FallingBlocksSession is the autoloaded, scene-independent networking hub.
// It owns the multiplayer peer, the roster state, and every RPC endpoint.
type FallingBlocksSession struct {
	NodeImpl

	roster     Roster
	isHost     bool
	active     bool // a network session is currently attached
	solo       bool
	myName     string
	joinState  int
	ownGarbage int32
	matchToken int64

	peer         RefENetMultiplayerPeer
	joinDeadline time.Time
}

func (c *FallingBlocksSession) GetClassName() string { return "FallingBlocksSession" }
func (c *FallingBlocksSession) GetParentClassName() string {
	return "Node"
}

// --- bootstrap ----------------------------------------------------------------

func (c *FallingBlocksSession) V_FallingBlocksSession_Ready() {
	c.roster = Roster{Phase: PhaseTitle, Winner: -1}
	c.configureRpc()
	c.connectNetSignals()
}

func (c *FallingBlocksSession) V_FallingBlocksSession_ExitTree() {
	c.cleanupPeer()
}

func (c *FallingBlocksSession) configureRpc() {
	anyPeer := map[string]map[string]any{
		"srv_hello":         {},
		"srv_set_ready":     {},
		"srv_request_start": {},
		"srv_garbage":       {},
		"srv_status":        {},
		"srv_game_over":     {},
		"srv_rematch":       {},
	}
	authOnly := []string{"cli_roster", "cli_match_start", "cli_match_over", "cli_garbage"}
	for name := range anyPeer {
		c.rpcModeFor(name, int64(MULTIPLAYER_API_RPC_MODE_RPC_MODE_ANY_PEER))
	}
	for _, name := range authOnly {
		c.rpcModeFor(name, int64(MULTIPLAYER_API_RPC_MODE_RPC_MODE_AUTHORITY))
	}
}

func (c *FallingBlocksSession) rpcModeFor(method string, mode int64) {
	gdsn := NewStringNameWithUtf8Chars(method)
	defer gdsn.Destroy()
	inner := NewDictionary()
	defer inner.Destroy()
	k := NewVariantGoString("rpc_mode")
	defer k.Destroy()
	v := NewVariantInt(int(mode))
	defer v.Destroy()
	inner.Set(k, v)
	cfgv := NewVariantDictionary(inner)
	defer cfgv.Destroy()
	c.RpcConfig(gdsn, cfgv)
}

func (c *FallingBlocksSession) connectNetSignals() {
	mp := c.GetMultiplayer()
	if !mp.IsValid() {
		log.Warn("FallingBlocksSession: no multiplayer api at ready")
		return
	}
	api := mp.Ptr()
	sigs := []struct {
		signal string
		method string
	}{
		{"peer_connected", "on_peer_connected"},
		{"peer_disconnected", "on_peer_disconnected"},
		{"connected_to_server", "on_net_connected"},
		{"connection_failed", "on_net_connection_failed"},
		{"server_disconnected", "on_server_disconnected"},
	}
	for _, s := range sigs {
		ssn := NewStringNameWithUtf8Chars(s.signal)
		msn := NewStringNameWithUtf8Chars(s.method)
		callable := NewCallableWithObjectStringName(c, msn)
		api.Connect(ssn, callable, 0)
		callable.Destroy()
		msn.Destroy()
		ssn.Destroy()
	}
}

// --- session management ---------------------------------------------------------

// Host starts listening; returns 0 on success.
func (c *FallingBlocksSession) Host(port Variant) Variant {
	c.Leave()
	p := int32(port.ToInt())
	peer, ref := newENetPeer()
	err := peer.CreateServer(p, 16, 0, 0, 0)
	if err != OK {
		ref.Unref()
		log.Info("FallingBlocksSession.Host: create_server failed", zap.Any("error", err))
		return NewVariantInt(int(JoinHostBusy))
	}
	c.attachPeer(ref)
	c.isHost = true
	c.active = true
	c.solo = false
	c.matchToken = 0
	id := c.selfID()
	c.roster = Roster{Phase: PhaseLobby, Winner: -1, Players: []*PlayerInfo{{
		ID: id, Name: c.displayName(), Host: true, Ready: true, Alive: true,
	}}}
	c.broadcastRoster()
	return NewVariantInt(0)
}

// Join connects to a host; completion is reported through GetJoinState.
func (c *FallingBlocksSession) Join(address Variant, port Variant) Variant {
	c.Leave()
	addr := address.ToString()
	defer addr.Destroy()
	peer, ref := newENetPeer()
	err := peer.CreateClient(addr, int32(port.ToInt()), 0, 0, 0, 0)
	if err != OK {
		ref.Unref()
		c.joinState = JoinFailed
		return NewVariantInt(int(JoinFailed))
	}
	c.attachPeer(ref)
	c.isHost = false
	c.active = true
	c.solo = false
	c.joinState = JoinPending
	c.joinDeadline = time.Now().Add(10 * time.Second)
	return NewVariantInt(0)
}

// Leave detaches from the current session.
func (c *FallingBlocksSession) Leave() {
	c.cleanupPeer()
	c.isHost = false
	c.active = false
	c.solo = false
	c.joinState = JoinIdle
	c.ownGarbage = 0
	c.roster = Roster{Phase: PhaseTitle, Winner: -1}
}

// Solo starts an offline match.
func (c *FallingBlocksSession) Solo() {
	c.Leave()
	c.solo = true
	c.active = false
	id := c.selfID()
	c.roster = Roster{
		Phase: PhaseMatch, Solo: true, Winner: -1, Token: 1,
		Players: []*PlayerInfo{{ID: id, Name: c.displayName(), Host: true, Ready: true, Alive: true}},
	}
	c.matchToken = 1
}

func (c *FallingBlocksSession) cleanupPeer() {
	if c.peer != nil {
		c.peer.Ptr().Close()
		c.peer.Unref()
	}
	c.peer = nil
}

func (c *FallingBlocksSession) attachPeer(ref RefENetMultiplayerPeer) {
	c.peer = ref
	mp := c.GetMultiplayer()
	if !mp.IsValid() {
		return
	}
	// MultiplayerAPI.SetMultiplayerPeer cannot be used here: godot-go v0.3.40
	// encodes Ref<> interface arguments as &iface for ptrcall, which the engine
	// dereferences as garbage. Route through Object.Call with an OBJECT variant.
	api := mp.Ptr()
	vpeer := NewVariantGodotObject(ref.Ptr().GetGodotObjectOwner())
	defer vpeer.Destroy()
	gdsn := NewStringNameWithLatin1Chars("set_multiplayer_peer")
	defer gdsn.Destroy()
	api.Call(gdsn, vpeer)
}

func (c *FallingBlocksSession) selfID() int32 {
	mp := c.GetMultiplayer()
	if !mp.IsValid() {
		return 1
	}
	return mp.Ptr().GetUniqueId()
}

func (c *FallingBlocksSession) displayName() string {
	if c.myName != "" {
		return c.myName
	}
	return "Player"
}

// --- UI-facing state -------------------------------------------------------------

func (c *FallingBlocksSession) GetRoster() String {
	return c.roster.cloneJSON()
}

func (c *FallingBlocksSession) ConsumeGarbage() Variant {
	n := c.ownGarbage
	c.ownGarbage = 0
	return NewVariantInt(int(n))
}

func (c *FallingBlocksSession) GetJoinState() Variant {
	if c.joinState == JoinPending && !c.joinDeadline.IsZero() && time.Now().After(c.joinDeadline) {
		c.joinState = JoinFailed
	}
	return NewVariantInt(c.joinState)
}

func (c *FallingBlocksSession) SelfId() Variant {
	return NewVariantInt(int(c.selfID()))
}

func (c *FallingBlocksSession) IsHostPeer() Variant {
	return NewVariantBool(c.isHost)
}

func (c *FallingBlocksSession) SetDisplayName(name string) {
	c.myName = name
}

// --- client -> host requests -------------------------------------------------------

func (c *FallingBlocksSession) RequestSetReady(on Variant) {
	if c.solo {
		return
	}
	if c.isHost {
		if p := c.roster.Find(c.selfID()); p != nil {
			p.Ready = on.ToBool()
			c.broadcastRoster()
		}
		return
	}
	c.RpcId(1, sn("srv_set_ready"), on)
}

func (c *FallingBlocksSession) RequestStart() {
	if c.solo {
		return
	}
	if c.isHost {
		c.hostStartMatch()
		return
	}
	c.RpcId(1, sn("srv_request_start"))
}

func (c *FallingBlocksSession) RequestRematch() {
	if c.solo {
		return
	}
	if c.isHost {
		c.hostStartMatch()
		return
	}
	c.RpcId(1, sn("srv_rematch"))
}

func (c *FallingBlocksSession) ReportGarbage(count Variant) {
	if c.solo || !c.active {
		return
	}
	if c.isHost {
		c.hostGarbage(c.selfID(), count)
		return
	}
	c.RpcId(1, sn("srv_garbage"), count)
}

func (c *FallingBlocksSession) ReportStatus(score Variant, lines Variant, pending Variant) {
	if c.solo || !c.active {
		return
	}
	if c.isHost {
		c.hostStatus(c.selfID(), score, lines, pending)
		return
	}
	c.RpcId(1, sn("srv_status"), score, lines, pending)
}

func (c *FallingBlocksSession) ReportGameOver() {
	if c.solo || !c.active {
		return
	}
	if c.isHost {
		c.hostGameOver(c.selfID())
		return
	}
	c.RpcId(1, sn("srv_game_over"))
}

// --- server-side handlers (run on the host) ------------------------------------------

func (c *FallingBlocksSession) senderId() int32 {
	mp := c.GetMultiplayer()
	if !mp.IsValid() {
		return c.selfID()
	}
	v := mp.Ptr().Call(sn("get_remote_sender_id"))
	defer v.Destroy()
	return int32(v.ToInt())
}

func (c *FallingBlocksSession) knownPeer(id int32) bool {
	if c.roster.Find(id) != nil {
		return true
	}
	return id == c.selfID()
}

func (c *FallingBlocksSession) SrvHello(name Variant) {
	if !c.isHost {
		return
	}
	id := c.senderId()
	if id == 0 {
		return
	}
	if c.roster.Find(id) == nil {
		if len(c.roster.Players) >= maxPlayers {
			return
		}
		nm := name.ToString()
		c.roster.Players = append(c.roster.Players, &PlayerInfo{ID: id, Name: nm.ToUtf8(), Alive: true})
		nm.Destroy()
		if len(c.roster.Players) >= maxPlayers {
			c.setRefuseNewConnections(true)
		}
	}
	c.broadcastRoster()
}

func (c *FallingBlocksSession) SrvSetReady(on Variant) {
	if !c.isHost {
		return
	}
	id := c.senderId()
	if !c.knownPeer(id) {
		return
	}
	if p := c.roster.Find(id); p != nil {
		p.Ready = on.ToBool()
		c.broadcastRoster()
	}
}

func (c *FallingBlocksSession) SrvRequestStart() {
	if !c.isHost {
		return
	}
	if !c.knownPeer(c.senderId()) && c.senderId() != 1 {
		return
	}
	c.hostStartMatch()
}

func (c *FallingBlocksSession) SrvRematch() {
	if !c.isHost {
		return
	}
	if !c.knownPeer(c.senderId()) {
		return
	}
	c.hostStartMatch()
}

func (c *FallingBlocksSession) SrvGarbage(count Variant) {
	if !c.isHost {
		return
	}
	id := c.senderId()
	if !c.knownPeer(id) {
		return
	}
	c.hostGarbage(id, count)
}

func (c *FallingBlocksSession) SrvStatus(score Variant, lines Variant, pending Variant) {
	if !c.isHost {
		return
	}
	id := c.senderId()
	if !c.knownPeer(id) {
		return
	}
	c.hostStatus(id, score, lines, pending)
}

func (c *FallingBlocksSession) SrvGameOver() {
	if !c.isHost {
		return
	}
	id := c.senderId()
	if !c.knownPeer(id) {
		return
	}
	c.hostGameOver(id)
}

// --- host logic --------------------------------------------------------------------

func (c *FallingBlocksSession) hostStartMatch() {
	if len(c.roster.Players) < 2 {
		return
	}
	for _, p := range c.roster.Players {
		if !p.Ready {
			return
		}
	}
	c.matchToken++
	c.roster.Token = c.matchToken
	c.roster.Phase = PhaseMatch
	c.roster.Winner = -1
	for _, p := range c.roster.Players {
		p.Alive = true
		p.Score = 0
		p.Lines = 0
		p.Pending = 0
	}
	c.Rpc(sn("cli_match_start"))
	c.broadcastRoster()
}

func (c *FallingBlocksSession) hostGarbage(from int32, count Variant) {
	if c.roster.Phase != PhaseMatch {
		return
	}
	n := count.ToInt()
	if n <= 0 {
		return
	}
	if from != c.selfID() {
		// The host is a recipient too: cli_garbage is remote-only for the host.
		c.ownGarbage += int32(n)
	}
	c.Rpc(sn("cli_garbage"), NewVariantInt(int(from)), NewVariantInt(int(n)))
}

func (c *FallingBlocksSession) hostStatus(id int32, score Variant, lines Variant, pending Variant) {
	p := c.roster.Find(id)
	if p == nil {
		return
	}
	p.Score = score.ToInt()
	p.Lines = lines.ToInt()
	p.Pending = pending.ToInt()
	c.broadcastRoster()
}

func (c *FallingBlocksSession) hostGameOver(id int32) {
	c.hostEliminate(id)
}

func (c *FallingBlocksSession) hostEliminate(id int32) {
	if c.roster.Phase != PhaseMatch {
		return
	}
	p := c.roster.Find(id)
	if p == nil || !p.Alive {
		return
	}
	p.Alive = false
	if c.roster.AliveCount() <= 1 {
		c.roster.Phase = PhaseOver
		c.roster.Winner = -1
		for _, q := range c.roster.Players {
			if q.Alive {
				c.roster.Winner = q.ID
			}
		}
		c.Rpc(sn("cli_match_over"), NewVariantInt(int(c.roster.Winner)))
	}
	c.broadcastRoster()
}

func (c *FallingBlocksSession) broadcastRoster() {
	v := c.roster.cloneVariant()
	defer v.Destroy()
	c.Rpc(sn("cli_roster"), v)
}

// --- client-side handlers ------------------------------------------------------------

func (c *FallingBlocksSession) CliRoster(data Variant) {
	s := data.ToString()
	defer s.Destroy()
	var r Roster
	if err := json.Unmarshal([]byte(s.ToUtf8()), &r); err != nil {
		log.Warn("FallingBlocksSession.CliRoster: bad payload", zap.Error(err))
		return
	}
	sort.Slice(r.Players, func(i, j int) bool { return r.Players[i].ID < r.Players[j].ID })
	c.roster = r
}

func (c *FallingBlocksSession) CliMatchStart() {
	c.matchToken++
	c.ownGarbage = 0
	c.roster.Phase = PhaseMatch
	c.roster.Token = c.matchToken
	c.roster.Winner = -1
	for _, p := range c.roster.Players {
		p.Alive = true
		p.Score = 0
		p.Lines = 0
		p.Pending = 0
	}
}

func (c *FallingBlocksSession) CliMatchOver(winner Variant) {
	c.roster.Phase = PhaseOver
	c.roster.Winner = int32(winner.ToInt())
}

func (c *FallingBlocksSession) CliGarbage(from Variant, count Variant) {
	if int32(from.ToInt()) == c.selfID() {
		return
	}
	c.ownGarbage += int32(count.ToInt())
}

// --- multiplayer signal sinks ----------------------------------------------------------

func (c *FallingBlocksSession) OnPeerConnected(id Variant) {
	if !c.isHost {
		return
	}
	_ = id
}

func (c *FallingBlocksSession) OnPeerDisconnected(id Variant) {
	if !c.isHost {
		return
	}
	rid := int32(id.ToInt())
	if c.roster.Phase == PhaseMatch {
		c.hostEliminate(rid)
	}
	kept := c.roster.Players[:0]
	for _, p := range c.roster.Players {
		if p.ID != rid {
			kept = append(kept, p)
		}
	}
	c.roster.Players = kept
	if len(kept) < maxPlayers {
		c.setRefuseNewConnections(false)
	}
	c.broadcastRoster()
}

func (c *FallingBlocksSession) OnNetConnected() {
	if c.isHost || !c.active || c.joinState != JoinPending {
		return
	}
	c.joinState = JoinOK
	name := NewVariantGoString(c.displayName())
	defer name.Destroy()
	c.RpcId(1, sn("srv_hello"), name)
}

func (c *FallingBlocksSession) OnNetConnectionFailed() {
	if !c.isHost {
		log.Info("falling-blocks-demo: join failed (connection_failed)")
		c.joinState = JoinFailed
	}
}

func (c *FallingBlocksSession) OnServerDisconnected() {
	if !c.isHost {
		c.Leave()
	}
}

// --- helpers ---------------------------------------------------------------------------

func (c *FallingBlocksSession) setRefuseNewConnections(refuse bool) {
	if c.peer != nil {
		c.peer.Ptr().SetRefuseNewConnections(refuse)
	}
}

// sn builds a transient StringName; caller must Destroy.
func sn(s string) StringName {
	return NewStringNameWithUtf8Chars(s)
}

func newENetPeer() (ENetMultiplayerPeer, RefENetMultiplayerPeer) {
	gdsn := NewStringNameWithLatin1Chars("ENetMultiplayerPeer")
	defer gdsn.Destroy()
	owner := (*GodotObject)(unsafe.Pointer(
		CallFunc_GDExtensionInterfaceClassdbConstructObject2(gdsn.AsGDExtensionConstStringNamePtr()),
	))
	ref := NewENetMultiplayerPeerWithGodotOwnerObject(owner)
	return ref.Ptr(), ref
}
