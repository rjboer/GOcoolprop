// prop_si_test.go
package props

import (
	"math"
	"testing"
)

// almostEqualRel checks relative difference between a and b.
func almostEqualRel(a, b, relTol float64) bool {
	diff := math.Abs(a - b)
	den := math.Max(math.Abs(a), math.Abs(b))
	if den == 0 {
		return diff == 0
	}
	return diff/den <= relTol
}

func TestPropSI_Water(t *testing.T) {
	// Water at 300 K, 101325 Pa
	// NOTE: In these tests "D" is interpreted as molar density [mol/m^3].

	rho, err := PropSI("D", "T", 300.0, "P", 101325.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(D) for Water failed: %v", err)
	}

	// Liquid water at ~300 K has ~55,500 mol/m^3 (≈1000 kg/m^3 / 0.018 kg/mol)
	rhoExpected := 55500.0
	if !almostEqualRel(rho, rhoExpected, 0.02) { // 2% tolerance
		t.Errorf("Water density mismatch: got %v mol/m^3, expected ~%v mol/m^3", rho, rhoExpected)
	}

	// Check Enthalpy is non-zero and not NaN
	h, err := PropSI("H", "T", 300.0, "P", 101325.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(H) for Water failed: %v", err)
	}
	if h == 0 || math.IsNaN(h) {
		t.Errorf("Water enthalpy looks invalid: %v", h)
	}
}

func TestPropSI_Nitrogen(t *testing.T) {
	// Nitrogen at 300 K, 1 atm (gas)
	// Here we treat D as *molar* density:
	// rho = P / (R T)
	const (
		P = 101325.0
		T = 300.0
		R = 8.314462618 // J/(mol K)
	)

	rho, err := PropSI("D", "T", T, "P", P, "Nitrogen")
	if err != nil {
		t.Fatalf("PropSI(D) for Nitrogen failed: %v", err)
	}

	rhoExpected := P / (R * T)                   // ≈ 40.62 mol/m^3
	if !almostEqualRel(rho, rhoExpected, 0.01) { // 1% tolerance
		t.Errorf("Nitrogen density mismatch: got %v mol/m^3, expected ~%v mol/m^3", rho, rhoExpected)
	}

	// Verify pressure via round-trip: T, rho -> P
	p, err := PropSI("P", "T", T, "D", rho, "Nitrogen")
	if err != nil {
		t.Fatalf("PropSI(P) for Nitrogen failed: %v", err)
	}
	if math.Abs(p-P) > 200.0 { // ±200 Pa tolerance
		t.Errorf("Nitrogen pressure verification failed: got %v Pa, expected ~%v Pa", p, P)
	}
}

func TestPropSI_Hydrogen(t *testing.T) {
	// Hydrogen at 300 K, 1 atm (gas)
	// For an ideal gas, molar density is the same as nitrogen at the same T and P.

	const (
		P = 101325.0
		T = 300.0
		R = 8.314462618
	)

	rho, err := PropSI("D", "T", T, "P", P, "Hydrogen")
	if err != nil {
		t.Fatalf("PropSI(D) for Hydrogen failed: %v", err)
	}

	rhoExpected := P / (R * T)                   // ≈ 40.62 mol/m^3
	if !almostEqualRel(rho, rhoExpected, 0.01) { // 1% tolerance
		t.Errorf("Hydrogen density mismatch: got %v mol/m^3, expected ~%v mol/m^3", rho, rhoExpected)
	}

	// Verify pressure via round-trip: T, rho -> P
	p, err := PropSI("P", "T", T, "D", rho, "Hydrogen")
	if err != nil {
		t.Fatalf("PropSI(P) for Hydrogen failed: %v", err)
	}
	if math.Abs(p-P) > 200.0 {
		t.Errorf("Hydrogen pressure verification failed: got %v Pa, expected ~%v Pa", p, P)
	}
}

func TestPropSI_WaterSaturationAndQuality(t *testing.T) {
	// 1) Saturation temperature at 1 atm: T(P=101325, Q=0) ≈ 373.124 K
	P := 101325.0

	Tsat, err := PropSI("T", "P", P, "Q", 0.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(T) for saturated liquid Water failed: %v", err)
	}

	if math.Abs(Tsat-373.124) > 0.5 {
		t.Errorf("Water Tsat mismatch: got %v K, expected ~373.124 K", Tsat)
	}

	// 2) Saturation pressure at T=300 K, Q=1 (saturated vapor)
	T := 300.0

	Psat, err := PropSI("P", "T", T, "Q", 1.0, "Water")
	if err != nil {
		t.Fatalf("PropSI(P) for saturated vapor Water failed: %v", err)
	}

	// Expected Psat at 300 K is ~3536 Pa (depending a bit on correlation)
	if math.Abs(Psat-3536.0) > 200.0 {
		t.Errorf("Water Psat mismatch: got %v Pa, expected ~3536 Pa", Psat)
	}

	// 3) Check Q output around a mid-quality mixture at 300 K.
	//    We obtain rhoL and rhoV from our own PropSI calls, then
	//    build a 50/50 volume mixture and check that Q ≈ 0.5.
	rhoL, err := PropSI("D", "T", T, "Q", 0.0, "Water") // saturated liquid
	if err != nil {
		t.Fatalf("PropSI(D) for saturated liquid Water failed: %v", err)
	}
	rhoV, err := PropSI("D", "T", T, "Q", 1.0, "Water") // saturated vapor
	if err != nil {
		t.Fatalf("PropSI(D) for saturated vapor Water failed: %v", err)
	}

	if rhoL <= 0 || rhoV <= 0 {
		t.Fatalf("Invalid saturation densities: rhoL=%v, rhoV=%v", rhoL, rhoV)
	}

	vL := 1.0 / rhoL
	vV := 1.0 / rhoV

	// 50/50 volume mixture
	vMix := 0.5*vL + 0.5*vV
	rhoMix := 1.0 / vMix

	Q, err := PropSI("Q", "T", T, "D", rhoMix, "Water")
	if err != nil {
		t.Fatalf("PropSI(Q) for Water mixture failed: %v", err)
	}

	if math.Abs(Q-0.5) > 0.02 {
		t.Errorf("Water quality mismatch: got %v, expected ~0.5", Q)
	}
}
