package crawler

import "math/rand"

// RunState is the lifecycle of one attempt.
type RunState int

const (
	Running RunState = iota
	Won
	Lost
)

// Event is one structured line produced by a verb: the shell renders Text in
// the combat log and may use the numeric fields for effects. Kind is one of
// "log", "attack", "level_up", "descend", "win", "lose".
type Event struct {
	Kind    string
	Text    string
	Roll    int // d20 value that was rolled, 0 for flavor lines
	Total   int // roll + bonuses
	DC      int // AC or DC the roll faced
	Penalty int // multiple attack penalty applied
	Degree  Degree
	Damage  int
	At      Cell   // map position of whoever took the damage
	Subject string // "player" or "enemy" on damage events
	EnemyID int    // the enemy involved on "enemy" damage events
}

// Run is one attempt: the player, the current floor, the seeded RNG, and the
// town bonuses applied at run start. Verbs return Events; the caller never
// mutates state directly.
type Run struct {
	Seed            int64
	rng             *rand.Rand
	rollD20         func() int // swappable for deterministic tests
	Floor           *Floor
	FloorNum        int
	FloorsDescended int
	Player          *Player
	State           RunState
	Enhanced        map[string]bool // class → Guild level-2 enhancement
	Visible         map[Cell]bool   // recomputed after every verb
	renownEarned    int
	kills           int
	bossDefeated    bool
	nextEnemyID     int
}

// RunBonuses carries the town meta-layer's run-start upgrades. The shell
// derives it from saved meta; nil means a fresh character.
type RunBonuses struct {
	StrikeDamageBonus int // Smithy
	WorldBonus        int // Library
	SearchBonus       int // Archive
	Potions           int // Clinic
	Enhanced          map[string]bool
}

// NewRun starts a fresh attempt on floor 1 with the given class and town
// bonuses.
func NewRun(seed int64, class string, bonuses *RunBonuses) *Run {
	if bonuses == nil {
		bonuses = &RunBonuses{}
	}
	p := ClassStatline(class, 1)
	p.StrikeDamageBonus = bonuses.StrikeDamageBonus
	p.WorldBonus = bonuses.WorldBonus
	p.SearchBonus = bonuses.SearchBonus
	p.Potions = bonuses.Potions

	enhanced := bonuses.Enhanced
	if enhanced == nil {
		enhanced = map[string]bool{}
	}
	r := &Run{
		Seed:     seed,
		rng:      rand.New(rand.NewSource(seed)),
		FloorNum: 1,
		Player:   p,
		State:    Running,
		Enhanced: enhanced,
	}
	r.rollD20 = func() int { return r.rng.Intn(20) + 1 }
	r.buildFloor()
	return r
}

// buildFloor generates the current floor, stamps features and monsters, seats
// the player at the entrance, and refreshes visibility.
func (r *Run) buildFloor() {
	g := GenerateFloor(r.rng)
	PlaceFeatures(r.rng, g, r.FloorNum, func(kind string) *Enemy {
		r.nextEnemyID++
		var e *Enemy
		if kind == BossKind {
			e = BossStatline()
		} else {
			e = MonsterStatline(kind)
		}
		e.ID = r.nextEnemyID
		return e
	})
	r.Floor = g.Floor
	r.Player.Pos = g.Entrance
	r.updateVisibility()
}

// updateVisibility recomputes the sight mask around the player and records
// everything seen into the floor's exploration memory.
func (r *Run) updateVisibility() {
	r.Visible = ComputeFOV(r.Floor, r.Player.Pos, SightRadius)
	r.Floor.MarkExplored(r.Visible)
}

// enemyPhaseMarker prefixes the enemy turn block in the log.
func enemyPhaseMarker() Event {
	return Event{Kind: "log", Text: "— the enemies act —"}
}

func yourTurnMarker() Event {
	return Event{Kind: "log", Text: "— your turn —"}
}

// mapPenalty is the PF2e multiple attack penalty for the nth Strike.
func mapPenalty(strikesSoFar int) int {
	switch strikesSoFar {
	case 0:
		return 0
	case 1:
		return -5
	default:
		return -10
	}
}
