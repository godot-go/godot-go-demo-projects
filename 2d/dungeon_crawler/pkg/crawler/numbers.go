package crawler

// This file is the tuning table: every number the balance hinges on lives
// here so playtesting edits touch one file.

// Dice is a count of identically sized dice.
type Dice struct{ N, Size int }

// PF2e cumulative XP thresholds: level 2 at 20 XP, then 40, 60, 80.
// Level 5 is the cap; the boss is the run's climax.
var xpCumulative = []int{0, 20, 40, 60, 80}

// LevelForXP maps accumulated XP onto a character level (capped at 5).
func LevelForXP(xp int) int {
	level := 1
	for level < len(xpCumulative) && xp >= xpCumulative[level] {
		level++
	}
	return level
}

// XPCumulative returns the XP needed to reach a level; levels above the cap
// report -1. The table is indexed one below the level it unlocks — reaching
// level N costs xpCumulative[N-1] — matching LevelForXP and awardKill.
func XPCumulative(level int) int {
	if level < 1 || level > len(xpCumulative) {
		return -1
	}
	return xpCumulative[level-1]
}

// Feature DCs, PF2e simple-DC "standard" tier.
const (
	DoorDC        = 15
	TrapDC        = 15
	TrapSpotDC    = 15
	HiddenDoorDC  = 15
	PotionHeal    = 10
	LayOnHandsPer = 4
)

// Renown payout weights.
const (
	FloorRenown         = 25
	BossFirstKillRenown = 100
)

// RenownPerKill scales with the slain monster's level.
func RenownPerKill(monsterLevel int) int { return 10 + 5*monsterLevel }

// ClassDef is a class as data: stat line, weapon, trained world skill, and
// its combat skill with its action cost.
type ClassDef struct {
	Name                         string
	HitBase                      int
	HitPerLevel                  int
	Str, Dex, Con, Int, Wis, Cha int
	Weapon                       Dice
	MeleeAbility                 string // "str" or "dex"
	KeyAbility                   string // signature ability, drives the class DC
	TrainedWorld                 string // "Athletics", "Thievery", or ""
	CombatSkill                  string // "" means the class skill is passive
	SkillCost                    int
	TrainedMartial               bool // PF2e weapon prof: Fighter only here
}

// Classes is the four-class roster.
var Classes = map[string]*ClassDef{
	"Fighter": {
		Name: "Fighter", HitBase: 14, HitPerLevel: 6,
		Str: 4, Dex: 1, Con: 3, Int: -1, Wis: 0, Cha: 0,
		Weapon: Dice{1, 8}, MeleeAbility: "str", KeyAbility: "str",
		TrainedWorld: "Athletics", CombatSkill: "power_attack", SkillCost: 2,
		TrainedMartial: true,
	},
	"Cleric": {
		Name: "Cleric", HitBase: 12, HitPerLevel: 5,
		Str: 2, Dex: 0, Con: 2, Int: -1, Wis: 4, Cha: 1,
		Weapon: Dice{1, 6}, MeleeAbility: "str", KeyAbility: "wis",
		TrainedWorld: "", CombatSkill: "lay_on_hands", SkillCost: 1,
	},
	"Wizard": {
		Name: "Wizard", HitBase: 10, HitPerLevel: 4,
		Str: -1, Dex: 1, Con: 1, Int: 4, Wis: 2, Cha: 0,
		Weapon: Dice{1, 4}, MeleeAbility: "str", KeyAbility: "int",
		TrainedWorld: "", CombatSkill: "ray", SkillCost: 1,
	},
	"Rogue": {
		Name: "Rogue", HitBase: 12, HitPerLevel: 5,
		Str: 0, Dex: 4, Con: 1, Int: 2, Wis: 1, Cha: 0,
		Weapon: Dice{1, 6}, MeleeAbility: "dex", KeyAbility: "dex",
		TrainedWorld: "Thievery", CombatSkill: "sneak_attack", SkillCost: 0,
	},
}

// ClassSkillCost returns the action cost of a skill for a class, or -1 when
// the class lacks the skill. "strike" costs everyone 1; passive class skills
// (cost 0) cannot be activated.
func ClassSkillCost(class, skill string) int {
	if skill == "strike" {
		return 1
	}
	def, ok := Classes[class]
	if !ok || def.CombatSkill != skill {
		return -1
	}
	return def.SkillCost
}

