package core

import (
	"math"
	"testing"
)

func centralDiff(f func(float64) float64, x, h float64) float64 {
	return (f(x+h) - f(x-h)) / (2 * h)
}

func secondCentralDiff(f func(float64) float64, x, h float64) float64 {
	return (f(x+h) - 2*f(x) + f(x-h)) / (h * h)
}

func almostRel(a, b, tol float64) bool {
	diff := math.Abs(a - b)
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return diff/scale <= tol
}

func validateTermDerivatives(t *testing.T, term HelmholtzTerm, tau, delta float64, tol float64) {
	t.Helper()
	htau := 1e-6 * math.Max(1, math.Abs(tau))
	hdelta := 1e-6 * math.Max(1, math.Abs(delta))

	gotDDelta := term.DDelta(tau, delta)
	wantDDelta := centralDiff(func(x float64) float64 { return term.Term(tau, x) }, delta, hdelta)
	if !almostRel(gotDDelta, wantDDelta, tol) {
		t.Fatalf("DDelta mismatch: got=%g want=%g", gotDDelta, wantDDelta)
	}

	gotDTau := term.DTau(tau, delta)
	wantDTau := centralDiff(func(x float64) float64 { return term.Term(x, delta) }, tau, htau)
	if !almostRel(gotDTau, wantDTau, tol) {
		t.Fatalf("DTau mismatch: got=%g want=%g", gotDTau, wantDTau)
	}

	gotDDelta2 := term.DDelta2(tau, delta)
	wantDDelta2 := secondCentralDiff(func(x float64) float64 { return term.Term(tau, x) }, delta, hdelta)
	if !almostRel(gotDDelta2, wantDDelta2, tol*100) {
		t.Fatalf("DDelta2 mismatch: got=%g want=%g", gotDDelta2, wantDDelta2)
	}

	gotDTau2 := term.DTau2(tau, delta)
	wantDTau2 := secondCentralDiff(func(x float64) float64 { return term.Term(x, delta) }, tau, htau)
	if !almostRel(gotDTau2, wantDTau2, tol*100) {
		t.Fatalf("DTau2 mismatch: got=%g want=%g", gotDTau2, wantDTau2)
	}

	gotMixed := term.DDeltaTau(tau, delta)
	wantMixed := centralDiff(func(x float64) float64 { return term.DTau(tau, x) }, delta, hdelta)
	if !almostRel(gotMixed, wantMixed, tol*20) {
		t.Fatalf("DDeltaTau mismatch: got=%g want=%g", gotMixed, wantMixed)
	}
}

func TestImplementedTermDerivativeValidation(t *testing.T) {
	tests := []struct {
		name  string
		term  HelmholtzTerm
		tau   float64
		delta float64
		tol   float64
	}{
		{"Lead", &IdealGasHelmholtzLead{A1: 1.2, A2: -0.4}, 1.3, 0.8, 1e-6},
		{"LogTau", &IdealGasHelmholtzLogTau{A: 2.5}, 1.3, 0.8, 1e-6},
		{"PlanckEinstein", &IdealGasHelmholtzPlanckEinstein{N: []float64{1.1, -0.3}, T: []float64{2.1, 5.4}}, 1.2, 0.9, 1e-6},
		{"Power", &IdealGasHelmholtzPower{N: []float64{0.1, -0.2, 0.3}, T: []float64{-1, 2, 3}}, 1.2, 0.9, 1e-6},
		{"PlanckEinsteinGeneralized", &IdealGasHelmholtzPlanckEinsteinGeneralized{N: []float64{0.4, 0.2}, Theta: []float64{-2.0, -4.0}, C: []float64{1, 1}, D: []float64{-1, -1}}, 1.2, 0.9, 1e-6},
		{"ResidualPower", &ResidualHelmholtzPower{N: []float64{0.5, -0.1}, D: []float64{1, 2}, T: []float64{0.5, 2}, L: []float64{0, 1}}, 1.1, 0.7, 1e-5},
		{"ResidualGaussian", &ResidualHelmholtzGaussian{N: []float64{0.2}, D: []float64{2}, T: []float64{1.5}, Eta: []float64{1.1}, Epsilon: []float64{0.9}, Beta: []float64{0.8}, Gamma: []float64{1.0}}, 1.2, 0.95, 1e-5},
		{"ResidualNonAnalytic", &ResidualHelmholtzNonAnalytic{N: []float64{-0.1}, A: []float64{3.5}, B: []float64{0.85}, Beta: []float64{0.3}, ABig: []float64{0.32}, C: []float64{28}, DBig: []float64{700}}, 1.2, 1.1, 1e-4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateTermDerivatives(t, tt.term, tt.tau, tt.delta, tt.tol)
		})
	}
}
