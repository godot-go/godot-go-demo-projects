package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// RegisterClassFallingBlocksLobby registers the lobby root class.
func RegisterClassFallingBlocksLobby() {
	ClassDBRegisterClass(NewFallingBlocksLobbyFromOwnerObject, []GDExtensionPropertyInfo{}, nil, func(t *FallingBlocksLobby) {
		ClassDBBindMethodVirtual(t, "V_FallingBlocksLobby_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_FallingBlocksLobby_Process", "_process", nil, nil)
		ClassDBBindMethod(t, "OnReady", "on_ready", nil, nil)
		ClassDBBindMethod(t, "OnStart", "on_start", nil, nil)
		ClassDBBindMethod(t, "OnLeave", "on_leave", nil, nil)
	})
}

func NewFallingBlocksLobbyFromOwnerObject(owner *GodotObject) GDClass {
	obj := &FallingBlocksLobby{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// FallingBlocksLobby is the root of Lobby.tscn.
type FallingBlocksLobby struct {
	ControlImpl

	session     *session
	infoLabel   Label
	slots       [4]Label
	readyButton Button
	startButton Button
	lastPoll    time.Time
	advancing   bool
	isHostPeer  bool
	autoReadied bool
	autoStarted bool
}

func (c *FallingBlocksLobby) GetClassName() string { return "FallingBlocksLobby" }
func (c *FallingBlocksLobby) GetParentClassName() string {
	return "Control"
}

func (c *FallingBlocksLobby) V_FallingBlocksLobby_Ready() {
	c.session = sessionNode(c)
	if c.session == nil {
		return // editor preview
	}
	c.infoLabel = findLabel(c, "Center/VBox/InfoLabel")
	for i := 0; i < 4; i++ {
		c.slots[i] = findLabel(c, fmt.Sprintf("Center/VBox/RosterBox/Slot%d", i))
	}
	c.isHostPeer = c.session.isHostPeer()
	c.readyButton = findButton(c, "Center/VBox/ButtonRow/ReadyButton")
	c.startButton = findButton(c, "Center/VBox/ButtonRow/StartButton")
	bindPressed(c.readyButton, c, "on_ready")
	bindPressed(c.startButton, c, "on_start")
	bindPressed(findButton(c, "Center/VBox/ButtonRow/LeaveButton"), c, "on_leave")
}

func (c *FallingBlocksLobby) V_FallingBlocksLobby_Process() {
	if c.session == nil || c.advancing {
		return
	}
	if time.Since(c.lastPoll) < 100*time.Millisecond {
		return
	}
	c.lastPoll = time.Now()

	var r Roster
	if err := json.Unmarshal([]byte(c.session.rosterJSON()), &r); err != nil {
		return
	}
	c.render(&r)

	if r.Phase == PhaseMatch {
		c.advancing = true
		changeScene(c, "res://Match.tscn")
		return
	}

	// Development automation: FALLING_BLOCKS_AUTOPLAY readies everyone and lets the
	// host start the match without a window.
	if os.Getenv("FALLING_BLOCKS_AUTOPLAY") != "" {
		if !c.isHostPeer && !c.autoReadied {
			c.session.requestSetReady(true)
			c.autoReadied = true
		}
		if c.isHostPeer {
			allReady := len(r.Players) >= 2
			for _, p := range r.Players {
				if !p.Ready && !p.Host {
					allReady = false
				}
			}
			if allReady && !c.autoStarted {
				c.autoStarted = true
				c.session.requestStart()
			}
		}
	}
}

func (c *FallingBlocksLobby) render(r *Roster) {
	selfID := c.session.selfId()
	isHost := c.session.isHostPeer()

	if isHost {
		setLabel(c.infoLabel, fmt.Sprintf("Lobby — you are hosting. Players: %d/%d", len(r.Players), maxPlayers))
		showNode(c.readyButton, false)
	} else {
		setLabel(c.infoLabel, "Lobby — joined host")
		showNode(c.readyButton, true)
	}

	for i := 0; i < 4; i++ {
		if i < len(r.Players) {
			p := r.Players[i]
			tag := "  "
			if p.Host {
				tag = "(HOST)"
			}
			state := "not ready"
			if p.Host || p.Ready {
				state = "READY"
			}
			marker := " "
			if p.ID == selfID {
				marker = ">"
			}
			setLabel(c.slots[i], fmt.Sprintf("%s %s %s [%s]", marker, p.Name, tag, state))
			showNode(c.slots[i], true)
		} else {
			showNode(c.slots[i], false)
		}
	}

	if isHost {
		allReady := len(r.Players) >= 2
		for _, p := range r.Players {
			if !p.Ready && !p.Host {
				allReady = false
			}
		}
		showNode(c.startButton, true)
		c.startButton.SetDisabled(!allReady)
	} else {
		showNode(c.startButton, false)
	}
}

func (c *FallingBlocksLobby) OnReady() {
	var r Roster
	if err := json.Unmarshal([]byte(c.session.rosterJSON()), &r); err != nil {
		return
	}
	if p := r.Find(c.session.selfId()); p != nil {
		c.session.requestSetReady(!p.Ready)
	}
}

func (c *FallingBlocksLobby) OnStart() {
	c.session.requestStart()
}

func (c *FallingBlocksLobby) OnLeave() {
	c.session.leave()
	changeScene(c, "res://Main.tscn")
}
