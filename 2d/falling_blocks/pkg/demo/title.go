package demo

import (
	"os"
	"strconv"
	"strings"

	"github.com/godot-go/godot-go/pkg/log"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// RegisterClassFallingBlocksTitle registers the title-screen root class.
func RegisterClassFallingBlocksTitle() {
	ClassDBRegisterClass(NewFallingBlocksTitleFromOwnerObject, []GDExtensionPropertyInfo{}, nil, func(t *FallingBlocksTitle) {
		ClassDBBindMethodVirtual(t, "V_FallingBlocksTitle_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_FallingBlocksTitle_Process", "_process", nil, nil)
		ClassDBBindMethod(t, "OnHost", "on_host", nil, nil)
		ClassDBBindMethod(t, "OnJoin", "on_join", nil, nil)
		ClassDBBindMethod(t, "OnSolo", "on_solo", nil, nil)
	})
}

func NewFallingBlocksTitleFromOwnerObject(owner *GodotObject) GDClass {
	obj := &FallingBlocksTitle{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// FallingBlocksTitle is the root of Main.tscn.
type FallingBlocksTitle struct {
	ControlImpl

	session    *session
	nameEdit   LineEdit
	portEdit   LineEdit
	addrEdit   LineEdit
	joinEdit   LineEdit
	errorLabel Label
	status     int
}

func (c *FallingBlocksTitle) GetClassName() string { return "FallingBlocksTitle" }
func (c *FallingBlocksTitle) GetParentClassName() string {
	return "Control"
}

func (c *FallingBlocksTitle) V_FallingBlocksTitle_Ready() {
	c.session = sessionNode(c)
	if c.session == nil {
		return // running in the editor preview; no autoload present
	}
	c.nameEdit = findLineEdit(c, "Center/VBox/NameRow/NameEdit")
	c.portEdit = findLineEdit(c, "Center/VBox/HostRow/PortEdit")
	c.addrEdit = findLineEdit(c, "Center/VBox/JoinRow/AddrEdit")
	c.joinEdit = findLineEdit(c, "Center/VBox/JoinRow/JoinPortEdit")
	c.errorLabel = findLabel(c, "Center/VBox/ErrorLabel")
	bindPressed(findButton(c, "Center/VBox/HostRow/HostButton"), c, "on_host")
	bindPressed(findButton(c, "Center/VBox/JoinRow/JoinButton"), c, "on_join")
	bindPressed(findButton(c, "Center/VBox/SoloButton"), c, "on_solo")

	// Development automation (headless LAN testing):
	//   FALLING_BLOCKS_AUTOHOST=<port>              auto-host on ready
	//   FALLING_BLOCKS_AUTOJOIN=<addr>:<port>       auto-join on ready
	if h := os.Getenv("FALLING_BLOCKS_AUTOHOST"); h != "" {
		setEdit(c.portEdit, h)
		c.OnHost()
		return
	}
	if j := os.Getenv("FALLING_BLOCKS_AUTOJOIN"); j != "" {
		addr, port := splitHostPort(j)
		setEdit(c.addrEdit, addr)
		setEdit(c.joinEdit, port)
		c.OnJoin()
	}
}

func setEdit(le LineEdit, s string) {
	gds := NewStringWithUtf8Chars(s)
	defer gds.Destroy()
	le.SetText(gds)
}

func splitHostPort(s string) (string, string) {
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, "7777"
}

func (c *FallingBlocksTitle) V_FallingBlocksTitle_Process() {
	if c.session == nil || c.status != JoinPending {
		return
	}
	switch c.session.joinState() {
	case JoinOK:
		c.status = JoinIdle
		changeScene(c, "res://Lobby.tscn")
	case JoinFailed:
		c.status = JoinIdle
		log.Info("falling-blocks-demo: title reports join failure to user")
		c.session.leave()
		c.setError("Could not reach host. Check the address/port and try again.")
	}
}

func (c *FallingBlocksTitle) setError(msg string) {
	showNode(c.errorLabel, msg != "")
	if msg != "" {
		setLabel(c.errorLabel, msg)
	}
}

func (c *FallingBlocksTitle) parsePort(text string, def int) int {
	if n, err := strconv.Atoi(text); err == nil && n > 0 && n < 65536 {
		return n
	}
	return def
}

func (c *FallingBlocksTitle) OnHost() {
	c.session.setDisplayName(c.nameEditText())
	if c.session.host(c.parsePort(lineEditText(c.portEdit), 7777)) == JoinHostBusy {
		c.setError("Could not host: that port is already in use.")
		return
	}
	c.setError("")
	changeScene(c, "res://Lobby.tscn")
}

func (c *FallingBlocksTitle) OnJoin() {
	c.session.setDisplayName(c.nameEditText())
	c.setError("Connecting...")
	c.status = JoinPending
	addr := lineEditText(c.addrEdit)
	if addr == "" {
		addr = "localhost"
	}
	c.session.join(addr, c.parsePort(lineEditText(c.joinEdit), 7777))
}

func (c *FallingBlocksTitle) OnSolo() {
	c.session.setDisplayName(c.nameEditText())
	c.session.solo()
	changeScene(c, "res://Match.tscn")
}

func (c *FallingBlocksTitle) nameEditText() string {
	t := lineEditText(c.nameEdit)
	if t == "" {
		return "Player"
	}
	return t
}
