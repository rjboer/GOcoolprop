package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"math"
	"testing"
)

func TestFlashTH_Nitrogen_Gas(t *testing.T) {
	// Load Nitrogen
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	// Test case: Nitrogen gas at 300K
	T := 300.0
	rhoExpected := 40.6 // mol/m³ (approximately ideal gas at 1 atm)

	// Calculate H at this state
	state := core.NewState(f)
	state.Update(T, rhoExpected)
	H_target := state.MolarEnthalpy()

	t.Logf("Test: T=%v K, rho=%v mol/m³, H=%v J/mol", T, rhoExpected, H_target)

	// Flash back
	rhoResult, err := FlashTH(f, T, H_target)
	if err != nil {
		t.Fatalf("FlashTH failed: %v", err)
	}

	// Check result
	relError := math.Abs(rhoResult-rhoExpected) / rhoExpected
	t.Logf("Result: rho=%v mol/m³, relative error=%v%%", rhoResult, relError*100)

	if relError > 0.01 { // 1% tolerance
		t.Errorf("Density mismatch: got %v, expected %v (error %.2f%%)",
			rhoResult, rhoExpected, relError*100)
	}
}

func TestFlashTH_Water_Liquid(t *testing.T) {
	// Load Water
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	// Test case: Liquid water at 300K, high density
	T := 300.0
	rhoExpected := 55000.0 // mol/m³ (liquid water)

	// Calculate H at this state
	state := core.NewState(f)
	state.Update(T, rhoExpected)
	H_target := state.MolarEnthalpy()

	t.Logf("Test: T=%v K, rho=%v mol/m³, H=%v J/mol", T, rhoExpected, H_target)

	// Flash back
	rhoResult, err := FlashTH(f, T, H_target)
	if err != nil {
		t.Fatalf("FlashTH failed: %v", err)
	}

	// Check result
	relError := math.Abs(rhoResult-rhoExpected) / rhoExpected
	t.Logf("Result: rho=%v mol/m³, relative error=%v%%", rhoResult, relError*100)

	if relError > 0.01 { // 1% tolerance
		t.Errorf("Density mismatch: got %v, expected %v (error %.2f%%)",
			rhoResult, rhoExpected, relError*100)
	}
}

func TestFlashTH_Hydrogen_Gas(t *testing.T) {
	// Load Hydrogen
	f, err := fluid.LoadFluidByName("Hydrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Hydrogen: %v", err)
	}

	// Test case: Hydrogen gas at 300K
	T := 300.0
	rhoExpected := 40.6 // mol/m³ (approximately ideal gas at 1 atm)

	// Calculate H at this state
	state := core.NewState(f)
	state.Update(T, rhoExpected)
	H_target := state.MolarEnthalpy()

	t.Logf("Test: T=%v K, rho=%v mol/m³, H=%v J/mol", T, rhoExpected, H_target)

	// Flash back
	rhoResult, err := FlashTH(f, T, H_target)
	if err != nil {
		t.Fatalf("FlashTH failed: %v", err)
	}

	// Check result
	relError := math.Abs(rhoResult-rhoExpected) / rhoExpected
	t.Logf("Result: rho=%v mol/m³, relative error=%v%%", rhoResult, relError*100)

	if relError > 0.01 { // 1% tolerance
		t.Errorf("Density mismatch: got %v, expected %v (error %.2f%%)",
			rhoResult, rhoExpected, relError*100)
	}
}
