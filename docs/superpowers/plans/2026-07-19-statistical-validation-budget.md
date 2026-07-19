# Statistical Validation Budget Implementation Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Calculate and report a statistically defensible independent audit budget for every fluid, validation suite, input pair, and output property, with family-wise confidence control.

**Architecture:** Define a statistical family as `(fluid, suite, input pair, output property)`. Use Bonferroni alpha allocation across all planned families, calculate the exact zero-failure binomial sample size for the configured detectable prevalence, and reject statistical acceptance unless every mandatory family reaches its independent sample quota with zero unexplained failures. Deterministic grids remain coverage tests and are reported separately.

**Tech Stack:** Go standard library, existing validator capability manifest/report writer, deterministic Halton sampling with recorded family seed.

## Global Constraints

- Confidence is family-wise, not a pooled average.
- Deterministic grids do not count as independent statistical samples.
- Every declared fluid, input pair, output property, and statistical suite is represented.
- Unsupported or unexecuted families cannot pass.
- The report must show planned, required, valid, failed, and accepted counts per family.

### Task 1: Statistical budget engine

**Files:** Create `validation/internal/stats/plan.go`, `validation/internal/stats/plan_test.go`.

- [ ] Test Bonferroni alpha allocation and zero-failure sample-size calculation.
- [ ] Test invalid confidence/prevalence/family inputs.
- [ ] Implement `Plan`, `Family`, `BuildPlan`, `RequiredZeroFailureSamples`, and `AcceptsFamily`.
- [ ] Verify `confidence=0.99`, prevalence `0.001`, and minimum `5000` produce a reproducible required count.

### Task 2: Family inventory and deterministic audit cases

**Files:** Create `validation/internal/stats/families.go`; modify `validation/internal/generator/cases.go`; add tests.

- [ ] Build families from discovered fluids and the candidate capability manifest.
- [ ] Include TD, PT, quasi-random, saturation, phase-boundary, critical-boundary, validity-boundary, flash, round-trip, and invalid-input suites.
- [ ] Generate independent family seeds from the run seed and stable family ID.
- [ ] Track valid reference samples separately from invalid/reference-rejected cases.

### Task 3: Statistical reporting and index integration

**Files:** Modify `validation/internal/report/index.go`, `validation/cmd/coolprop-validate/main.go`, `validation/config/validation.yaml`.

- [ ] Add a statistical validation section to `index.md` listing confidence, prevalence target, family count, required sample budget, actual valid samples, failures, and acceptance.
- [ ] Add one suite row per required suite with planned families, required points, executed points, status, and report link.
- [ ] Mark unexecuted suites `planned_not_executed`; never present the current screening-only run as statistically accepted.
- [ ] Record statistical settings in the run manifest and resolved configuration.

### Task 4: Acceptance verification

- [ ] Run unit tests for exact sample-size math and family accounting.
- [ ] Run a small three-fluid executable smoke and verify `index.md` exposes all statistical suites and reports incomplete statistical acceptance.
- [ ] Run full Go tests, vet, build, and diff checks.
