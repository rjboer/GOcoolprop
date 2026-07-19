package stats

import (
	"fmt"
	"math"
	"sort"
)

type Family struct {
	ID          string `json:"id"`
	Fluid       string `json:"fluid"`
	Suite       string `json:"suite"`
	InputPair   string `json:"input_pair"`
	Output      string `json:"output"`
	Required    int    `json:"required"`
	Planned     int    `json:"planned"`
	Valid       int    `json:"valid"`
	Failed      int    `json:"failed"`
	Unsupported bool   `json:"unsupported"`
}

type Plan struct {
	Confidence                  float64  `json:"confidence"`
	DetectableFailurePrevalence float64  `json:"detectable_failure_prevalence"`
	FamilyAlpha                 float64  `json:"family_alpha"`
	FamilyCount                 int      `json:"family_count"`
	MinimumSamples              int      `json:"minimum_samples"`
	SafetyMultiplier            float64  `json:"safety_multiplier"`
	BaseRequiredSamples         int      `json:"base_required_samples"`
	RequiredSamples             int      `json:"required_samples"`
	Families                    []Family `json:"families"`
}

func RequiredZeroFailureSamples(confidence, prevalence float64, familyCount int) (int, float64, error) {
	if confidence <= 0 || confidence >= 1 {
		return 0, 0, fmt.Errorf("confidence must be between 0 and 1")
	}
	if prevalence <= 0 || prevalence >= 1 {
		return 0, 0, fmt.Errorf("detectable prevalence must be between 0 and 1")
	}
	if familyCount < 1 {
		return 0, 0, fmt.Errorf("family count must be positive")
	}
	alpha := (1 - confidence) / float64(familyCount)
	n := int(math.Ceil(math.Log(alpha) / math.Log1p(-prevalence)))
	return n, alpha, nil
}

func BuildPlan(fluids, suites, inputPairs, outputs []string, minimumSamples int, confidence, prevalence float64) (Plan, error) {
	if minimumSamples < 1 {
		return Plan{}, fmt.Errorf("minimum samples must be positive")
	}
	count := len(fluids) * len(suites) * len(inputPairs) * len(outputs)
	if count < 1 {
		return Plan{}, fmt.Errorf("statistical family inventory is empty")
	}
	n, alpha, err := RequiredZeroFailureSamples(confidence, prevalence, count)
	if err != nil {
		return Plan{}, err
	}
	baseN := n
	if n < minimumSamples {
		n = minimumSamples
	}
	const safetyMultiplier = 1.5
	n = int(math.Ceil(float64(n) * safetyMultiplier))
	families := make([]Family, 0, count)
	for _, fluid := range fluids {
		for _, suite := range suites {
			for _, pair := range inputPairs {
				for _, output := range outputs {
					families = append(families, Family{ID: fmt.Sprintf("%s/%s/%s/%s", fluid, suite, pair, output), Fluid: fluid, Suite: suite, InputPair: pair, Output: output, Required: n, Planned: n})
				}
			}
		}
	}
	sort.Slice(families, func(i, j int) bool { return families[i].ID < families[j].ID })
	return Plan{Confidence: confidence, DetectableFailurePrevalence: prevalence, FamilyAlpha: alpha, FamilyCount: count, MinimumSamples: minimumSamples, SafetyMultiplier: safetyMultiplier, BaseRequiredSamples: baseN, RequiredSamples: count * n, Families: families}, nil
}

func AcceptsFamily(f Family) bool {
	return !f.Unsupported && f.Required > 0 && f.Valid >= f.Required && f.Failed == 0
}

// IntegerFollowUp returns every planned ordinal for a failed family. The caller
// uses these ordinals to replay the complete deterministic family, not a sample.
func IntegerFollowUp(familyID string, planned int) []int {
	if planned <= 0 {
		return nil
	}
	ordinals := make([]int, planned)
	for i := range ordinals {
		ordinals[i] = i
	}
	return ordinals
}
