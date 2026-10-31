package crawler

import (
	"fmt"
	"strings"
)

// skillBlurbs describes each class skill for the character sheet: the base
// blurb and the clause appended when the Guild level-2 enhancement is
// active. The numbers mirror the rules helpers (LayOnHandsHeal, RayReach…).
var skillBlurbs = map[string][2]string{
	"Fighter": {"Class skill — Power Attack: 2 actions, Strike with doubled weapon dice",
		" · Guild-enhanced +4 damage"},
	"Cleric": {"Class skill — Lay on Hands: 1 action, heal 4 HP per level",
		" · Guild-enhanced 6 HP per level"},
	"Wizard": {"Class skill — Wizard spellcasting: Space casts the prepared spell (B opens the spellbook)",
		" · Guild-enhanced Ray of Frost: reach 4, 2d6"},
	"Rogue": {"Class skill — Sneak Attack: passive +1d6 precision against targets that have not attacked you",
		" · Guild-enhanced 2d6"},
}

// ClassBlurb is the one-line character-generation pitch for a class; the
// gate roster and the start screen share it so the descriptions never drift.
var classBlurbs = map[string]string{
	"Fighter": "d10 hits · Power Attack · shields",
	"Cleric":  "d8 hits · Lay on Hands · heal",
	"Wizard":  "d6 hits · prepared spells",
	"Rogue":   "d8 hits · Sneak Attack · traps",
}

func ClassBlurb(class string) string { return classBlurbs[class] }

// CharacterSheet renders the PF2e sheet shown in the C modal: statline,
// derived defense, abilities, the full skill list with ranks, the class
// skill, and the active town bonuses. Every number comes from the same
// helpers the verbs roll with, so the sheet cannot drift from the game.
func CharacterSheet(p *Player, enhanced bool) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s — Level %d\n", strings.ToUpper(p.Class), p.Level)
	if next := XPCumulative(p.Level + 1); next < 0 {
		fmt.Fprintf(&b, "XP %d (max level)\n", p.XP)
	} else {
		fmt.Fprintf(&b, "XP %d/%d\n", p.XP, next)
	}
	fmt.Fprintf(&b, "HP %d/%d · AC %d · Speed 25 ft\n", p.HP, p.MaxHP, p.AC)
	fmt.Fprintf(&b, "Saves F%+d R%+d W%+d · Class DC %d\n",
		p.SaveMod("fort"), p.SaveMod("ref"), p.SaveMod("will"), p.ClassDC())
	w := p.EquippedWeapon()
	strike := fmt.Sprintf("%s %dd%d %+d", w.Name, w.Dice.N, w.Dice.Size, p.StrikeMeleeMod()+p.WeaponProfMod())
	if w.Finesse {
		strike += " · finesse"
	}
	if w.Deadly > 0 {
		strike += fmt.Sprintf(" · deadly d%d", w.Deadly)
	}
	if p.StrikeDamageBonus > 0 {
		strike += fmt.Sprintf(" (+%d Smithy)", p.StrikeDamageBonus)
	}
	b.WriteString(strike + "\n")

	fmt.Fprintf(&b, "\nSTR %+d  DEX %+d  CON %+d  INT %+d  WIS %+d  CHA %+d\n",
		p.AbilityMod("str"), p.AbilityMod("dex"), p.AbilityMod("con"),
		p.AbilityMod("int"), p.AbilityMod("wis"), p.AbilityMod("cha"))

	var trained, untrained []string
	for _, s := range SkillNames {
		if p.SkillTrained(s) {
			trained = append(trained, fmt.Sprintf("T %s %+d", s, p.SkillMod(s)))
		} else {
			untrained = append(untrained, fmt.Sprintf("%s %+d", s, p.SkillMod(s)))
		}
	}
	b.WriteString("\nSkills (T = trained, level + 2):\n")
	b.WriteString(strings.Join(trained, "   ") + "\n")
	b.WriteString("U " + strings.Join(untrained, " · ") + "\n")

	blurb := skillBlurbs[p.Class]
	b.WriteString("\n" + blurb[0])
	if enhanced {
		b.WriteString(blurb[1])
	}
	b.WriteString("\n")

	if len(p.Spells) > 0 {
		b.WriteString("\nSpellbook")
		if line := p.SlotSummary(); line != "" {
			b.WriteString(" (" + line + ")")
		}
		b.WriteString(" — 1-" + fmt.Sprint(len(p.Spells)) + " in the B book:\n")
		for i, s := range p.Spells {
			mark := " "
			if i == p.Prepared {
				mark = "▸"
			}
			kind := fmt.Sprintf("%d-level", s.SlotLevel)
			if s.Cantrip {
				kind = "cantrip"
			}
			fmt.Fprintf(&b, "%s %d. %s — %s, %d action · %s\n", mark, i+1, s.Name, kind, s.Actions, s.Blurb)
		}
	}

	if line := activeBonusLine(p); line != "" {
		b.WriteString("\n" + line + "\n")
	}
	return b.String()
}

// activeBonusLine lists the town run-start bonuses the character carries.
func activeBonusLine(p *Player) string {
	var parts []string
	if p.StrikeDamageBonus > 0 {
		parts = append(parts, fmt.Sprintf("Smithy +%d strike damage", p.StrikeDamageBonus))
	}
	if p.WorldBonus > 0 {
		parts = append(parts, fmt.Sprintf("Library +%d skill checks", p.WorldBonus))
	}
	if p.SearchBonus > 0 {
		parts = append(parts, fmt.Sprintf("Archive +%d search range", p.SearchBonus))
	}
	if len(parts) == 0 {
		return ""
	}
	return "Active bonuses: " + strings.Join(parts, " · ")
}
