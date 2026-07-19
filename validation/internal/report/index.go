package report

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"GOcoolprop/validation/internal/stats"
)

type Fluid struct {
	Name, Status, Report  string
	Total, Passed, Failed int
}

func Index(path string, runID string, fluids []Fluid) error {
	s := fmt.Sprintf("# CoolProp Validation Run %s\n\n| Fluid | Status | Mandatory | Passed | Failed | Failure %% | Report |\n|---|---|---:|---:|---:|---:|---|\n", runID)
	for _, f := range fluids {
		rate := 0.0
		if f.Total > 0 {
			rate = 100 * float64(f.Failed) / float64(f.Total)
		}
		s += fmt.Sprintf("| %s | %s | %d | %d | %d | %.6f | [%s](%s) |\n", f.Name, f.Status, f.Total, f.Passed, f.Failed, rate, f.Report, filepath.ToSlash(f.Report))
	}
	return os.WriteFile(path, []byte(s), 0644)
}

func IndexWithPlan(path string, runID string, fluids []Fluid, plan stats.Plan) error {
	s := fmt.Sprintf("# CoolProp Validation Run %s\n\n", runID)
	s += "## Fluid screening results\n\n"
	s += "| Fluid | Status | Mandatory | Passed | Failed | Failure % | Report |\n|---|---|---:|---:|---:|---:|---|\n"
	for _, f := range fluids {
		rate := 0.0
		if f.Total > 0 {
			rate = 100 * float64(f.Failed) / float64(f.Total)
		}
		s += fmt.Sprintf("| %s | %s | %d | %d | %d | %.6f | [%s](%s) |\n", f.Name, f.Status, f.Total, f.Passed, f.Failed, rate, f.Report, filepath.ToSlash(f.Report))
	}
	s += "\n## Statistical validation budget\n\n"
	s += fmt.Sprintf("Statistical families are defined as fluid × suite × input pair × output property. Confidence is family-wise Bonferroni-adjusted, deterministic grids are not counted as independent audit samples, and zero unexplained failures are required for acceptance.\n\n| Setting | Value |\n|---|---:|\n| Confidence | %.4f |\n| Detectable failure prevalence | %.6f |\n| Family-wise alpha | %.12g |\n| Families | %d |\n| Required independent samples | %d |\n| Minimum samples per family | %d |\n| Safety multiplier | %.2fx |\n\n", plan.Confidence, plan.DetectableFailurePrevalence, plan.FamilyAlpha, plan.FamilyCount, plan.RequiredSamples, plan.MinimumSamples, plan.SafetyMultiplier)
	s += "The current executable records the screening smoke matrix separately. Statistical acceptance remains incomplete until every listed family reaches its independent valid-sample quota.\n\n"
	s += "| Suite | Families | Required independent samples | Status |\n|---|---:|---:|---|\n"
	counts := map[string]int{}
	for _, family := range plan.Families {
		counts[family.Suite]++
	}
	suites := make([]string, 0, len(counts))
	for suite := range counts {
		suites = append(suites, suite)
	}
	sort.Strings(suites)
	for _, suite := range suites {
		familyCount := counts[suite]
		s += fmt.Sprintf("| %s | %d | %d | planned_not_executed |\n", suite, familyCount, familyCount*int(math.Ceil(float64(plan.RequiredSamples)/float64(plan.FamilyCount))))
	}
	return os.WriteFile(path, []byte(s), 0644)
}
