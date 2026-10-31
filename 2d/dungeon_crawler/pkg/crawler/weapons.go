package crawler

// Weapon is a PF2e-flavored melee weapon: weapon dice, damage type for
// display, and the three traits this demo models — Finesse (attack and damage
// with DEX when better than the class ability), Deadly (a bonus die of the
// given size when a Strike scores a critical success), and Martial gating
// (characters untrained in martial weapons take the PF2e −2 attack penalty).
type Weapon struct {
	Name       string
	Dice       Dice
	DamageType string
	Finesse    bool
	Deadly     int // bonus die size on a critical hit; 0 = none
	Martial    bool
}

// StarterWeapons are the class sidearms. Their dice reproduce the tuned
// ClassDef.Weapon values so introducing gear never rebalanced the classes.
var StarterWeapons = map[string]Weapon{
	"Fighter": {Name: "Longsword", Dice: Dice{1, 8}, DamageType: "slashing", Martial: true},
	"Cleric":  {Name: "Mace", Dice: Dice{1, 6}, DamageType: "bludgeoning"},
	"Wizard":  {Name: "Club", Dice: Dice{1, 4}, DamageType: "bludgeoning"},
	"Rogue":   {Name: "Short Sword", Dice: Dice{1, 6}, DamageType: "slashing", Finesse: true},
}

// LootWeapons is the dungeon drop pool. The martial entries deliberately
// tempt non-Fighters: they hit −2 with them but still gain the beefier dice.
var LootWeapons = []Weapon{
	{Name: "Battle Axe", Dice: Dice{1, 8}, DamageType: "slashing", Deadly: 12, Martial: true},
	{Name: "Rapier", Dice: Dice{1, 8}, DamageType: "piercing", Finesse: true, Deadly: 8, Martial: true},
	{Name: "Warhammer", Dice: Dice{1, 10}, DamageType: "bludgeoning", Martial: true},
	{Name: "Greataxe", Dice: Dice{1, 12}, DamageType: "slashing", Deadly: 12, Martial: true},
}

// EquippedWeapon returns the wielded weapon, falling back to the class
// sidearm for characters built without a pack.
func (p *Player) EquippedWeapon() Weapon {
	if p.Equipped >= 0 && p.Equipped < len(p.Weapons) {
		return p.Weapons[p.Equipped]
	}
	return StarterWeapons[p.Class]
}

// WeaponProfMod is the PF2e proficiency penalty: −2 when wielding a martial
// weapon without training; trained or simple weapons add nothing.
func (p *Player) WeaponProfMod() int {
	w := p.EquippedWeapon()
	if !w.Martial || p.classOf().TrainedMartial {
		return 0
	}
	return -2
}

// EquipWeapon swaps the wielded weapon. PF2e Interact: it costs one action in
// the dungeon; the already-wielded and invalid choices cost nothing.
func (r *Run) EquipWeapon(i int) []Event {
	if r.State != Running {
		return []Event{{Kind: "log", Text: "The run is over."}}
	}
	p := r.Player
	if i < 0 || i >= len(p.Weapons) {
		return []Event{{Kind: "log", Text: "You do not have that weapon."}}
	}
	if i == p.Equipped {
		return []Event{{Kind: "log", Text: "Already wielded."}}
	}
	if p.ActionsLeft <= 0 {
		return []Event{{Kind: "log", Text: "No actions left this turn."}}
	}
	p.ActionsLeft--
	p.Equipped = i
	ev := Event{Kind: "log", Text: "You ready the " + p.Weapons[i].Name + "."}
	return append([]Event{ev}, r.maybeEndTurn()...)
}

// rollLootWeapon picks a drop the player does not already carry so the pack
// fills with variety before duplicates; random within that pool.
func (r *Run) rollLootWeapon() Weapon {
	var fresh []Weapon
	for _, w := range LootWeapons {
		owned := false
		for _, have := range r.Player.Weapons {
			if have.Name == w.Name {
				owned = true
				break
			}
		}
		if !owned {
			fresh = append(fresh, w)
		}
	}
	pool := LootWeapons
	if len(fresh) > 0 {
		pool = fresh
	}
	return pool[r.rng.Intn(len(pool))]
}
