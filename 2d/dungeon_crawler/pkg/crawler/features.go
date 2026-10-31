package crawler

import "fmt"

// SightRadius is how far the player can see; SearchRangeBase how far the
// Search action probes (widened by the Archive).
const (
	SightRadius     = 6
	SearchRangeBase = 2
)

// bumpFeature handles moving into a blocked feature tile. Hidden doors are
// secrets: bumping one feels exactly like bumping a wall.
func (r *Run) bumpFeature(c Cell) []Event {
	p := r.Player
	switch r.Floor.FeatureAt(c) {
	case FeatureLockedDoor:
		if p.ActionsLeft <= 0 {
			return []Event{{Kind: "log", Text: "No actions left this turn."}}
		}
		p.ActionsLeft--
		roll := r.rollD20()
		// PF2e split: the Fighter forces the door with Athletics, the
		// Rogue picks the lock with Thievery; the check is whichever
		// world skill the class trains.
		skill := p.TrainedWorld()
		if skill == "" {
			skill = "Athletics"
		}
		res := p.SkillCheck(skill, roll, DoorDC, true)
		var ev Event
		switch {
		case res.Untrained:
			ev = Event{Kind: "log",
				Text: "The door is locked. You lack the Athletics or Thievery training to open it."}
		case res.Degree >= Success:
			r.Floor.SetFeature(c, FeatureOpenDoor)
			ev = Event{Kind: "log", Roll: roll, Total: res.Total, DC: res.DC, Degree: res.Degree}
			if skill == "Thievery" {
				ev.Text = fmt.Sprintf("You pick the lock and the door swings open (d20=%d total %d vs DC %d).", roll, res.Total, res.DC)
			} else {
				ev.Text = fmt.Sprintf("You force the door open (d20=%d total %d vs DC %d).", roll, res.Total, res.DC)
			}
		default:
			ev = Event{Kind: "attack", Roll: roll, Total: res.Total, DC: res.DC, Degree: res.Degree}
			if skill == "Thievery" {
				ev.Text = fmt.Sprintf("The pick slips — the lock holds (d20=%d total %d vs DC %d).", roll, res.Total, res.DC)
			} else {
				ev.Text = fmt.Sprintf("You slam the locked door — it holds (d20=%d total %d vs DC %d).", roll, res.Total, res.DC)
			}
		}
		return append([]Event{ev}, r.maybeEndTurn()...)
	}
	return nil
}

// arriveOnTile resolves effects of entering a tile: potions, traps, stairs.
func (r *Run) arriveOnTile() []Event {
	p := r.Player
	switch r.Floor.FeatureAt(p.Pos) {
	case FeaturePotion:
		r.Floor.SetFeature(p.Pos, FeatureNone)
		p.Potions++
		return []Event{{Kind: "log", Text: "You pick up a potion."}}
	case FeatureWeapon:
		r.Floor.SetFeature(p.Pos, FeatureNone)
		w := r.rollLootWeapon()
		p.Weapons = append(p.Weapons, w)
		return []Event{{Kind: "log", Text: "You find a " + w.Name + "!"}}
	case FeatureTrapArmed:
		r.Floor.SetFeature(p.Pos, FeatureTrapDisabled)
		dmg := 4 + r.FloorNum
		p.HP -= dmg
		ev := Event{Kind: "log", Damage: dmg, Text: fmt.Sprintf("A trap snaps shut — %d damage!", dmg),
			At: p.Pos, Subject: "player"}
		if p.HP <= 0 {
			p.HP = 0
			r.State = Lost
			return append([]Event{ev}, Event{Kind: "lose", Text: "A trap was the end of you."})
		}
		return []Event{ev}
	case FeatureStairs:
		return r.descend()
	}
	return nil
}

// stepIntoTrapAttempt handles moving toward an armed trap tile: a trained
// Rogue disarms it from the edge; anyone else (or a failed check) walks in and
// triggers it, handled by arriveOnTile.
func (r *Run) stepIntoTrapAttempt(c Cell) []Event {
	p := r.Player
	if !p.SkillTrained("Thievery") {
		return nil // walk in and trigger
	}
	roll := r.rollD20()
	res := p.SkillCheck("Thievery", roll, TrapDC, true)
	if res.Degree >= Success {
		p.ActionsLeft--
		r.Floor.SetFeature(c, FeatureTrapDisabled)
		ev := Event{Kind: "log", Roll: roll, Total: res.Total, DC: res.DC, Degree: res.Degree,
			Text: fmt.Sprintf("You disarm the trap (d20=%d total %d vs DC %d).", roll, res.Total, res.DC)}
		return append([]Event{ev}, r.maybeEndTurn()...)
	}
	p.ActionsLeft--
	ev := Event{Kind: "attack", Roll: roll, Total: res.Total, DC: res.DC, Degree: res.Degree,
		Text: "Your hands fumble the mechanism..."}
	p.Pos = c
	return append(append([]Event{ev}, r.arriveOnTile()...), r.maybeEndTurn()...)
}

