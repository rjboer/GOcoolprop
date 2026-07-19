package generator

import "testing"

func TestSpecializedGeneratorsAreDeterministicAndStable(t *testing.T) {
	e := Envelope{TMin: 10, TMax: 500, PMin: 100, PMax: 1e8, RhoMin: 0.01, RhoMax: 1000}
	anchors := Anchors("Water", e)
	if len(anchors) < 10 || anchors[0].ID >= anchors[len(anchors)-1].ID {
		t.Fatalf("unexpected anchors: %d", len(anchors))
	}
	sat := Saturation("Water", e, 4)
	if len(sat) != 4*9 || sat[0].Stage != "saturation" {
		t.Fatalf("unexpected saturation cases: %d", len(sat))
	}
	invalid := InvalidInputs("Water", e)
	if len(invalid) < 8 || invalid[0].ID >= invalid[len(invalid)-1].ID {
		t.Fatalf("unexpected invalid cases: %d", len(invalid))
	}
	local := Adaptive("Water", Case{ID: "Water/screen/PT/0001", Fluid: "Water", Stage: "screen", Input1: "T", Input2: "P", Output: "Dmass", Value1: 300, Value2: 100000}, 5)
	if len(local) != 25 || local[0].Stage != "adaptive" {
		t.Fatalf("unexpected adaptive cases: %d", len(local))
	}
}
