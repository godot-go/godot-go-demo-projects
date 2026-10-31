package crawler

// The PF2e proficiency layer: skill ranks, skill modifiers, class DC, and
// saves. The feature verbs route through SkillCheck, so the character sheet
// displays exactly the numbers the game rolls against.

// SkillNames is the PF2e skill list in canonical order; the character sheet
// renders all of them with rank marks.
var SkillNames = []string{
	"Acrobatics", "Arcana", "Athletics", "Crafting", "Deception", "Diplomacy",
	"Intimidation", "Lore", "Medicine", "Nature", "Occultism", "Performance",
	"Religion", "Society", "Stealth", "Survival", "Thievery", "Perception",
}

// skillKeyAbility maps each skill to its key ability.
var skillKeyAbility = map[string]string{
	"Acrobatics": "dex", "Arcana": "int", "Athletics": "str", "Crafting": "int",
	"Deception": "cha", "Diplomacy": "cha", "Intimidation": "cha", "Lore": "int",
	"Medicine": "wis", "Nature": "wis", "Occultism": "int", "Performance": "cha",
	"Religion": "wis", "Society": "int", "Stealth": "dex", "Survival": "wis",
	"Thievery": "dex", "Perception": "wis",
}

// worldSkills are the skills the demo's feature checks roll (doors, traps,
// search); the Library's WorldBonus applies to these alone.
var worldSkills = map[string]bool{"Athletics": true, "Thievery": true, "Perception": true}

// SkillKey returns the key ability of a skill.
func SkillKey(skill string) string { return skillKeyAbility[skill] }

// SkillTrained reports whether the character is trained in a skill.
// Everyone is trained in Perception — the Search verb relies on it — and
// each class adds its world skill (Fighter Athletics, Rogue Thievery).
// Everything else is untrained in the demo roster.
func (p *Player) SkillTrained(skill string) bool {
	if skill == "Perception" {
		return true
	}
	return skill == p.TrainedWorld()
}

// worldBonus returns the Library bonus where it applies.
func (p *Player) worldBonus(skill string) int {
	if worldSkills[skill] {
		return p.WorldBonus
	}
	return 0
}

// SkillMod is the full PF2e skill modifier: key ability + trained (level+2)
// + the Library's WorldBonus on world skills.
func (p *Player) SkillMod(skill string) int {
	return p.AbilityMod(SkillKey(skill)) + ProfBonus(p.Level, p.SkillTrained(skill)) + p.worldBonus(skill)
}

// SkillCheck resolves a skill check with the character's ranks. It feeds
// RunCheck the same pieces SkillMod totals, so the sheet never lies.
func (p *Player) SkillCheck(skill string, roll, dc int, trainedOnly bool) CheckResult {
	return RunCheck(roll, p.Level, p.SkillTrained(skill),
		p.AbilityMod(SkillKey(skill))+p.worldBonus(skill), dc, trainedOnly)
}

// KeyAbility is the class's signature ability, driving its class DC.
func (p *Player) KeyAbility() string { return p.classOf().KeyAbility }

// ClassDC = 10 + trained proficiency + key ability modifier.
func (p *Player) ClassDC() int {
	return 10 + ProfBonus(p.Level, true) + p.AbilityMod(p.KeyAbility())
}

// saveAbility maps each save to the ability it uses.
var saveAbility = map[string]string{"fort": "con", "ref": "dex", "will": "wis"}

// SaveMod is a saving-throw bonus: trained (level+2) + key ability. The
// demo has no save-forcing mechanic yet; the character sheet displays these
// derived from the PF2e formula.
func (p *Player) SaveMod(save string) int {
	return ProfBonus(p.Level, true) + p.AbilityMod(saveAbility[save])
}
