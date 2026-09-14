package fallingblocks

import "math/rand"

// Randomizer implements the 7-bag randomizer: every consecutive group of
// seven pieces contains each tetromino exactly once.
type Randomizer struct {
	rnd   *rand.Rand
	queue []Kind
}

// NewRandomizer creates a bag randomizer seeded deterministically.
func NewRandomizer(rnd *rand.Rand) *Randomizer {
	r := &Randomizer{rnd: rnd}
	r.refill()
	r.refill()
	return r
}

func (r *Randomizer) refill() {
	// Copy: AllKinds shares its backing array, which Shuffle would mutate.
	bag := append([]Kind(nil), AllKinds()...)
	r.rnd.Shuffle(len(bag), func(i, j int) { bag[i], bag[j] = bag[j], bag[i] })
	r.queue = append(r.queue, bag...)
}

// Next pops the next piece kind from the bag.
func (r *Randomizer) Next() Kind {
	if len(r.queue) < 8 {
		r.refill()
	}
	k := r.queue[0]
	r.queue = r.queue[1:]
	return k
}

// Preview returns the piece that will be produced by the next Next call.
func (r *Randomizer) Preview() Kind {
	if len(r.queue) < 8 {
		r.refill()
	}
	return r.queue[0]
}
