package flash

import (
	"GOcoolprop/pkg/fluid"
	"testing"
)

func TestDensityTPTreatsHighPressureSubcriticalStatesAsLiquid(t *testing.T) {
	cases := []struct {
		fluidName string
		T         float64
		P         float64
	}{
		{fluidName: "Water", T: 273.16172684000003, P: 22064000.000000026},
		{fluidName: "Hydrogen", T: 13.957986043, P: 1296400.000000001},
	}

	for _, tc := range cases {
		t.Run(tc.fluidName, func(t *testing.T) {
			f, err := fluid.LoadFluidByName(tc.fluidName, "../../data")
			if err != nil {
				t.Fatal(err)
			}
			phase, err := inferPhaseTP(f, tc.T, tc.P)
			if err != nil {
				t.Fatalf("phase inference failed: %v", err)
			}
			if phase != phaseLiquid {
				t.Fatalf("expected liquid phase preference, got %v", phase)
			}
			rho, err := DensityTP(f, tc.T, tc.P)
			if err != nil {
				t.Fatal(err)
			}
			if rho <= f.States.Critical.RhoMolar {
				t.Fatalf("expected compressed-liquid density above critical density: got %g critical=%g", rho, f.States.Critical.RhoMolar)
			}
		})
	}
}
