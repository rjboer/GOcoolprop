package props

import (
	"math"
	"testing"
)

func almostEqualRel(a, b, relTol float64) bool {
	diff := math.Abs(a - b)
	den := math.Max(math.Abs(a), math.Abs(b))
	if den == 0 {
		return diff == 0
	}
	return diff/den <= relTol
}

func TestPropSI_Water(t *testing.T) {
	rho, err := PropSI("D", "T", 300.0, "P", 101325.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(D) for Water failed: %v", err)
	}

	rhoExpected := 55500.0
	if !almostEqualRel(rho, rhoExpected, 0.02) {
		t.Errorf("Water density mismatch: got %v mol/m^3, expected ~%v mol/m^3", rho, rhoExpected)
	}

	h, err := PropSI("H", "T", 300.0, "P", 101325.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(H) for Water failed: %v", err)
	}
	if h == 0 || math.IsNaN(h) {
		t.Errorf("Water enthalpy looks invalid: %v", h)
	}
}

func TestPropSI_Nitrogen(t *testing.T) {
	const (
		P = 101325.0
		T = 300.0
		R = 8.314462618
	)

	rho, err := PropSI("D", "T", T, "P", P, "Nitrogen")
	if err != nil {
		t.Fatalf("PropSI(D) for Nitrogen failed: %v", err)
	}

	rhoExpected := P / (R * T)
	if !almostEqualRel(rho, rhoExpected, 0.01) {
		t.Errorf("Nitrogen density mismatch: got %v mol/m^3, expected ~%v mol/m^3", rho, rhoExpected)
	}

	p, err := PropSI("P", "T", T, "D", rho, "Nitrogen")
	if err != nil {
		t.Fatalf("PropSI(P) for Nitrogen failed: %v", err)
	}
	if math.Abs(p-P) > 200.0 {
		t.Errorf("Nitrogen pressure verification failed: got %v Pa, expected ~%v Pa", p, P)
	}
}

func TestPropSI_Hydrogen(t *testing.T) {
	const (
		P = 101325.0
		T = 300.0
		R = 8.314462618
	)

	rho, err := PropSI("D", "T", T, "P", P, "Hydrogen")
	if err != nil {
		t.Fatalf("PropSI(D) for Hydrogen failed: %v", err)
	}

	rhoExpected := P / (R * T)
	if !almostEqualRel(rho, rhoExpected, 0.01) {
		t.Errorf("Hydrogen density mismatch: got %v mol/m^3, expected ~%v mol/m^3", rho, rhoExpected)
	}

	p, err := PropSI("P", "T", T, "D", rho, "Hydrogen")
	if err != nil {
		t.Fatalf("PropSI(P) for Hydrogen failed: %v", err)
	}
	if math.Abs(p-P) > 200.0 {
		t.Errorf("Hydrogen pressure verification failed: got %v Pa, expected ~%v Pa", p, P)
	}
}

func TestPropSI_WaterSaturationEndpoints(t *testing.T) {
	P := 101325.0

	Tsat, err := PropSI("T", "P", P, "Q", 0.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(T) for saturated liquid Water failed: %v", err)
	}
	if math.Abs(Tsat-373.124) > 0.5 {
		t.Errorf("Water Tsat mismatch: got %v K, expected ~373.124 K", Tsat)
	}

	T := 300.0
	Psat, err := PropSI("P", "T", T, "Q", 1.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(P) for saturated vapor Water failed: %v", err)
	}
	if math.Abs(Psat-3536.0) > 200.0 {
		t.Errorf("Water Psat mismatch: got %v Pa, expected ~3536 Pa", Psat)
	}

	qL, err := PropSI("Q", "T", T, "Q", 0.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(Q) for saturated liquid Water failed: %v", err)
	}
	if qL != 0 {
		t.Errorf("expected saturated liquid endpoint quality 0, got %v", qL)
	}

	qV, err := PropSI("Q", "T", T, "Q", 1.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(Q) for saturated vapor Water failed: %v", err)
	}
	if qV != 1 {
		t.Errorf("expected saturated vapor endpoint quality 1, got %v", qV)
	}
}

func TestPropSI_RejectsInteriorQualityInputs(t *testing.T) {
	if _, err := PropSI("D", "T", 300.0, "Q", 0.5, "Water"); err == nil {
		t.Fatalf("expected T,Q with interior quality to fail")
	}
	if _, err := PropSI("D", "P", 101325.0, "Q", 0.5, "Water"); err == nil {
		t.Fatalf("expected P,Q with interior quality to fail")
	}
}

func TestPropSI_RejectsInteriorQualityOutput(t *testing.T) {
	rhoL, err := PropSI("D", "T", 300.0, "Q", 0.0, "Water")
	if err != nil {
		t.Fatalf("rhoL: %v", err)
	}
	rhoV, err := PropSI("D", "T", 300.0, "Q", 1.0, "Water")
	if err != nil {
		t.Fatalf("rhoV: %v", err)
	}

	vMix := 0.5*(1.0/rhoL) + 0.5*(1.0/rhoV)
	rhoMix := 1.0 / vMix

	if _, err := PropSI("Q", "T", 300.0, "D", rhoMix, "Water"); err == nil {
		t.Fatalf("expected Q evaluation for interior two-phase state to fail")
	}
}
