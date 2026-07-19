package candidate

import (
	"context"
	"math"
	"testing"
)

func TestEngineReturnsMassDensityForDmass(t *testing.T) {
	r := Engine{DataDir: "../../../data"}.Evaluate(context.Background(), Request{Fluid: "Water", Output: "Dmass", Input1: "T", Value1: 300, Input2: "P", Value2: 101325})
	if r.Error != "" || r.Value < 900 || r.Value > 1100 {
		t.Fatalf("result=%+v", r)
	}
}

func TestEngineContainsPanics(t *testing.T) {
	r := Engine{DataDir: "data"}.Evaluate(context.Background(), Request{Fluid: "does-not-exist", Output: "P", Input1: "T", Value1: 300, Input2: "P", Value2: 101325})
	if r.Error == "" {
		t.Fatal("expected error")
	}
}

func TestMetadataIncludesTriplePressure(t *testing.T) {
	meta := Engine{DataDir: "../../../data"}.Metadata("Water")
	if meta.Pmin < 600 || meta.Pmin > 700 {
		t.Fatalf("expected Water triple pressure, got %+v", meta)
	}
	if meta.RhoTripleVapor <= 0 || meta.RhoTripleVapor > 1 {
		t.Fatalf("expected Water triple vapor density, got %+v", meta)
	}
}

func TestEngineConvertsMassSpecificThermodynamicOutputs(t *testing.T) {
	e := Engine{DataDir: "../../../data"}
	hMass := e.Evaluate(context.Background(), Request{Fluid: "Water", Output: "Hmass", Input1: "T", Value1: 300, Input2: "P", Value2: 101325})
	hMolar := e.Evaluate(context.Background(), Request{Fluid: "Water", Output: "Hmolar", Input1: "T", Value1: 300, Input2: "P", Value2: 101325})
	if hMass.Error != "" || hMolar.Error != "" {
		t.Fatalf("mass=%+v molar=%+v", hMass, hMolar)
	}
	if math.Abs(hMass.Value*e.Metadata("Water").MolarMass-hMolar.Value) > 1e-6 {
		t.Fatalf("expected mass/molar enthalpy conversion: mass=%g molar=%g", hMass.Value, hMolar.Value)
	}
}

func TestCapabilitiesDeclareUnitsAndCoverageBoundaries(t *testing.T) {
	cap := (Engine{}).Capabilities()
	for _, output := range []string{"Dmass", "Dmolar", "Hmass", "Hmolar", "Smass", "Smolar", "Cpmass", "Cvmolar"} {
		if !contains(cap.Outputs, output) {
			t.Fatalf("missing output capability %q: %+v", output, cap.Outputs)
		}
	}
	if cap.Units["Hmass"] != "J/kg" || cap.Units["Hmolar"] != "J/mol" {
		t.Fatalf("missing enthalpy units: %+v", cap.Units)
	}
	if !cap.Saturation || !cap.Phase {
		t.Fatalf("expected saturation and phase capability: %+v", cap)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
