package crawler

import "testing"

func TestProfBonus(t *testing.T) {
	if got := ProfBonus(3, true); got != 5 {
		t.Errorf("trained level 3: got %d, want 5 (level+2)", got)
	}
	if got := ProfBonus(1, true); got != 3 {
		t.Errorf("trained level 1: got %d, want 3", got)
	}
	if got := ProfBonus(5, false); got != 0 {
		t.Errorf("untrained: got %d, want 0", got)
	}
}

func TestCheckSuccessAndFailure(t *testing.T) {
	// trained L1 (+3) + ability 2 = +5 vs DC 15: roll 10 ties total 15 → success.
	ok := RunCheck(10, 1, true, 2, 15, false)
	if ok.Degree != Success {
		t.Errorf("total 15 vs DC 15: degree %d, want Success", ok.Degree)
	}
	bad := RunCheck(9, 1, true, 2, 15, false)
	if bad.Degree != Failure {
		t.Errorf("total 14 vs DC 15: degree %d, want Failure", bad.Degree)
	}
}

func TestNat20RaisesDegree(t *testing.T) {
	// Total 20 (roll 20, no bonus) can't reach DC 25, but a natural 20 makes
	// it at least a success...
	raised := RunCheck(20, 0, false, 0, 25, false)
	if raised.Degree != Success {
		t.Errorf("nat 20 vs DC 25: degree %d, want Success", raised.Degree)
	}
	// ...and turns an honest success into a critical success.
	crit := RunCheck(20, 1, true, 1, 10, false)
	if crit.Degree != CriticalSuccess {
		t.Errorf("nat 20 over DC 10: degree %d, want CriticalSuccess", crit.Degree)
	}
}

func TestNat1LowersDegree(t *testing.T) {
	// A huge bonus cannot save a natural 1 from failing...
	fumbled := RunCheck(1, 5, true, 10, 10, false)
	if fumbled.Degree != Failure {
		t.Errorf("nat 1 over DC: degree %d, want Failure", fumbled.Degree)
	}
	// ...and a hopeless total becomes a critical failure.
	crit := RunCheck(1, 0, false, 0, 30, false)
	if crit.Degree != CriticalFailure {
		t.Errorf("nat 1 vs DC 30: degree %d, want CriticalFailure", crit.Degree)
	}
}

func TestTrainedOnlyRejectsUntrained(t *testing.T) {
	res := RunCheck(20, 2, false, 5, 10, true)
	if res.Degree != Failure || !res.Untrained {
		t.Errorf("trained-only attempt untrained: degree %d untrained=%v, want Failure true", res.Degree, res.Untrained)
	}
}
