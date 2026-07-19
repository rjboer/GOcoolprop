# Full CoolProp Validation Coverage Implementation Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Extend the independent validator from screening-only behavior into a complete, deterministic validation system covering the candidate capability manifest, all declared properties/input pairs, difficult thermodynamic regions, failures, adaptive refinement, and reconciled reports.

**Architecture:** Keep the validator separate from the Go production API. Build coverage from an explicit candidate capability manifest, evaluate candidate and persistent Python reference workers through normalized case records, and write deterministic per-case results before aggregate reports. Unsupported candidate functionality is explicit and never counted as a pass.

**Tech Stack:** Go 1.23.6 standard library, existing pure-Go CoolProp packages, persistent Python CoolProp 7.2.0 worker, JSONL, YAML configuration, Markdown/JSON/CSV.GZ reports.

## Global Constraints

- `D`/`Dmass` means kg/m³ and `Dmolar` means mol/m³.
- Candidate/reference units and fixed reference state must match; no hidden validator conversions may mask contract defects.
- Every declared capability and planned case receives an outcome.
- Deterministic case IDs, seeds, ordering, and statistics are invariant under bounded concurrency.
- A failed fluid does not abort other fluids.
- No fluid moves to `passed/` before all mandatory suites and adaptive work complete.

---

### Task 1: Capability manifest and unit-correct candidate adapter

**Files:** validation/internal/candidate/adapter.go, validation/internal/candidate/adapter_test.go

Add explicit fluids, aliases, input pairs, outputs, units, saturation, phase, and transport capability fields. Normalize canonical CoolProp names and convert only the candidate API’s mass/molar representation at the adapter boundary. Add tests for H/U/S/Cp/Cv mass and molar conversions, density aliases, supported pairs, and explicit unsupported outputs.

- [ ] Write failing conversion and manifest tests.
- [ ] Verify the tests fail before implementation.
- [ ] Implement the minimal manifest and conversion mapping.
- [ ] Run candidate tests.

### Task 2: Structured case records and comparator outcomes

**Files:** validation/internal/compare/metrics.go, validation/internal/compare/outcome.go, validation/internal/compare/*_test.go, validation/internal/generator/cases.go

Add stable case IDs, normalized error categories, phase comparison, zero-floor absolute tolerance, property-specific tolerances, and outcomes `passed`, `failed_numeric`, `failed_phase`, `consistent_error`, `error_mismatch`, `unsupported`, `panic`, `timeout`, and `validator_error`. Ensure every result preserves 17-significant-digit inputs.

- [ ] Add intentional offset, scale, near-zero, mass/molar, phase, panic, timeout, and consistent-error tests.
- [ ] Verify red tests.
- [ ] Implement comparator and stable record types.
- [ ] Verify green tests and deterministic ordering.

### Task 3: Reference batching, restart, timeout, and protocol containment

**Files:** validation/internal/reference/protocol.go, validation/internal/reference/protocol_test.go, validation/reference/coolprop_reference.py

Add batch calls, request deadlines, normalized startup metadata, malformed-response handling, worker restart/retry limits, and reference error categories. Keep one Python process per reference worker and make worker failures case-visible.

### Task 4: Complete deterministic generators

**Files:** validation/internal/generator/anchors.go, saturation.go, boundaries.go, flashes.go, invalid.go, adaptive.go and tests

Implement anchors, T–Dmass/T–Dmolar, P–T, Halton/Sobol-compatible deterministic sampling, T–Q/P–Q saturation, phase offsets, triple/critical/reducing/validity clouds, all manifest input pairs in both orders, round trips, invalid inputs, and local 5×5/21×21 refinement. Every generator must sort by stable case ID.

### Task 5: Bounded scheduler and progress modes

**Files:** validation/internal/scheduler/*, validation/internal/progress/*, validation/cmd/coolprop-validate/main.go

Implement bounded candidate/reference queues, batching, context deadlines, retries, checkpoints, resume, deterministic result ordering, and text/JSON/none progress. Expose `--config`, `--progress`, and resolved defaults from `validation/config/validation.yaml`.

### Task 6: Per-fluid matrices and statistical reports

**Files:** validation/internal/report/*, validation/internal/storage/*, validation/cmd/coolprop-validate/main.go

Write sorted `test-matrix.csv.gz`, failures, hotspots, coverage, progress, candidate/reference errors, adaptive reports, `summary.md`, `summary.json`, `manifest.json`, and `config.resolved.yaml`. Generate `index.md` after every fluid and at completion; reconcile all totals and report failure-rate separately from numerical error statistics.

### Task 7: Full acceptance and capability inventory

**Files:** validation/docs/validation-method.md, validation/docs/validation-plan.md, validation/config/validation.yaml, validation/acceptance/*

Run Water, Nitrogen, and Hydrogen through preflight, anchors, screening, saturation, boundaries, flashes, round trips, invalid inputs, adaptive reporting, and resume tests. Discover every JSON fluid and produce an explicit capability inventory showing implemented, unsupported, and failed CoolProp features. Never claim full CoolProp coverage while any declared mandatory capability remains unsupported.

## Execution order

Implement Tasks 1–2 first because all later suites depend on their contract and outcome model. Then implement Tasks 3–5, followed by reporting and acceptance. Each task must follow TDD and finish with focused tests before moving to the next task.

## Definition of done

The executable accepts `--config validation/config/validation.yaml`, runs all declared fluids/capabilities through every configured suite, writes complete reproducible run artifacts, reconciles raw/per-fluid/top-level totals, and explicitly reports any CoolProp functionality that the Go candidate does not implement.
