package crawler

// Degree is the four-step PF2e outcome scale.
type Degree int

const (
	CriticalSuccess Degree = 2
	Success         Degree = 1
	Failure         Degree = -1
	CriticalFailure Degree = -2
)

// ProfBonus returns the PF2e proficiency bonus for a skill: trained adds
// level + 2, untrained adds nothing.
func ProfBonus(level int, trained bool) int {
	if trained {
		return level + 2
	}
	return 0
}

// CheckResult is the outcome of one d20 check against a DC.
type CheckResult struct {
	Roll      int
	Total     int
	DC        int
	Degree    Degree
	Untrained bool // trained-only attempt made untrained (auto-fail)
}

// RunCheck resolves a d20 check from its rolled value: total = roll +
// proficiency + ability modifier, compared against the DC, with the natural
// 20 raising and the natural 1 lowering the degree by one step. A
// trained-only check attempted untrained fails automatically with Untrained
// set. The caller supplies the roll so the engine can source it from the
// run's seeded RNG or from a test.
func RunCheck(roll int, level int, trained bool, abilityMod int, dc int, trainedOnly bool) CheckResult {
	if trainedOnly && !trained {
		return CheckResult{DC: dc, Degree: Failure, Untrained: true}
	}
	total := roll + ProfBonus(level, trained) + abilityMod

	degree := Failure
	if total >= dc {
		degree = Success
	}
	switch roll {
	case 20:
		if degree == Failure {
			degree = Success
		} else {
			degree = CriticalSuccess
		}
	case 1:
		if degree == Success {
			degree = Failure
		} else {
			degree = CriticalFailure
		}
	}
	return CheckResult{Roll: roll, Total: total, DC: dc, Degree: degree}
}
