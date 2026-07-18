package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
)

// FlashPH solves for temperature and density given pressure and molar enthalpy.
// Only stable single-phase states and saturation endpoints are supported.
func FlashPH(fluidData *fluid.FluidData, P_target, H_target float64) (float64, float64, error) {
	phase, err := inferPhaseFromSaturationAtPressure(fluidData, P_target, H_target, func(s *core.State) float64 {
		return s.MolarEnthalpy()
	}, "P-H")
	if err != nil {
		return 0, 0, err
	}
	return flashPressureProperty(fluidData, P_target, H_target, phase, func(s *core.State) float64 {
		return s.MolarEnthalpy()
	}, "P-H")
}
