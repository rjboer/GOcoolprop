package saturation

import (
	"GOcoolprop/pkg/fluid"
	"math"
	"testing"
)

func TestSaturation_Water(t *testing.T) {
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	// Test Psat at 373.15 K (100 C) -> Should be ~101325 Pa
	T := 373.15
	P_expected := 101325.0

	P_calc, err := Psat(f, T)
	if err != nil {
		t.Fatalf("Psat failed: %v", err)
	}

	t.Logf("Water Psat at %v K: %v Pa (Expected ~%v)", T, P_calc, P_expected)

	if math.Abs(P_calc-P_expected)/P_expected > 0.01 { // 1% tolerance
		t.Errorf("Psat mismatch: got %v, expected %v", P_calc, P_expected)
	}

	// Test Tsat at 101325 Pa -> Should be ~373.15 K
	T_calc, err := Tsat(f, P_expected)
	if err != nil {
		t.Fatalf("Tsat failed: %v", err)
	}

	t.Logf("Water Tsat at %v Pa: %v K (Expected ~%v)", P_expected, T_calc, T)

	if math.Abs(T_calc-T) > 0.1 {
		t.Errorf("Tsat mismatch: got %v, expected %v", T_calc, T)
	}

	// Test RhoL at 300 K -> Should be ~997 kg/m3 / 0.018 kg/mol = 55388 mol/m3
	T = 300.0
	RhoL_expected := 55388.0

	RhoL_calc, err := RhoL(f, T)
	if err != nil {
		t.Fatalf("RhoL failed: %v", err)
	}

	t.Logf("Water RhoL at %v K: %v mol/m3 (Expected ~%v)", T, RhoL_calc, RhoL_expected)

	if math.Abs(RhoL_calc-RhoL_expected)/RhoL_expected > 0.01 {
		t.Errorf("RhoL mismatch: got %v, expected %v", RhoL_calc, RhoL_expected)
	}
}

func TestSaturation_Nitrogen(t *testing.T) {
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	// Test Tsat at 101325 Pa -> Should be ~77.35 K
	P := 101325.0
	T_expected := 77.35

	T_calc, err := Tsat(f, P)
	if err != nil {
		t.Fatalf("Tsat failed: %v", err)
	}

	t.Logf("Nitrogen Tsat at %v Pa: %v K (Expected ~%v)", P, T_calc, T_expected)

	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Tsat mismatch: got %v, expected %v", T_calc, T_expected)
	}
}
