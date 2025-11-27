package fluid

import (
	"testing"
)

func TestFluidRegistry(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"R134a", "R134a.json"},
		{"r-134a", "R134a.json"},
		{"R-134A", "R134a.json"},
		{"nitrogen", "Nitrogen.json"},
		{"N2", "Nitrogen.json"},
		{"air", "Air.json"},
		{"CO2", "CarbonDioxide.json"},
		{"ammonia", "Ammonia.json"},
		{"NH3", "Ammonia.json"},
		{"water", "Water.json"},
		{"H2O", "Water.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename, err := GetFluidFilename(tt.name)
			if err != nil {
				t.Fatalf("GetFluidFilename(%s) failed: %v", tt.name, err)
			}
			if filename != tt.expected {
				t.Errorf("GetFluidFilename(%s) = %s, expected %s", tt.name, filename, tt.expected)
			}
		})
	}
}

func TestLoadCommonFluids(t *testing.T) {
	fluids := []string{
		"Air",
		"Nitrogen",
		"Oxygen",
		"Water",
		"R134a",
		"R410A",
		"R32",
		"Ammonia",
		"CarbonDioxide",
		"Methane",
		"Propane",
	}

	for _, name := range fluids {
		t.Run(name, func(t *testing.T) {
			fluid, err := LoadFluidByName(name, "../../data")
			if err != nil {
				t.Fatalf("Failed to load %s: %v", name, err)
			}
			if fluid == nil {
				t.Errorf("Loaded fluid %s is nil", name)
			}
			if len(fluid.EOS) == 0 {
				t.Errorf("Fluid %s has no EOS data", name)
			}
		})
	}
}
