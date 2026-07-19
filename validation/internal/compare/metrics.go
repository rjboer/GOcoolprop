package compare

import (
	"math"
	"strings"
)

type Tolerance struct{ Absolute, Relative float64 }
type MetricResult struct {
	Absolute, Relative, Normalized float64
	RelativeMeaningful, Pass       bool
}

func Metric(candidate, reference float64, tol Tolerance) MetricResult {
	a := math.Abs(candidate - reference)
	meaningful := math.Abs(reference) > tol.Absolute
	r := 0.0
	if meaningful {
		r = a / math.Abs(reference)
	}
	n := a / (tol.Absolute + tol.Relative*math.Abs(reference))
	return MetricResult{Absolute: a, Relative: r, Normalized: n, RelativeMeaningful: meaningful, Pass: n <= 1}
}

const (
	ClassNone       = ""
	ClassMassMolar  = "mass_molar_confusion"
	ClassUnit       = "unit_conversion"
	ClassReference  = "reference_state_offset"
	ClassEOS        = "direct_eos_deviation"
	ClassSaturation = "saturation_correlation_deviation"
	ClassFlash      = "inverse_flash_convergence_failure"
	ClassPhase      = "phase_selection_error"
	ClassRuntime    = "crash_or_timeout"
	ClassContract   = "out_of_range_contract_mismatch"
	ClassUnknown    = "unknown"
)

func Classify(candidate, reference, tolerance float64) string {
	if reference != 0 && candidate/reference > 40 && candidate/reference < 60 {
		return ClassMassMolar
	}
	if reference != 0 && math.Abs(candidate-reference) > tolerance && math.Abs(candidate-reference) < 1e-6*math.Max(1, math.Abs(reference)) {
		return ClassReference
	}
	return ClassUnknown
}

func NormalizeError(err error) string {
	if err == nil {
		return ""
	}
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "not supported"), strings.Contains(s, "not implemented"):
		return "not_implemented"
	case strings.Contains(s, "out of range"), strings.Contains(s, "valid range"):
		return "out_of_range"
	case strings.Contains(s, "converg"):
		return "no_convergence"
	case strings.Contains(s, "saturation"), strings.Contains(s, "ambig"):
		return "phase_ambiguity"
	default:
		return "internal_error"
	}
}
