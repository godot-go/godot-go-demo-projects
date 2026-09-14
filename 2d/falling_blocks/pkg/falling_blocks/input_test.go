package fallingblocks

import (
	"testing"
	"time"
)

// A single quick tap must produce exactly one move: frames inside the
// initial delay window must not repeat.
func TestShiftRepeaterSingleTapMovesOnce(t *testing.T) {
	var s ShiftRepeater
	t0 := time.Now()
	if !s.Step(t0, 1, true) {
		t.Fatal("initial press should move")
	}
	// Simulate held frames every 16ms for 100ms (a quick tap's tail).
	for i := 1; i <= 6; i++ {
		if s.Step(t0.Add(time.Duration(i*16)*time.Millisecond), 1, false) {
			t.Fatalf("frame %d within delay window should not repeat", i)
		}
	}
}

// Holding past the initial delay must auto-repeat at the repeat rate.
func TestShiftRepeaterHoldRepeats(t *testing.T) {
	var s ShiftRepeater
	t0 := time.Now()
	if !s.Step(t0, 1, true) {
		t.Fatal("initial press should move")
	}
	if !s.Step(t0.Add(200*time.Millisecond), 1, false) {
		t.Fatal("hold past delay should repeat")
	}
	if s.Step(t0.Add(220*time.Millisecond), 1, false) {
		t.Fatal("frame inside repeat interval should not move")
	}
	if !s.Step(t0.Add(270*time.Millisecond), 1, false) {
		t.Fatal("hold past repeat interval should move again")
	}
}

// Changing direction must move immediately.
func TestShiftRepeaterDirectionChangeMoves(t *testing.T) {
	var s ShiftRepeater
	t0 := time.Now()
	if !s.Step(t0, 1, true) {
		t.Fatal("initial press should move")
	}
	if !s.Step(t0.Add(30*time.Millisecond), -1, false) {
		t.Fatal("direction change should move immediately")
	}
}

// Releasing must reset: no move on empty dir, next press moves.
func TestShiftRepeaterReleaseResets(t *testing.T) {
	var s ShiftRepeater
	t0 := time.Now()
	if !s.Step(t0, 1, true) {
		t.Fatal("initial press should move")
	}
	if s.Step(t0.Add(30*time.Millisecond), 0, false) {
		t.Fatal("released direction must not move")
	}
	if !s.Step(t0.Add(40*time.Millisecond), 1, true) {
		t.Fatal("press after release should move")
	}
}
