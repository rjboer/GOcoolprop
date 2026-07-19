package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest identifies the exact software, environment, and inputs used by a run.
type Manifest struct {
	RunID            string   `json:"run_id"`
	StartedAt        string   `json:"started_at,omitempty"`
	CompletedAt      string   `json:"completed_at,omitempty"`
	GoVersion        string   `json:"go_version"`
	PythonVersion    string   `json:"python_version"`
	CoolPropVersion  string   `json:"coolprop_version"`
	CoolPropGit      string   `json:"coolprop_git_revision,omitempty"`
	Backend          string   `json:"backend"`
	ReferenceState   string   `json:"reference_state"`
	PythonExecutable string   `json:"python_executable"`
	Seed             int64    `json:"seed"`
	CandidateWorkers int      `json:"candidate_workers"`
	ReferenceWorkers int      `json:"reference_workers"`
	BatchSize        int      `json:"batch_size"`
	Fluids           []string `json:"fluids"`
	GitRevision      string   `json:"candidate_git_revision,omitempty"`
	GitDirty         bool     `json:"candidate_git_dirty"`
	OS               string   `json:"os"`
	Arch             string   `json:"arch"`
}

// ConfigSnapshot is the resolved, run-effective validator configuration.
type ConfigSnapshot struct {
	ReferenceImplementation string
	ReferenceVersion        string
	Backend                 string
	ReferenceState          string
	ReferenceWorkers        int
	BatchSize               int
	CandidateWorkers        int
	ActiveFluids            int
	Seed                    int64
	TDTemperaturePoints     int
	TDDensityPoints         int
	PTTemperaturePoints     int
	PTPressurePoints        int
	QuasiRandomPoints       int
	StatisticsEnabled       bool
	Confidence              float64
	DetectablePrevalence    float64
	MinimumAuditSamples     int
	FamilyWiseAdjustment    string
	SafetyMultiplier        float64
}

func WriteManifest(path string, manifest Manifest) error {
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(b, '\n'))
}

// WriteResolvedConfig writes a stable YAML subset without adding a dependency.
func WriteResolvedConfig(path string, c ConfigSnapshot) error {
	content := fmt.Sprintf(`reference:
  implementation: %s
  version: "%s"
  backend: %s
  reference_state: %s
  worker_processes: %d
  batch_size: %d
execution:
  candidate_workers: %d
  active_fluids: %d
  seed: %d
screening:
  td_temperature_points: %d
  td_density_points: %d
  pt_temperature_points: %d
  pt_pressure_points: %d
  low_discrepancy_points: %d
statistics:
  enabled: %t
  confidence: %g
  detectable_failure_prevalence: %g
  minimum_valid_audit_samples: %d
  family_wise_adjustment: %s
  safety_multiplier: %g
`, c.ReferenceImplementation, c.ReferenceVersion, c.Backend, c.ReferenceState,
		c.ReferenceWorkers, c.BatchSize, c.CandidateWorkers, c.ActiveFluids, c.Seed,
		c.TDTemperaturePoints, c.TDDensityPoints, c.PTTemperaturePoints, c.PTPressurePoints,
		c.QuasiRandomPoints, c.StatisticsEnabled, c.Confidence, c.DetectablePrevalence,
		c.MinimumAuditSamples, c.FamilyWiseAdjustment, c.SafetyMultiplier)
	return writeFile(path, []byte(content))
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
