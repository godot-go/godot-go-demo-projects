package demo

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// session wraps the autoloaded FallingBlocksSession node. godot-go v0.3.40 cannot
// lazily bind native-owned custom class instances (binding_create is a stub),
// so every call goes through the generic Object.Call + Variant path instead of
// a type assertion.
type session struct {
	obj Object
}

func sessionNode(from Node) *session {
	n := findNode(from, "/root/Session")
	if n == nil {
		return nil
	}
	return &session{obj: n}
}

func (s *session) call(name string, args ...Variant) Variant {
	gdsn := NewStringNameWithUtf8Chars(name)
	defer gdsn.Destroy()
	return s.obj.Call(gdsn, args...)
}

func (s *session) callVoid(name string) {
	v := s.call(name)
	v.Destroy()
}

func (s *session) setDisplayName(name string) {
	v := NewVariantGoString(name)
	defer v.Destroy()
	r := s.call("set_display_name", v)
	r.Destroy()
}

func (s *session) host(port int) int {
	v := NewVariantInt(port)
	defer v.Destroy()
	r := s.call("host", v)
	defer r.Destroy()
	return r.ToInt()
}

func (s *session) join(address string, port int) int {
	va := NewVariantGoString(address)
	defer va.Destroy()
	vp := NewVariantInt(port)
	defer vp.Destroy()
	r := s.call("join", va, vp)
	defer r.Destroy()
	return r.ToInt()
}

func (s *session) leave()          { s.callVoid("leave") }
func (s *session) solo()           { s.callVoid("solo") }
func (s *session) requestStart()   { s.callVoid("request_start") }
func (s *session) requestRematch() { s.callVoid("request_rematch") }
func (s *session) reportGameOver() { s.callVoid("report_game_over") }

func (s *session) requestSetReady(on bool) {
	v := NewVariantBool(on)
	defer v.Destroy()
	r := s.call("request_set_ready", v)
	r.Destroy()
}

func (s *session) reportGarbage(count int) {
	v := NewVariantInt(count)
	defer v.Destroy()
	r := s.call("report_garbage", v)
	r.Destroy()
}

func (s *session) reportStatus(score, lines, pending int) {
	vs := NewVariantInt(score)
	defer vs.Destroy()
	vl := NewVariantInt(lines)
	defer vl.Destroy()
	vp := NewVariantInt(pending)
	defer vp.Destroy()
	r := s.call("report_status", vs, vl, vp)
	r.Destroy()
}

func (s *session) rosterJSON() string {
	v := s.call("get_roster")
	defer v.Destroy()
	str := v.ToString()
	defer str.Destroy()
	return str.ToUtf8()
}

func (s *session) consumeGarbage() int {
	v := s.call("consume_garbage")
	defer v.Destroy()
	return v.ToInt()
}

func (s *session) joinState() int {
	v := s.call("get_join_state")
	defer v.Destroy()
	return v.ToInt()
}

func (s *session) selfId() int32 {
	v := s.call("self_id")
	defer v.Destroy()
	return int32(v.ToInt())
}

func (s *session) isHostPeer() bool {
	v := s.call("is_host_peer")
	defer v.Destroy()
	return v.ToBool()
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
	return ObjectCastTo(from.GetNodeOrNull(path), "Label").(Label)
}

func findLineEdit(from Node, p string) LineEdit {
	path := nodePath(p)
	defer path.Destroy()
	return ObjectCastTo(from.GetNodeOrNull(path), "LineEdit").(LineEdit)
}

func findButton(from Node, p string) Button {
	path := nodePath(p)
	defer path.Destroy()
	return ObjectCastTo(from.GetNodeOrNull(path), "Button").(Button)
}

func findPanel(from Node, p string) Panel {
	path := nodePath(p)
	defer path.Destroy()
	return ObjectCastTo(from.GetNodeOrNull(path), "Panel").(Panel)
}

func findTimer(from Node, p string) Timer {
	path := nodePath(p)
	defer path.Destroy()
	return ObjectCastTo(from.GetNodeOrNull(path), "Timer").(Timer)
}

func setLabel(l Label, text string) {
	s := NewStringWithUtf8Chars(text)
	defer s.Destroy()
	l.SetText(s)
}

func lineEditText(le LineEdit) string {
	s := le.GetText()
	defer s.Destroy()
	return s.ToUtf8()
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

func bindPressed(b Button, owner Object, method string) {
	bindSignal(b, owner, "pressed", method)
}

func changeScene(from Node, path string) {
	s := NewStringWithUtf8Chars(path)
	defer s.Destroy()
	tree := from.GetTree()
	tree.ChangeSceneToFile(s)
}

func actionPressed(name string) bool {
	input := GetInputSingleton()
	sn := NewStringNameWithUtf8Chars(name)
	defer sn.Destroy()
	return input.IsActionPressed(sn, true)
}

func actionJustPressed(name string) bool {
	input := GetInputSingleton()
	sn := NewStringNameWithUtf8Chars(name)
	defer sn.Destroy()
	return input.IsActionJustPressed(sn, true)
}
