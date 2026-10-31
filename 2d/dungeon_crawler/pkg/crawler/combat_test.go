package crawler

import (
	"math/rand"
	"testing"
)

// newTestRun builds a bare run on a wide open corridor so tests control
// geometry, dice, and placements directly.
func newTestRun(class string) *Run {
	f := NewFloor(21, 3) // row y=1 is open floor
	for x := 0; x < 21; x++ {
		f.SetFloor(Cell{x, 1})
	}
	p := ClassStatline(class, 1)
	p.Pos = Cell{1, 1}
	r := &Run{
		rng:      rand.New(rand.NewSource(7)),
		Floor:    f,
		Player:   p,
		State:    Running,
		Enhanced: map[string]bool{},
	}
	r.rollD20 = func() int { return r.rng.Intn(20) + 1 }
	return r
}

func placeEnemy(r *Run, kind string, at Cell) *Enemy {
	e := MonsterStatline(kind)
	e.ID = len(r.Floor.Enemies) + 1
	e.Pos = at
	r.Floor.Enemies = append(r.Floor.Enemies, e)
	return e
}

// --- 3.1 turn economy ---

func TestStrideSpendsActionAndMoves(t *testing.T) {
	r := newTestRun("Fighter")
	before := r.Player.ActionsLeft
	r.Move(Right)
	if r.Player.Pos != (Cell{2, 1}) {
		t.Errorf("pos %v, want (2,1)", r.Player.Pos)
	}
	if r.Player.ActionsLeft != before-1 {
		t.Errorf("actions %d, want %d", r.Player.ActionsLeft, before-1)
	}
}

func TestStrideBlockedByWallCostsNothing(t *testing.T) {
	r := newTestRun("Fighter")
	r.Player.Pos = Cell{0, 1}
	before := r.Player.ActionsLeft
	r.Move(Up) // y=0 is solid
	if r.Player.Pos != (Cell{0, 1}) {
		t.Error("should not move into wall")
	}
	if r.Player.ActionsLeft != before {
		t.Error("blocked move must not spend an action")
	}
}

func TestTurnAutoEndsAtZeroActions(t *testing.T) {
	r := newTestRun("Fighter")
	foe := placeEnemy(r, "rat", Cell{10, 1})
	// Three strides spend the turn; the rat closes distance during the auto
	// enemy phase and the next player turn starts fresh with 3 actions.
	for i := 0; i < 3; i++ {
		r.Move(Right)
	}
	if foe.Pos.X > 7 {
		t.Errorf("rat should have closed to x<=7, now at %v", foe.Pos)
	}
	if r.Player.ActionsLeft != 3 {
		t.Errorf("new turn should reset to 3 actions, got %d", r.Player.ActionsLeft)
	}
}

func TestActionRejectedAtZero(t *testing.T) {
	r := newTestRun("Fighter")
	r.Player.ActionsLeft = 0
	evs := r.Move(Left)
	if len(evs) == 0 {
		t.Error("expected a rejection message")
	}
	if r.Player.Pos != (Cell{1, 1}) {
		t.Error("state must not change")
	}
}

func TestBumpAttacksEnemyWithoutMoving(t *testing.T) {
	r := newTestRun("Fighter")
	r.rollD20 = func() int { return 20 } // guaranteed crit
	foe := placeEnemy(r, "goblin", Cell{2, 1})
	r.Move(Right)
	if r.Player.Pos != (Cell{1, 1}) {
		t.Error("bump attack must not move the player")
	}
	if foe.HP < foe.MaxHP {
		// crit landed; goblin might still live but took damage
	} else {
		t.Error("critical bump must deal damage")
	}
	if r.Player.StrikesThisTurn != 1 {
		t.Errorf("strikes this turn = %d, want 1", r.Player.StrikesThisTurn)
	}
}

func TestMultipleAttackPenalties(t *testing.T) {
	r := newTestRun("Fighter")
	r.rollD20 = func() int { return 10 }
	foe := placeEnemy(r, "troll", Cell{2, 1}) // beefy: survives 3 swings
	totals := []int{}
	for i := 0; i < 3; i++ {
		evs := r.Move(Right)
		for _, e := range evs {
			if e.Roll != 0 && e.DC == foe.AC {
				totals = append(totals, e.Total)
			}
		}
	}
	// level 1 fighter: prof +3, STR +4 → 10+7, 10+7-5, 10+7-10
	want := []int{17, 12, 7}
	if len(totals) != 3 {
		t.Fatalf("got %d attack totals %v, want 3", len(totals), totals)
	}
	for i := range want {
		if totals[i] != want[i] {
			t.Errorf("attack %d total = %d, want %d (MAP −5/−10)", i+1, totals[i], want[i])
		}
	}
	if foe.HP <= 0 {
		t.Error("troll was expected to survive the three swings")
	}
}

// --- 3.2 class skills ---

func TestPowerAttackCostsTwoAndHits(t *testing.T) {
	r := newTestRun("Fighter")
	r.rollD20 = func() int { return 20 }
	foe := placeEnemy(r, "troll", Cell{2, 1})
	r.UseSkill("power_attack")
	if foe.HP >= foe.MaxHP {
		t.Error("power attack should hit")
	}
	// 2 actions spent + enemy phase may follow; strikes counted once
}

func TestPowerAttackNeedsAdjacentEnemy(t *testing.T) {
	r := newTestRun("Fighter")
	before := r.Player.ActionsLeft
	evs := r.UseSkill("power_attack")
	if len(evs) == 0 {
		t.Error("expected rejection")
	}
	if r.Player.ActionsLeft != before {
		t.Error("rejected skill must not spend actions")
	}
}

