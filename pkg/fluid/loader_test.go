package fluid

import (
	"testing"
)

func TestLoadWater(t *testing.T) {
	// Assuming running from pkg/fluid, data is at ../../data
	path := "../../data/Water.json"
	fluid, err := LoadFluid(path)
	if err != nil {
		t.Fatalf("Failed to load Water.json: %v", err)
	}

	if fluid.Info.Name != "Water" {
		t.Errorf("Expected name Water, got %s", fluid.Info.Name)
	}

	if len(fluid.EOS) == 0 {
		t.Fatal("No EOS found")
	}

	eos := fluid.EOS[0]
	if eos.MolarMass == 0 {
		t.Error("Molar mass is 0")
	}

	if len(eos.Alpha0) == 0 {
		t.Error("No Alpha0 terms")
	}

	if len(eos.AlphaR) == 0 {
		t.Error("No AlphaR terms")
	}
}
