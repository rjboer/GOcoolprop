package validation

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/saturation"
	"math"
	"testing"
)

func relErr(a, b float64) float64 {
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) / scale
}

func buildState(t *testing.T, name string) (*fluid.FluidData, *core.State) {
	t.Helper()
	f, err := fluid.LoadFluidByName(name, "../../data")
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	s, err := core.NewState(f)
	if err != nil {
		t.Fatalf("new state %s: %v", name, err)
	}
	return f, s
}

func TestStructuralComplianceCoreFluids(t *testing.T) {
	for _, name := range []string{"Water", "Nitrogen", "Hydrogen"} {
		_, _ = buildState(t, name)
	}
}

func TestStructuralComplianceUnsupportedFluidFails(t *testing.T) {
	f, err := fluid.LoadFluidByName("CarbonDioxide", "../../data")
	if err != nil {
		t.Fatalf("load CarbonDioxide: %v", err)
	}
	if _, err := core.NewState(f); err == nil {
		t.Fatalf("expected unsupported term failure for CarbonDioxide")
	}
}

func TestStateMathCompliance(t *testing.T) {
	type point struct {
		fluid string
		T     float64
		rho   float64
	}
	points := []point{
		{"Water", 300, 55317.3},
		{"Nitrogen", 300, 40.6},
		{"Hydrogen", 300, 40.6},
	}

	for _, p := range points {
		t.Run(p.fluid, func(t *testing.T) {
			_, s := buildState(t, p.fluid)
			s.Update(p.T, p.rho)

			if relErr(s.MolarEnthalpy(), s.MolarInternalEnergy()+s.Pressure()/s.Rho) > 1e-8 {
				t.Fatalf("enthalpy identity failed")
			}
			if s.Cp() < s.Cv() {
				t.Fatalf("cp < cv")
			}

			dT := 1e-4 * p.T
			drho := 1e-5 * p.rho

			_, s1 := buildState(t, p.fluid)
			s1.Update(p.T+dT, p.rho)
			_, s2 := buildState(t, p.fluid)
			s2.Update(p.T-dT, p.rho)
			if relErr(s.DPdT(), (s1.Pressure()-s2.Pressure())/(2*dT)) > 5e-4 {
				t.Fatalf("DPdT mismatch")
			}
			if relErr(s.DHdT(), (s1.MolarEnthalpy()-s2.MolarEnthalpy())/(2*dT)) > 5e-4 {
				t.Fatalf("DHdT mismatch")
			}
			if relErr(s.DSdT(), (s1.MolarEntropy()-s2.MolarEntropy())/(2*dT)) > 5e-4 {
				t.Fatalf("DSdT mismatch")
			}

			_, s3 := buildState(t, p.fluid)
			s3.Update(p.T, p.rho+drho)
			_, s4 := buildState(t, p.fluid)
			s4.Update(p.T, p.rho-drho)
			if relErr(s.DPdRho(), (s3.Pressure()-s4.Pressure())/(2*drho)) > 5e-4 {
				t.Fatalf("DPdRho mismatch")
			}
			if relErr(s.DHdRho(), (s3.MolarEnthalpy()-s4.MolarEnthalpy())/(2*drho)) > 5e-4 {
				t.Fatalf("DHdRho mismatch")
			}
			if relErr(s.DSdRho(), (s3.MolarEntropy()-s4.MolarEntropy())/(2*drho)) > 5e-4 {
				t.Fatalf("DSdRho mismatch")
			}
		})
	}
}

func TestCoreFluidReferencePoints(t *testing.T) {
	type refPoint struct {
		name  string
		point func(*fluid.FluidData) fluid.StatePoint
	}
	tests := []refPoint{
		{"Water-reducing", func(f *fluid.FluidData) fluid.StatePoint { return f.EOS[0].States.Reducing }},
		{"Water-critical", func(f *fluid.FluidData) fluid.StatePoint { return f.States.Critical }},
		{"Nitrogen-reducing", func(f *fluid.FluidData) fluid.StatePoint { return f.EOS[0].States.Reducing }},
		{"Hydrogen-reducing", func(f *fluid.FluidData) fluid.StatePoint { return f.EOS[0].States.Reducing }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fluidName := tt.name[:len(tt.name)-len("-reducing")]
			if tt.name == "Water-critical" {
				fluidName = "Water"
			}
			f, s := buildState(t, fluidName)
			sp := tt.point(f)
			s.Update(sp.T, sp.RhoMolar)

			if relErr(s.Pressure(), sp.P) > 1e-3 {
				t.Fatalf("pressure mismatch: got=%g want=%g", s.Pressure(), sp.P)
			}
			if relErr(s.MolarEnthalpy(), sp.HMolar) > 1e-3 {
				t.Fatalf("enthalpy mismatch: got=%g want=%g", s.MolarEnthalpy(), sp.HMolar)
			}
			if relErr(s.MolarEntropy(), sp.SMolar) > 1e-3 {
				t.Fatalf("entropy mismatch: got=%g want=%g", s.MolarEntropy(), sp.SMolar)
			}
		})
	}
}

func TestSaturationReferencePoints(t *testing.T) {
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("load Water: %v", err)
	}
	p, err := saturation.Psat(f, 373.1243)
	if err != nil {
		t.Fatalf("Psat: %v", err)
	}
	if relErr(p, 101325) > 1e-2 {
		t.Fatalf("water Psat mismatch")
	}
	tsat, err := saturation.Tsat(f, 101325)
	if err != nil {
		t.Fatalf("Tsat: %v", err)
	}
	if math.Abs(tsat-373.1243) > 0.2 {
		t.Fatalf("water Tsat mismatch")
	}
}
