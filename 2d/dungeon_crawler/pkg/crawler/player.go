package crawler

// Player is the character. Town upgrade bonuses (StrikeDamageBonus from the
// Smithy, WorldBonus from the Library, SearchBonus from the Archive) are
// applied once at run start and held fixed for the run.
type Player struct {
	Class             string
	Level             int
	XP                int
	HP                int
	MaxHP             int
	AC                int
	Potions           int
	Pos               Cell
	ActionsLeft       int
	StrikesThisTurn   int
	StrikeDamageBonus int
	WorldBonus        int
	SearchBonus       int
	AttAbility        map[string]int
	Weapons           []Weapon    // owned gear; index 0 is the class sidearm
	Equipped          int         // index into Weapons
	Spells            []Spell     // wizard grimoire; index 0 is the default cast
	Prepared          int         // index into Spells cast by the Space key
	Slots             map[int]int // slot level -> slots remaining
}

// resetTurn starts a player turn: three action points and no attack penalty.
func (p *Player) resetTurn() {
	p.ActionsLeft = 3
	p.StrikesThisTurn = 0
}

// TrainedWorld returns the class-trained world skill ("Athletics",
// "Thievery", or "" when untrained for feature checks).
func (p *Player) TrainedWorld() string { return p.classOf().TrainedWorld }
