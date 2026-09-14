package fallingblocks

import "time"

// Auto-shift timing for horizontal movement: a single tap moves exactly
// once; holding repeats only after an initial delay (DAS), then at a
// faster repeat rate.
const (
	ShiftDelay  = 170 * time.Millisecond
	ShiftRepeat = 60 * time.Millisecond
)

// ShiftRepeater decides, per frame, whether a held direction should move
// the piece. It is pure Go (no Godot dependency) so it can be unit tested.
type ShiftRepeater struct {
	dir  int
	next time.Time
}

// Step reports whether the piece should move this frame.
// dir is -1 (left), +1 (right), or 0 (released); justPressed reports a
// fresh key press this frame.
func (s *ShiftRepeater) Step(now time.Time, dir int, justPressed bool) bool {
	if dir == 0 {
		s.dir = 0
		return false
	}
	if justPressed || dir != s.dir {
		s.dir = dir
		s.next = now.Add(ShiftDelay)
		return true
	}
	if !now.Before(s.next) {
		s.next = now.Add(ShiftRepeat)
		return true
	}
	return false
}
