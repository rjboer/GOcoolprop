package compare

import "testing"

func TestCompareClassifiesNumericPhaseAndErrorOutcomes(t *testing.T) {
	pass := Compare(10, "", "liquid", 10.00000001, "", "liquid", Tolerance{Absolute: 1e-6, Relative: 1e-8})
	if pass.Outcome != OutcomePassed {
		t.Fatalf("pass outcome = %+v", pass)
	}
	phase := Compare(10, "", "liquid", 10, "", "gas", Tolerance{Absolute: 1e-9, Relative: 1e-9})
	if phase.Outcome != OutcomeFailedPhase {
		t.Fatalf("phase outcome = %+v", phase)
	}
	errorMismatch := Compare(0, "", "", 0, "out of range", "", Tolerance{})
	if errorMismatch.Outcome != OutcomeErrorMismatch {
		t.Fatalf("error mismatch outcome = %+v", errorMismatch)
	}
	consistent := Compare(0, "out of range", "", 0, "out of range", "", Tolerance{})
	if consistent.Outcome != OutcomeConsistentError {
		t.Fatalf("consistent outcome = %+v", consistent)
	}
}

func TestToleranceForUsesStageAndPropertyPolicy(t *testing.T) {
	if got := ToleranceFor("Dmass", "screen/PT"); got.Relative != 1e-8 {
		t.Fatalf("P-T tolerance = %+v", got)
	}
	if got := ToleranceFor("V", "screen/PT"); got.Relative != 1e-6 {
		t.Fatalf("transport tolerance = %+v", got)
	}
	if got := ToleranceFor("Dmass", "eos/TD"); got.Relative != 1e-9 {
		t.Fatalf("EOS tolerance = %+v", got)
	}
}