// ClassStatline builds a fresh level-N character of the class. AC is
// 12 + level + DEX; HP is the class line plus CON per level.
func ClassStatline(class string, level int) *Player {
	def := Classes[class]
	conMod := def.Con
	maxHP := def.HitBase + def.HitPerLevel*(level-1) + conMod*level
	p := &Player{
		Class:   class,
		Level:   level,
		HP:      maxHP,
		MaxHP:   maxHP,
		AC:      12 + level + def.Dex,
		Potions: 0,
		Weapons: []Weapon{StarterWeapons[class]},
		AttAbility: map[string]int{
			"str": def.Str, "dex": def.Dex, "con": def.Con,
			"int": def.Int, "wis": def.Wis, "cha": def.Cha,
		},
	}
	p.resetTurn()
	syncWizardBook(p)
	return p
}

// AbilityMod returns the modifier for "str", "dex", "int", and so on.
func (p *Player) AbilityMod(ability string) int { return p.AttAbility[ability] }

// classOf returns the class definition for a character.
func (p *Player) classOf() *ClassDef { return Classes[p.Class] }

// StrikeMeleeMod is the ability bonus on melee attack rolls. A finesse weapon
// uses DEX when it beats the class melee ability; otherwise the class line.
func (p *Player) StrikeMeleeMod() int {
	mod := p.AbilityMod(p.classOf().MeleeAbility)
	if p.EquippedWeapon().Finesse {
		if d := p.AbilityMod("dex"); d > mod {
			mod = d
		}
	}
	return mod
}

// StrikeDamage returns the weapon dice for a Strike: the wielded weapon.
func (p *Player) StrikeDamage() Dice { return p.EquippedWeapon().Dice }

// PowerAttackDamage is a Fighter's 2-action Strike damage: doubled weapon
// dice. The Guild level-2 enhancement's flat +4 is applied by the verb.
func PowerAttackDamage(p *Player) Dice {
	w := p.StrikeDamage()
	w.N *= 2
	return w
}

// LayOnHandsHeal is a Cleric's self-heal: 4 HP per level, 6 with the Guild
// level-2 enhancement.
func LayOnHandsHeal(p *Player, enhanced bool) int {
	if enhanced {
		return 6 * p.Level
	}
	return LayOnHandsPer * p.Level
}

// RayReach and RayDamage describe the Wizard's spell: a 1-action attack at
// reach 3 (4 enhanced) dealing 1d6 + level (+1d6 enhanced), keyed off INT.
func RayReach(p *Player, enhanced bool) int {
	if enhanced {
		return 4
	}
	return 3
}

func RayDamage(p *Player, enhanced bool) Dice {
	d := Dice{1, 6}
	if enhanced {
		d.N++
	}
	return d
}

// SneakAttackBonus is a Rogue's extra precision dice on a Strike against a
// target that has not attacked since the last enemy phase: 1d6, or 2d6
// enhanced.
func SneakAttackDice(p *Player, enhanced bool) int {
	if p.Class != "Rogue" {
		return 0
	}
	if enhanced {
		return 2
	}
	return 1
}

// The bestiary. Species by level; floors draw spawns from FloorSpawns.
var bestiary = map[string]int{
	"rat": 0, "goblin": 1, "skeleton": 2, "orc": 3, "shadow": 3,
	"troll": 4, "hag": 4, "wraith": 5,
}

// FloorSpawns lists eligible species per dungeon floor.
var FloorSpawns = map[int][]string{
	1: {"rat", "goblin"},
	2: {"goblin", "skeleton"},
	3: {"skeleton", "orc"},
	4: {"orc", "troll", "hag"},
	5: {"troll", "wraith"},
}

// BossKind names the floor-5 boss.
const BossKind = "Grim Warden"

// MonsterXP is the experience a species awards.
func MonsterXP(kind string) int {
	if kind == BossKind {
		return 100
	}
	return 10 + 5*bestiary[kind]
}

// MonsterStatline generates an enemy from its species level: HP 8+6L,
// AC 13+L, attack 6+L, 1d6+L damage.
func MonsterStatline(kind string) *Enemy {
	l := bestiary[kind]
	hp := 8 + 6*l
	return &Enemy{
		Kind:     kind,
		Level:    l,
		HP:       hp,
		MaxHP:    hp,
		AC:       13 + l,
		AtkBonus: 6 + l,
		DmgDice:  1, DmgMod: l,
		XP: MonsterXP(kind),
	}
}

// BossStatline builds the floor-5 elite: tripled HP, +2 attack, 2d6+5.
func BossStatline() *Enemy {
	l := bestiary["wraith"]
	hp := (8 + 6*l) * 5 / 2 // 2.5× a same-tier monster: lethal to a raw run, winnable with town upgrades
	return &Enemy{
		Kind:     BossKind,
		Level:    l,
		HP:       hp,
		MaxHP:    hp,
		AC:       18,
		AtkBonus: 12,
		DmgDice:  2, DmgMod: 5,
		XP:     100,
		IsBoss: true,
	}
}
