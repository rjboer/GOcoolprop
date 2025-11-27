package core

import (
	"GOcoolprop/pkg/fluid"
	"math"
	"testing"
)

func TestWaterCore(t *testing.T) {
	// Load Water
	f, err := fluid.LoadFluid("../../data/Water.json")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	state := NewState(f)

	var Temp, RhoMolar, P, ExpectedP, R, RhoIdeal, P_Ideal float64

	// Test Triple Point Liquid
	// T=273.16, Rho=55496.95514 -> P=611.655
	Temp = 273.16
	RhoMolar = 55496.95514

	state.Update(Temp, RhoMolar)
	P = state.Pressure()
	ExpectedP = 611.655

	if math.Abs(P-ExpectedP) > 10 {
		t.Errorf("Triple point pressure mismatch: got %v, expected %v", P, ExpectedP)
	}

	// Test Low Pressure Vapor (Ideal Gas-like)
	// T=300, P=1000 Pa.
	// Rho approx P/RT = 1000 / (8.314 * 300) = 0.4009
	Temp = 300.0
	ExpectedP = 1000.0
	R = f.EOS[0].GasConstant
	RhoIdeal = ExpectedP / (R * Temp)

	state.Update(Temp, RhoIdeal)
	P_Ideal = state.Pressure()

	if math.Abs(P_Ideal-ExpectedP) > 10 {
		t.Errorf("Low pressure vapor mismatch: got %v, expected %v", P_Ideal, ExpectedP)
	}

	// Test Liquid at 300K, 1 atm
	Temp = 300.0
	RhoMolar = 55317.3
	state.Update(Temp, RhoMolar)
	P = state.Pressure()
	ExpectedP = 101325.0

	if math.Abs(P-ExpectedP) > 3000 { // 3% tolerance
		t.Errorf("Liquid pressure mismatch: got %v, expected %v", P, ExpectedP)
	}
}

func TestNitrogenCore(t *testing.T) {
	f, err := fluid.LoadFluid("../../data/Nitrogen.json")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	state := NewState(f)

	// T=300, P=101325. Expected Rho=40.6
	Temp := 300.0
	RhoMolar := 40.6

	state.Update(Temp, RhoMolar)
	P := state.Pressure()
	ExpectedP := 101325.0

	if math.Abs(P-ExpectedP) > 1000 {
		t.Errorf("Nitrogen Gas pressure mismatch: got %v, expected %v", P, ExpectedP)
	}

	// Check high density point found by solver
	RhoHigh := 37148.0
	state.Update(Temp, RhoHigh)
	P_High := state.Pressure()
	t.Logf("Pressure at Rho=%v is %v", RhoHigh, P_High)
}
