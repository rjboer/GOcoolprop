package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
)

// FlashPS solves for temperature and density given pressure and molar entropy.
// Only stable single-phase states and saturation endpoints are supported.
func FlashPS(fluidData *fluid.FluidData, P_target, S_target float64) (float64, float64, error) {
	phase, err := inferPhaseFromSaturationAtPressure(fluidData, P_target, S_target, func(s *core.State) float64 {
		return s.MolarEntropy()
	}, "P-S")
	if err != nil {
		return 0, 0, err
	}
	return flashPressureProperty(fluidData, P_target, S_target, phase, func(s *core.State) float64 {
		return s.MolarEntropy()
	}, "P-S")
}
