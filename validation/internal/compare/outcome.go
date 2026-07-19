package compare

import "strings"

const (
	OutcomePassed          = "passed"
	OutcomeFailedNumeric   = "failed_numeric"
	OutcomeFailedPhase     = "failed_phase"
	OutcomeConsistentError = "consistent_error"
	OutcomeErrorMismatch   = "error_mismatch"
	OutcomeUnsupported     = "unsupported"
	OutcomePanic           = "panic"
	OutcomeTimeout         = "timeout"
	OutcomeValidatorError  = "validator_error"
)

type CaseResult struct {
	Outcome                string
	Metric                 MetricResult
	Tolerance              Tolerance
	CandidateErrorCategory string
	ReferenceErrorCategory string
	CandidatePhase         string
	ReferencePhase         string
	Classification         string
}

func Compare(candidate float64, candidateError, candidatePhase string, reference float64, referenceError, referencePhase string, tolerance Tolerance) CaseResult {
	result := CaseResult{
		Tolerance:              tolerance,
		CandidateErrorCategory: NormalizeErrorText(candidateError),
		ReferenceErrorCategory: NormalizeErrorText(referenceError),
		CandidatePhase:         candidatePhase,
		ReferencePhase:         referencePhase,
	}
	candidateFailed := candidateError != ""
	referenceFailed := referenceError != ""
	if candidateFailed || referenceFailed {
		if candidateFailed && referenceFailed && result.CandidateErrorCategory == result.ReferenceErrorCategory {
			result.Outcome = OutcomeConsistentError
			return result
		}
		result.Outcome = OutcomeErrorMismatch
		return result
	}
	result.Metric = Metric(candidate, reference, tolerance)
	if candidatePhase != "" && referencePhase != "" && normalizePhase(candidatePhase) != normalizePhase(referencePhase) {
		result.Outcome = OutcomeFailedPhase
		return result
	}
	if result.Metric.Pass {
		result.Outcome = OutcomePassed
	} else {
		result.Outcome = OutcomeFailedNumeric
		result.Classification = Classify(candidate, reference, tolerance.Absolute)
	}
	return result
}

func normalizePhase(phase string) string {
	p := strings.ToLower(strings.TrimSpace(phase))
	if strings.HasPrefix(p, "supercritical") {
		return "supercritical"
	}
	return p
}

func NormalizeErrorText(err string) string {
	if err == "" {
		return ""
	}
	return NormalizeError(errorString(err))
}

type errorString string

func (e errorString) Error() string { return string(e) }

func ToleranceFor(property, stage string) Tolerance {
	t := Tolerance{Absolute: 1e-9, Relative: 1e-8}
	if stage == "eos/TD" {
		t.Relative = 1e-9
	}
	if stage == "flash" {
		t.Relative = 1e-7
	}
	if stage == "saturation" {
		t.Relative = 1e-7
	}
	for _, transport := range []string{"V", "L", "I", "VISCOSITY", "CONDUCTIVITY", "SURFACE_TENSION"} {
		if property == transport {
			t.Relative = 1e-6
		}
	}
	return t
}