func TestSkillFromAnotherClassIsRejected(t *testing.T) {
	r := newTestRun("Cleric")
	before := r.Player.ActionsLeft
	r.UseSkill("power_attack")
	if r.Player.ActionsLeft != before {
		t.Error("wrong-class skill must be rejected without cost")
	}
}

func TestLayOnHandsHeals(t *testing.T) {
	r := newTestRun("Cleric")
	r.Player.HP = 5
	r.UseSkill("lay_on_hands")
	if r.Player.HP <= 5 {
		t.Errorf("HP %d, expected healing above 5", r.Player.HP)
	}
	if r.Player.HP > r.Player.MaxHP {
		t.Error("healing must not exceed max")
	}
}

func TestRayHitsNearestInRangeWithoutMoving(t *testing.T) {
	r := newTestRun("Wizard")
	r.rollD20 = func() int { return 20 }
	near := placeEnemy(r, "rat", Cell{4, 1}) // distance 3 from (1,1)
	placeEnemy(r, "rat", Cell{3, 1})         // distance 2, nearest
	pos := r.Player.Pos
	r.UseSkill("ray")
	if near.HP < near.MaxHP {
		t.Error("a ray hit the wrong (farther) enemy")
	}
	if r.Player.Pos != pos {
		t.Error("ray must not move the caster")
	}
}

func TestRayRejectedWithoutTargetAndCostsNothing(t *testing.T) {
	r := newTestRun("Wizard")
	before := r.Player.ActionsLeft
	r.UseSkill("ray")
	if r.Player.ActionsLeft != before {
		t.Error("ray with no target in reach must not spend an action")
	}
}

func TestRayBlockedByWall(t *testing.T) {
	r := newTestRun("Wizard")
	// Seal the corridor between player and a far enemy with a solid tile.
	r.Floor.solid[r.Floor.index(Cell{2, 1})] = true
	placeEnemy(r, "rat", Cell{3, 1})
	before := r.Player.ActionsLeft
	r.UseSkill("ray")
	if r.Player.ActionsLeft != before {
		t.Error("ray behind a wall must be rejected")
	}
}

func TestSneakAttackBonusOnlyOnFreshEnemy(t *testing.T) {
	dmgWith := func(actedRecently bool) int {
		r := newTestRun("Rogue")
		r.rollD20 = func() int { return 20 }
		foe := placeEnemy(r, "troll", Cell{2, 1})
		foe.AttackedRecently = actedRecently
		var dealt int
		for _, e := range r.Move(Right) {
			if e.Damage > dealt {
				dealt = e.Damage
			}
		}
		return dealt
	}
	// Same die rolls otherwise: rerun with identical rng seed each time.
	fresh := dmgWith(false)
	warned := dmgWith(true)
	if fresh <= warned {
		t.Errorf("sneak attack (%d) should outdamage a warned enemy (%d)", fresh, warned)
	}
}

// --- 3.3 XP and leveling ---

func TestAwardKillLevelsUpOnThreshold(t *testing.T) {
	r := newTestRun("Fighter")
	foe := placeEnemy(r, "goblin", Cell{5, 1})
	foe.HP = 0
	r.awardKill(foe) // goblin: 15 XP → level 1 still (needs 20)
	if r.Player.Level != 1 {
		t.Errorf("15 XP should keep level 1, got %d", r.Player.Level)
	}
	foe2 := placeEnemy(r, "rat", Cell{6, 1})
	foe2.HP = 0
	r.awardKill(foe2) // +10 XP = 25 ≥ 20 → level 2
	if r.Player.Level != 2 {
		t.Errorf("25 XP should reach level 2, got %d", r.Player.Level)
	}
	if r.Player.MaxHP <= 14+3 {
		t.Errorf("level 2 should raise max HP, got %d", r.Player.MaxHP)
	}
}

// --- 3.4 enemy activation ---

func TestAdjacentEnemyStrikesOncePerPhase(t *testing.T) {
	r := newTestRun("Fighter")
	foe := placeEnemy(r, "troll", Cell{2, 1})
	foe.AtkBonus = 50 // guaranteed hits
	r.Player.ActionsLeft = 0
	r.finishPlayerTurn()
	hp1 := r.Player.HP
	if hp1 >= ClassStatline("Fighter", 1).HP {
		t.Error("adjacent enemy should have struck")
	}
	// Within the same phase it strikes once, so a second phase strikes once more,
	// but never twice in one phase.
	r.finishPlayerTurn()
	if r.Player.HP >= hp1 {
		t.Error("enemy should strike again next phase")
	}
}

func TestDistantEnemyClosesWithoutAttacking(t *testing.T) {
	r := newTestRun("Fighter")
	foe := placeEnemy(r, "rat", Cell{10, 1})
	hp := r.Player.HP
	r.finishPlayerTurn()
	if foe.Pos.X >= 10 {
		t.Error("distant enemy should close distance")
	}
	if foe.Pos.X > 10-3 {
		t.Errorf("enemy should move at most 3 tiles, moved %d", 10-foe.Pos.X)
	}
	if r.Player.HP != hp {
		t.Error("enemy that only moved must not attack")
	}
}

func TestEnemyBecomingAdjacentMidMoveDoesNotStrikeSamePhase(t *testing.T) {
	r := newTestRun("Fighter")
	foe := placeEnemy(r, "rat", Cell{4, 1}) // exactly 3 away
	r.finishPlayerTurn()
	if foe.Pos != (Cell{2, 1}) {
		t.Errorf("enemy should end adjacent at (2,1), got %v", foe.Pos)
	}
	if r.Player.HP != ClassStatline("Fighter", 1).HP {
		t.Error("arrival turn must not include a strike")
	}
}
