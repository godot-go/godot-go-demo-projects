package demo

import (
	"fmt"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// bound resolves the wrapper the engine bound for n's runtime class. A node
// returned by a generated method is typed to that method's declared class
// (Node), so `node.(Label)` reports false for a Label; resolve first, then
// assert the result.
func bound(n Node) Object {
	if n == nil {
		panic("bound: node not found")
	}
	w, err := ObjectFromInstanceBinding(n.GetGodotObjectOwner())
	if err != nil {
		panic(fmt.Sprintf("bound: %v", err))
	}
	return w
}

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
	return bound(from.GetNodeOrNull(path)).(Label)
}

func findTimer(from Node, p string) Timer {
	path := nodePath(p)
	defer path.Destroy()
	return bound(from.GetNodeOrNull(path)).(Timer)
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

func actionJustPressed(name string) bool {
	input := GetInputSingleton()
	sn := NewStringNameWithUtf8Chars(name)
	defer sn.Destroy()
	return input.IsActionJustPressed(sn, true)
}

// drawFilledRect paints an axis-aligned rectangle. It uses draw_colored_polygon
// because godot-go mis-marshals draw_rect's filled/width arguments against
// Godot 4.8, warning on every call.
func drawFilledRect(c *SnakeGame, x, y, w, h float32, col Color) {
	pts := NewPackedVector2Array()
	defer pts.Destroy()
	pts.Append(NewVector2WithFloat32Float32(x, y))
	pts.Append(NewVector2WithFloat32Float32(x+w, y))
	pts.Append(NewVector2WithFloat32Float32(x+w, y+h))
	pts.Append(NewVector2WithFloat32Float32(x, y+h))
	uvs := NewPackedVector2Array()
	defer uvs.Destroy()
	c.DrawColoredPolygon(pts, col, uvs, RefTexture2D(nil))
}
