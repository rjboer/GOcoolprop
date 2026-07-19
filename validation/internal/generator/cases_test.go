package generator

import "testing"

func TestScreeningCasesAreDeterministic(t *testing.T) {
	cfg := Config{TDTemperaturePoints: 3, TDDensityPoints: 3, PTTemperaturePoints: 2, PTPressurePoints: 2, QuasiRandomPoints: 4, Seed: 7}
	a := Screening("Water", Envelope{TMin: 273, TMax: 373, PMin: 1000, PMax: 1e6, RhoMin: 1, RhoMax: 1000}, cfg)
	b := Screening("Water", Envelope{TMin: 273, TMax: 373, PMin: 1000, PMax: 1e6, RhoMin: 1, RhoMax: 1000}, cfg)
	if len(a) != len(b) {
		t.Fatalf("lengths %d and %d", len(a), len(b))
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Value1 != b[i].Value1 || a[i].Value2 != b[i].Value2 {
			t.Fatalf("case %d differs", i)
		}
	}
}
