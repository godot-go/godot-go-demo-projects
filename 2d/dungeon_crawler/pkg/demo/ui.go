package demo

import (
	"math"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

func nodePath(p string) NodePath {
	s := NewStringWithLatin1Chars(p)
	defer s.Destroy()
	return NewNodePathWithString(s)
}

func findNode(from Node, p string) Node {
	path := nodePath(p)
	defer path.Destroy()
	return from.GetNodeOrNull(path)
}

func findLabel(from Node, p string) Label {
	path := nodePath(p)
	defer path.Destroy()
	return from.GetNodeOrNull(path).(Label)
}

func findColorRect(from Node, p string) ColorRect {
	path := nodePath(p)
	defer path.Destroy()
	return from.GetNodeOrNull(path).(ColorRect)
}

func findTimer(from Node, p string) Timer {
	path := nodePath(p)
	defer path.Destroy()
	return from.GetNodeOrNull(path).(Timer)
}

func setLabel(l Label, text string) {
	s := NewStringWithUtf8Chars(text)
	defer s.Destroy()
	l.SetText(s)
}

func showNode(n CanvasItem, visible bool) {
	if visible {
		n.Show()
	} else {
		n.Hide()
	}
}

// bindSignal connects a Go method of owner to obj's signal.
func bindSignal(obj Object, owner Object, signal string, method string) {
	ssn := NewStringNameWithUtf8Chars(signal)
	defer ssn.Destroy()
	msn := NewStringNameWithUtf8Chars(method)
	defer msn.Destroy()
	callable := NewCallableWithObjectStringName(owner, msn)
	defer callable.Destroy()
	obj.Connect(ssn, callable, 0)
}

// actionName interns the per-frame poll names: building and destroying a
// StringName on every poll was the game's dominant steady-state allocation,
// and interning keeps the GC (and its gctrace) quiet. Entries live for the
// process; there are only a couple dozen action names.
var (
	actionNames   = map[string]StringName{}
	newActionName = NewStringNameWithUtf8Chars // swapped out by cache tests
)

func actionName(name string) StringName {
	if sn, ok := actionNames[name]; ok {
		return sn
	}
	sn := newActionName(name)
	actionNames[name] = sn
	return sn
}

// actionJustPressed is for one-shot verbs: exact matching keeps modified
// presses (Ctrl+E, Shift+Q, …) from firing them.
func actionJustPressed(name string) bool {
	return GetInputSingleton().IsActionJustPressed(actionName(name), true)
}

// directionJustPressed polls a movement action with exact matching OFF:
// Godot rejects a Shift+arrow event outright under exact matching, and the
// Shift-tilt diagonals need that press to arrive.
func directionJustPressed(name string) bool {
	return GetInputSingleton().IsActionJustPressed(actionName(name), false)
}

// actionTapped is for modal toggles and dismissals: key auto-repeat (echo)
// must not re-fire them while a key is held.
func actionTapped(name string) bool {
	return GetInputSingleton().IsActionJustPressed(actionName(name), false)
}

// shiftHeld reports whether a Shift key is down.
func shiftHeld() bool {
	return GetInputSingleton().IsKeyPressed(constant.KEY_SHIFT)
}

// drawFilledRect paints an axis-aligned rectangle. CanvasItem.draw_rect
// itself warns on every call (godot-go's filled/width marshalling trips
// Godot 4.8's width check), so the ink lands through the lower-level
// RenderingServer.canvas_item_add_rect instead — same command, no warning,
// and no per-call array churn.
// Server-level rect fills are cached per node pointer: two FFI getters per
// pixel fill add real time when thousands run per frame.
var (
	fillServer  RenderingServer
	fillTargets = map[*DungeonCrawler]RID{}
)

func drawFilledRect(c *DungeonCrawler, x, y, w, h float32, col Color) {
	rid, ok := fillTargets[c]
	if !ok {
		if fillServer == nil {
			fillServer = GetRenderingServerSingleton()
		}
		rid = c.GetCanvasItem()
		fillTargets[c] = rid
	}
	fillServer.CanvasItemAddRect(rid, NewRect2WithFloat32Float32Float32Float32(x, y, w, h), col, false)
}

// drawPolygon paints an arbitrary filled polygon through the same
// draw_colored_polygon path that replaced draw_rect.
func drawPolygon(c *DungeonCrawler, pts [][2]float32, col Color) {
	arr := NewPackedVector2Array()
	defer arr.Destroy()
	for _, p := range pts {
		arr.Append(NewVector2WithFloat32Float32(p[0], p[1]))
	}
	uvs := NewPackedVector2Array()
	defer uvs.Destroy()
	c.DrawColoredPolygon(arr, col, uvs, RefTexture2D(nil))
}

// drawDisc approximates a filled circle as a 14-gon — enough roundness at
// tile scale without a texture pipeline.
func drawDisc(c *DungeonCrawler, cx, cy, r float32, col Color) {
	pts := make([][2]float32, 14)
	for i := range pts {
		a := float64(i) * 2 * math.Pi / 14
		pts[i] = [2]float32{cx + r*float32(math.Cos(a)), cy + r*float32(math.Sin(a))}
	}
	drawPolygon(c, pts, col)
}
