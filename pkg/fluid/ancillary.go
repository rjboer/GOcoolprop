package fluid

import (
	"math"
)

// Evaluate calculates the value of the ancillary curve at temperature T.
// Supported types: "pV", "pL", "rhoV", "rhoLnoexp"
func (ac *AncillaryCurve) Evaluate(T float64) float64 {
	// Check bounds (optional, but good practice)
	// if T < ac.TMin || T > ac.TMax { ... }

	// Reducing temperature
	Tc := ac.TR
	if Tc == 0 {
		// Fallback if T_r is not explicitly set (should not happen for valid JSON)
		Tc = ac.TMax
	}

	theta := 1.0 - T/Tc

	// Calculate sum(n_i * theta^t_i)
	sum := 0.0
	for i := range ac.N {
		sum += ac.N[i] * math.Pow(theta, ac.T[i])
	}

	switch ac.Type {
	case "pV", "pL":
		// p = pc * exp( (Tc/T) * sum )
		return ac.ReducingValue * math.Exp((Tc/T)*sum)

	case "rhoV":
		// rho = rhoc * exp( (Tc/T) * sum )
		return ac.ReducingValue * math.Exp((Tc/T)*sum)

	case "rhoLnoexp":
		// rho = rhoc * (1 + sum)
		return ac.ReducingValue * (1.0 + sum)

	default:
		// Unknown type or not implemented
		return 0.0
	}
}
