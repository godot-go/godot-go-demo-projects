// Package crawler holds the dungeon crawler rules: the Pathfinder-2e-flavored
// turn engine, d20 checks, dungeon generation and visibility, the run
// lifecycle, and the persistent town meta-layer. It knows nothing about Godot.
//
// The engine is verb-driven: a caller (the game shell) calls Stride, Strike,
// UseSkill, Search, EndTurn, and friends; each verb returns Events describing
// what happened so the caller can render a combat log, then repaints from the
// engine state. Randomness is injected via NewRun's seed so any run reproduces
// from its seed.
package crawler

// Cell is a grid position with (0,0) at the top-left.
type Cell struct{ X, Y int }

// Dir is one of the eight movement directions: four orthogonal and four
// diagonal.
type Dir int

const (
	Up Dir = iota
	Down
	Left
	Right
	UpLeft
	UpRight
	DownLeft
	DownRight
)

// Delta returns the one-cell offset for the direction.
func (d Dir) Delta() Cell {
	switch d {
	case Up:
		return Cell{0, -1}
	case Down:
		return Cell{0, 1}
	case Left:
		return Cell{-1, 0}
	case UpLeft:
		return Cell{-1, -1}
	case UpRight:
		return Cell{1, -1}
	case DownLeft:
		return Cell{-1, 1}
	case DownRight:
		return Cell{1, 1}
	default:
		return Cell{1, 0}
	}
}

// isDiagonal reports whether d moves on both axes.
func (d Dir) isDiagonal() bool { return d >= UpLeft }

// Feature is the interactive content stamped on a floor tile. FeatureNone is
// plain floor. Locked and hidden doors block movement until their check
// succeeds; revealed doors (FeatureOpenDoor) do not.
type Feature int

const (
	FeatureNone Feature = iota
	FeatureStairs
	FeatureLockedDoor
	FeatureOpenDoor
	FeatureHiddenDoor
	FeatureTrapArmed
	FeatureTrapDisabled
	FeaturePotion
	FeatureWeapon
)

// solid reports whether the feature blocks movement.
func (f Feature) solid() bool {
	return f == FeatureLockedDoor || f == FeatureHiddenDoor
}

// Floor is one dungeon level: geometry, features, exploration memory, and the
// monsters standing on it.
type Floor struct {
	W, H     int
	solid    []bool
	Features []Feature
	Explored []bool
	Spotted  []bool // armed traps the player has spotted; run-only
	Enemies  []*Enemy
}

// NewFloor returns a fully solid floor of the given size.
func NewFloor(w, h int) *Floor {
	solid := make([]bool, w*h)
	for i := range solid {
		solid[i] = true
	}
	return &Floor{
		W:        w,
		H:        h,
		solid:    solid,
		Features: make([]Feature, w*h),
		Explored: make([]bool, w*h),
		Spotted:  make([]bool, w*h),
	}
}

func (f *Floor) index(c Cell) int { return c.Y*f.W + c.X }

// InBounds reports whether c lies on the floor.
func (f *Floor) InBounds(c Cell) bool {
	return c.X >= 0 && c.X < f.W && c.Y >= 0 && c.Y < f.H
}

// SetFloor clears solidity at c (used by the generator).
func (f *Floor) SetFloor(c Cell) {
	if f.InBounds(c) {
		f.solid[f.index(c)] = false
	}
}

// SetWall restores solidity at c (used to seal surplus vault doorways).
func (f *Floor) SetWall(c Cell) {
	if f.InBounds(c) {
		f.solid[f.index(c)] = true
	}
}

// IsSolid reports whether terrain blocks the tile (wall).
func (f *Floor) IsSolid(c Cell) bool {
	return f.InBounds(c) && f.solid[f.index(c)]
}

// TileWalkable reports whether a creature may occupy the tile.
func (f *Floor) TileWalkable(c Cell) bool {
	return f.InBounds(c) && !f.solid[f.index(c)] && !f.Features[f.index(c)].solid()
}

// SetFeature stamps interactive content at c.
func (f *Floor) SetFeature(c Cell, feat Feature) {
	if f.InBounds(c) {
		f.Features[f.index(c)] = feat
	}
}

// FeatureAt returns the feature at c.
func (f *Floor) FeatureAt(c Cell) Feature {
	if !f.InBounds(c) {
		return FeatureNone
	}
	return f.Features[f.index(c)]
}

// SpotTrap marks an armed trap at c as spotted; the mark sticks for the
// floor's lifetime (harmless once the trap itself is spent).
func (f *Floor) SpotTrap(c Cell) {
	if f.InBounds(c) {
		f.Spotted[f.index(c)] = true
	}
}

// TrapSpotted reports whether the trap at c was spotted by a Search.
func (f *Floor) TrapSpotted(c Cell) bool {
	return f.InBounds(c) && f.Spotted[f.index(c)]
}

// EnemyAt returns the living enemy occupying c, or nil.
func (f *Floor) EnemyAt(c Cell) *Enemy {
	for _, e := range f.Enemies {
		if e.HP > 0 && e.Pos == c {
			return e
		}
	}
	return nil
}

// Enemy is a hostile creature. It shares the player's action economy and acts
// on its own after the player's turn ends. AttackedRecently drives the Rogue's
// Sneak Attack: it is true only if the enemy damaged the player during the
// most recent enemy phase.
type Enemy struct {
	ID               int
	Kind             string
	Level            int
	Pos              Cell
	HP               int
	MaxHP            int
	AC               int
	AtkBonus         int
	DmgDice          int // number of d6 damage dice
	DmgMod           int
	XP               int
	IsBoss           bool
	AttackedRecently bool
}
