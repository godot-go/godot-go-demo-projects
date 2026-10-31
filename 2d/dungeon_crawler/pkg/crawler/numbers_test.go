package crawler

import "testing"

func TestXPThresholds(t *testing.T) {
	want := map[int]int{0: 1, 19: 1, 20: 2, 39: 2, 40: 3, 60: 4, 79: 4, 80: 5, 500: 5}
	for xp, level := range want {
		if got := LevelForXP(xp); got != level {
			t.Errorf("LevelForXP(%d) = %d, want %d", xp, got, level)
		}
	}
}

// XPCumulative is what the HUD bar and character sheet divide by, so its
// target must be the XP at which levels actually flip (awardKill follows
// LevelForXP): reaching N happens at XPCumulative(N) and not before.
func TestXPCumulativeMatchesLevelThresholds(t *testing.T) {
	for level := 2; level <= 5; level++ {
		next := XPCumulative(level)
		if LevelForXP(next) != level {
			t.Errorf("XPCumulative(%d)=%d, but LevelForXP says %d", level, next, LevelForXP(next))
		}
		if LevelForXP(next-1) == level {
			t.Errorf("XPCumulative(%d)=%d overshoots: level %d is already reached at %d", level, next, level, next-1)
		}
	}
	if got := XPCumulative(6); got != -1 {
		t.Errorf("XPCumulative above cap = %d, want -1", got)
	}
}

// A typical floor-1 clear (4 weakest-tier kills) must push a fresh character
// past the level-2 threshold, or early floors feel dead.
func TestFloorOneClearReachesLevel2(t *testing.T) {
	xp := 0
	for _, kind := range FloorSpawns[1] {
		xp += MonsterXP(kind)
	}
	xp *= 2 // ~4 kills: two of each floor-1 species
	if LevelForXP(xp) < 2 {
		t.Errorf("two rounds of floor-1 species give %d XP, not enough for level 2", xp)
	}
}

func TestMonsterStatlinesScaleWithLevel(t *testing.T) {
	weak := MonsterStatline("rat")
	strong := MonsterStatline("troll")
	if weak.HP <= 0 || strong.HP <= weak.HP {
		t.Errorf("HP must scale up: rat=%d troll=%d", weak.HP, strong.HP)
	}
	if strong.AtkBonus <= weak.AtkBonus || strong.AC <= weak.AC {
		t.Errorf("AC/attack must scale up: rat=%d/%d troll=%d/%d", weak.AC, weak.AtkBonus, strong.AC, strong.AtkBonus)
	}
	if MonsterXP("troll") <= MonsterXP("rat") {
		t.Error("XP must scale with monster level")
	}
	boss := BossStatline()
	if !boss.IsBoss || boss.HP <= MonsterStatline("wraith").HP {
		t.Errorf("boss must outstat the deepest tier, got HP %d", boss.HP)
	}
}

func TestClassLinesAreDistinct(t *testing.T) {
	fighter := ClassStatline("Fighter", 1)
	wizard := ClassStatline("Wizard", 1)
	if fighter.MaxHP <= wizard.MaxHP {
		t.Errorf("Fighter should be hardier: %d vs %d", fighter.MaxHP, wizard.MaxHP)
	}
	if ClassSkillCost("Fighter", "power_attack") != 2 {
		t.Error("Power Attack costs 2 actions")
	}
	if ClassSkillCost("Cleric", "lay_on_hands") != 1 {
		t.Error("Lay on Hands costs 1 action")
	}
	if ClassSkillCost("Wizard", "ray") != 1 {
		t.Error("Ray costs 1 action")
	}
	if ClassSkillCost("Rogue", "strike") != 1 {
		t.Error("Strike costs 1 action")
	}
}

func TestRenownWeights(t *testing.T) {
	if RenownPerKill(0) <= 0 || RenownPerKill(4) <= RenownPerKill(0) {
		t.Error("renown per kill must be positive and scale with monster level")
	}
	if FloorRenown <= 0 || BossFirstKillRenown <= 0 {
		t.Error("floor and boss renown must be positive")
	}
}
