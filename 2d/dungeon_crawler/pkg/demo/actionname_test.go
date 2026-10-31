package demo

import (
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

func TestActionNameCachesOneStringNamePerAction(t *testing.T) {
	origName, origCache := newActionName, actionNames
	defer func() { newActionName, actionNames = origName, origCache }()

	calls := map[string]int{}
	seq := 0
	newActionName = func(s string) StringName {
		calls[s]++
		seq++
		var sn StringName // a real handle needs a live engine; tag a stand-in
		sn[0] = byte(seq)
		return sn
	}
	actionNames = map[string]StringName{}

	up1 := actionName("move_up")
	up2 := actionName("move_up")
	if up1 != up2 {
		t.Error("repeat lookup built a second StringName")
	}
	down := actionName("move_down")
	if down == up1 {
		t.Error("different actions share one StringName")
	}
	if len(calls) != 2 || calls["move_up"] != 1 || calls["move_down"] != 1 {
		t.Errorf("constructed = %v, want exactly one per name", calls)
	}
}
