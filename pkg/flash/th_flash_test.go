package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/saturation"
	"math"
	"testing"
)

func TestFlashTH_Nitrogen_Gas(t *testing.T) {
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	T := 300.0
	rhoExpected := 40.6

	state, err := core.NewState(f)
	if err != nil {
		t.Fatalf("Failed to build state: %v", err)
	}
	state.Update(T, rhoExpected)
	H_target := state.MolarEnthalpy()

	rhoResult, err := FlashTH(f, T, H_target)
	if err != nil {
		t.Fatalf("FlashTH failed: %v", err)
	}
	if math.Abs(rhoResult-rhoExpected)/rhoExpected > 0.01 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoResult, rhoExpected)
	}
}

func TestFlashTH_Water_SaturatedLiquidEndpoint(t *testing.T) {
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	T := 300.0
	rhoExpected, err := saturation.RhoL(f, T)
	if err != nil {
		t.Fatalf("RhoL failed: %v", err)
	}

	state, err := core.NewState(f)
	if err != nil {
		t.Fatalf("Failed to build state: %v", err)
	}
	state.Update(T, rhoExpected)
	H_target := state.MolarEnthalpy()

	rhoResult, err := FlashTH(f, T, H_target)
	if err != nil {
		t.Fatalf("FlashTH failed: %v", err)
	}
	if math.Abs(rhoResult-rhoExpected)/rhoExpected > 1e-4 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoResult, rhoExpected)
	}
}

func TestFlashTH_Hydrogen_Gas(t *testing.T) {
	f, err := fluid.LoadFluidByName("Hydrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Hydrogen: %v", err)
	}

	T := 300.0
	rhoExpected := 40.6

	state, err := core.NewState(f)
	if err != nil {
		t.Fatalf("Failed to build state: %v", err)
	}
	state.Update(T, rhoExpected)
	H_target := state.MolarEnthalpy()

	rhoResult, err := FlashTH(f, T, H_target)
	if err != nil {
		t.Fatalf("FlashTH failed: %v", err)
	}
	if math.Abs(rhoResult-rhoExpected)/rhoExpected > 0.01 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoResult, rhoExpected)
	}
}