// Search probes for hidden doors and unspotted traps with a trained
// Perception roll. The single roll reveals the nearest hidden thing in
// reach — door or trap, doors winning ties.
func (r *Run) Search() []Event {
	defer r.updateVisibility()
	if r.State != Running {
		return []Event{{Kind: "log", Text: "The run is over."}}
	}
	p := r.Player
	if p.ActionsLeft <= 0 {
		return []Event{{Kind: "log", Text: "No actions left this turn."}}
	}
	p.ActionsLeft--
	roll := r.rollD20()
	reach := SearchRangeBase + p.SearchBonus

	doorRes := p.SkillCheck("Perception", roll, HiddenDoorDC, false)
	trapRes := p.SkillCheck("Perception", roll, TrapSpotDC, false)

	door, doorD := Cell{}, reach+1
	if doorRes.Degree >= Success {
		for y := 0; y < r.Floor.H; y++ {
			for x := 0; x < r.Floor.W; x++ {
				c := Cell{x, y}
				if r.Floor.FeatureAt(c) != FeatureHiddenDoor {
					continue
				}
				if d := abs(c.X-p.Pos.X) + abs(c.Y-p.Pos.Y); d <= reach && d < doorD {
					door, doorD = c, d
				}
			}
		}
	}
	trap, trapD := Cell{}, reach+1
	if trapRes.Degree >= Success {
		for y := 0; y < r.Floor.H; y++ {
			for x := 0; x < r.Floor.W; x++ {
				c := Cell{x, y}
				if r.Floor.FeatureAt(c) != FeatureTrapArmed || r.Floor.TrapSpotted(c) {
					continue
				}
				if d := abs(c.X-p.Pos.X) + abs(c.Y-p.Pos.Y); d <= reach && d < trapD {
					trap, trapD = c, d
				}
			}
		}
	}

	var ev Event
	switch {
	case doorD <= reach && doorD <= trapD:
		r.Floor.SetFloor(door)
		r.Floor.SetFeature(door, FeatureOpenDoor)
		ev = Event{Kind: "log", Roll: roll, Total: doorRes.Total, DC: doorRes.DC, Degree: doorRes.Degree,
			Text: fmt.Sprintf("Search finds a hidden door! (d20=%d total %d vs DC %d)", roll, doorRes.Total, doorRes.DC)}
	case trapD <= reach:
		r.Floor.SpotTrap(trap)
		ev = Event{Kind: "log", Roll: roll, Total: trapRes.Total, DC: trapRes.DC, Degree: trapRes.Degree,
			Text: fmt.Sprintf("You spot a trap! (d20=%d total %d vs DC %d)", roll, trapRes.Total, trapRes.DC)}
	default:
		ev = Event{Kind: "log", Roll: roll, Total: doorRes.Total, DC: doorRes.DC, Degree: doorRes.Degree,
			Text: "You search the area and find nothing."}
	}
	return append([]Event{ev}, r.maybeEndTurn()...)
}

// Quaff drinks a carried potion. It is free — no action cost.
func (r *Run) Quaff() []Event {
	if r.State != Running {
		return []Event{{Kind: "log", Text: "The run is over."}}
	}
	p := r.Player
	if p.Potions <= 0 {
		return []Event{{Kind: "log", Text: "You carry no potions."}}
	}
	if p.HP >= p.MaxHP {
		return []Event{{Kind: "log", Text: "You are at full health."}}
	}
	p.Potions--
	before := p.HP
	p.HP += PotionHeal
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	return []Event{{Kind: "log", Text: fmt.Sprintf("The potion restores %d HP.", p.HP-before)}}
}

// descend generates and enters the next floor. A fresh floor is a fresh
// day for the wizard: spell slots come back full.
func (r *Run) descend() []Event {
	r.FloorNum++
	r.FloorsDescended++
	r.buildFloor()
	evs := []Event{{Kind: "descend", Text: fmt.Sprintf("You descend to floor %d.", r.FloorNum)}}
	if r.Player.Class == "Wizard" && len(r.Player.Slots) > 0 {
		refreshWizardSlots(r.Player)
		evs = append(evs, Event{Kind: "log", Text: "You ready your spells anew — slots refilled."})
	}
	return evs
}

// Payout converts the run's deeds into renown. The boss bonus is added only
// the first time the player ever kills the boss (the shell knows via meta).
func (r *Run) Payout(bossFirstKill bool) int {
	total := r.renownEarned + r.FloorsDescended*FloorRenown
	if r.bossDefeated && bossFirstKill {
		total += BossFirstKillRenown
	}
	return total
}

// enemyTriggersTrap handles an AI stepping onto an armed trap. The damage is
// environmental, so no XP or renown is attributed; a slain enemy is simply
// removed.
func (r *Run) enemyTriggersTrap(e *Enemy) []Event {
	if r.Floor.FeatureAt(e.Pos) != FeatureTrapArmed {
		return nil
	}
	r.Floor.SetFeature(e.Pos, FeatureTrapDisabled)
	dmg := 4 + r.FloorNum
	e.HP -= dmg
	evs := []Event{{Kind: "log", Damage: dmg, Text: fmt.Sprintf("The %s triggers a trap!", e.Kind)}}
	if e.HP <= 0 {
		e.HP = 0
		r.removeEnemy(e)
		evs = append(evs, Event{Kind: "log", Text: fmt.Sprintf("The %s is killed by the trap.", e.Kind)})
	}
	return evs
}
