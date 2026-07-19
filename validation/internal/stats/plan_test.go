package stats

import (
	"math"
	"testing"
)

func TestRequiredZeroFailureSamplesUsesFamilyWiseConfidence(t *testing.T) {
	n, alpha, err := RequiredZeroFailureSamples(0.99, 0.001, 1)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4603 || math.Abs(alpha-0.01) > 1e-12 {
		t.Fatalf("single-family budget: n=%d alpha=%g", n, alpha)
	}
	n, alpha, err = RequiredZeroFailureSamples(0.99, 0.001, 100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 9206 || math.Abs(alpha-0.0001) > 1e-12 {
		t.Fatalf("Bonferroni budget: n=%d alpha=%g", n, alpha)
	}
}

func TestBuildPlanAppliesMinimumAndTracksEveryFamily(t *testing.T) {
	p, err := BuildPlan([]string{"Water", "Nitrogen"}, []string{"pt", "saturation"}, []string{"T,P"}, []string{"Dmass", "Hmass"}, 5000, 0.99, 0.001)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Families) != 8 || p.RequiredSamples != 8*10023 || p.SafetyMultiplier != 1.5 {
		t.Fatalf("plan size: families=%d required=%d", len(p.Families), p.RequiredSamples)
	}
	if p.Families[0].Required != 10023 {
		t.Fatalf("family budget=%d", p.Families[0].Required)
	}
	if !AcceptsFamily(Family{Required: 5000, Valid: 5000}) || AcceptsFamily(Family{Required: 5000, Valid: 5000, Failed: 1}) {
		t.Fatal("unexpected family acceptance")
	}
}

func TestIntegerFollowUpEnumeratesEveryOrdinalExactlyOnce(t *testing.T) {
	ordinals := IntegerFollowUp("Water/pt/T,P/Dmass", 7)
	if len(ordinals) != 7 || ordinals[0] != 0 || ordinals[6] != 6 {
		t.Fatalf("unexpected integer follow-up: %v", ordinals)
	}
}

func TestBuildPlanRejectsInvalidStatisticalParameters(t *testing.T) {
	for _, tc := range [][2]float64{{0, 0.001}, {0.99, 0}, {0.99, 1}} {
		if _, err := BuildPlan([]string{"Water"}, []string{"pt"}, []string{"T,P"}, []string{"P"}, 5000, tc[0], tc[1]); err == nil {
			t.Fatalf("expected invalid parameters for %+v", tc)
		}
	}
	if _, err := BuildPlan([]string{"Water"}, nil, []string{"T,P"}, []string{"P"}, 5000, 0.99, 0.001); err == nil {
		t.Fatal("expected empty family error")
	}
}
