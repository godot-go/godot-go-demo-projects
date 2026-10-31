package demo

import "math"

// Idle animation math shared by every procedural avatar. Everything is a
// pure function of the clock (c.now) and a static per-entity phase, so the
// same frame always paints the same way and entities never wobble in
// lockstep (phases come from fxPhase).

// bobAmp is the vertical idle sway in pixels — enough to read as alive,
// small enough to stay inside a 21 px tile.
const bobAmp = 1.2

// bob returns the vertical idle offset: a slow sine starting at rest.
func bob(t, phase float64) float32 {
	return bobAmp * float32(math.Sin(t*2.6+phase))
}

// blinkClosed reports whether the eyes are shut: a ~0.2s window once every
// ~4s of character time, staggered by phase.
func blinkClosed(t, phase float64) bool {
	f := math.Mod(t*0.25+phase*0.37, 1)
	return f < 0.05
}

// pulse01 maps a sine onto [0,1] at the given cycles-per-second — the work
// horse for breathing, glows, and opacity throbs.
func pulse01(t, phase, cyclesPerSec float64) float64 {
	return 0.5 + 0.5*math.Sin(t*cyclesPerSec*2*math.Pi+phase)
}

// sweep01 drives periodic sheens (sword glints, blade shines): during a
// 25% window of every period it reports progress 0→1, outside it false.
func sweep01(t, phase, period float64) (float64, bool) {
	f := math.Mod(t/period+phase, 1)
	if f >= 0.25 {
		return 0, false
	}
	return f / 0.25, true
}
