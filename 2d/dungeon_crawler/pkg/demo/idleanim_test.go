package demo

import "testing"

func TestBobBoundedAndMoving(t *testing.T) {
	if got := bob(0, 0); got != 0 {
		t.Errorf("bob(0,0) = %v, want 0 baseline", got)
	}
	seenMin, seenMax := float32(99), float32(-99)
	changed := false
	prev := bob(0, 1.3)
	for i := 0; i <= 2000; i++ {
		v := bob(float64(i)*0.008, 1.3)
		if v < -bobAmp || v > bobAmp {
			t.Fatalf("bob exceeded amplitude: %v", v)
		}
		if v != prev {
			changed = true
		}
		prev = v
		if v < seenMin {
			seenMin = v
		}
		if v > seenMax {
			seenMax = v
		}
	}
	if !changed {
		t.Error("bob never moved")
	}
	if seenMax < bobAmp*0.9 || seenMin > -bobAmp*0.9 {
		t.Errorf("bob never approached its amplitude: [%v, %v]", seenMin, seenMax)
	}
	if bob(2.718, 0.9) != bob(2.718, 0.9) {
		t.Error("bob is not deterministic")
	}
}

func TestBlinkWindows(t *testing.T) {
	const step = 0.004
	closed := 0
	total := 0
	runLen := 0
	maxRun := 0
	for tt := 0.0; tt <= 60; tt += step {
		total++
		if blinkClosed(tt, 0.4) {
			closed++
			runLen++
			if runLen > maxRun {
				maxRun = runLen
			}
		} else {
			runLen = 0
		}
	}
	if closed == 0 {
		t.Fatal("no blinks in 60s")
	}
	frac := float64(closed) / float64(total)
	if frac < 0.01 || frac > 0.15 {
		t.Errorf("closed fraction %.3f outside [0.01,0.15]", frac)
	}
	if float64(maxRun)*step > 0.3 {
		t.Errorf("one blink lasted %vs, too long to read", float64(maxRun)*step)
	}
}

func TestPulse01Range(t *testing.T) {
	lo, hi := 2.0, -2.0
	for i := 0; i <= 1000; i++ {
		v := pulse01(float64(i)*0.01, 0.7, 1.0)
		if v < 0 || v > 1 {
			t.Fatalf("pulse01 out of [0,1]: %v", v)
		}
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	if lo > 0.05 || hi < 0.95 {
		t.Errorf("pulse01 never spanned its range: [%v, %v]", lo, hi)
	}
}

func TestSweepWindows(t *testing.T) {
	const step = 0.004
	hits := 0
	lastStart := -100.0
	for tt := 0.0; tt <= 30; tt += step {
		k, active := sweep01(tt, 0.2, 2.0)
		if !active {
			continue
		}
		if k < 0 || k > 1 {
			t.Fatalf("sweep progress %v outside [0,1]", k)
		}
		if hits == 0 || tt-lastStart > 0.9 {
			hits++
			lastStart = tt
		}
	}
	if hits < 5 {
		t.Errorf("expected repeated glint windows, saw %d runs", hits)
	}
}
