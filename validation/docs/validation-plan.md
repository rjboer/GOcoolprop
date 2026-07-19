# Go CoolProp Validation Application Plan

## Objective

Build a separate validation application that compares the pure-Go CoolProp implementation with a pinned official CoolProp Python reference. The validator qualifies the declared candidate contract across fluids, properties, input pairs, phase regions, validity boundaries, error behavior, and adaptive diagnostic regions.

Validation is against official CoolProp software, not experimental measurements. The candidate and reference paths must remain independent; the candidate must never select or filter the reference domain.

## Definition of complete

A continuous domain cannot be enumerated point-by-point. A complete run therefore means:

- the declared validity envelope is covered by deterministic grids, boundary suites, saturation suites, and recorded quasi-random audit points;
- every declared input pair, output, alias, and capability is exercised;
- every planned case ends as `passed`, `failed_numeric`, `failed_phase`, `consistent_error`, `error_mismatch`, `unsupported`, `panic`, `timeout`, or `validator_error`;
- regions with deviations are adaptively investigated or explicitly classified as systematic/unresolved;
- no mandatory case, capability, validity cell, or random audit quota is silently skipped;
- raw results, configuration, versions, seeds, metadata, and statistics reconcile with the fluid reports and `index.md`.

## Contract and reference policy

The public contract must remain CoolProp-compatible:

- `D`, `Dmass`, and `DMASS` mean mass density in kg/m³;
- `Dmolar` and `DMOLAR` mean molar density in mol/m³;
- `T` is K, `P` is Pa, mass properties are per kg, molar properties are per mol, viscosity is Pa·s, conductivity is W/(m·K), surface tension is N/m, and molar mass is kg/mol;
- candidate and reference use the same fixed reference state for the complete run (`DEF` by default);
- aliases, canonical fluid names, input-pair direction, and unsupported capabilities are explicit contract data.

Preflight must verify versions, Git revision, Python version, backend, reference state, fluid availability, metadata (including triple/critical/reducing states), aliases, units, deterministic repeated calls, and NaN/infinity rejection. A contract mismatch fails the fluid before numerical testing.

## Architecture

Create a separate Go application under `validation/` with these boundaries:

- `validation/cmd/coolprop-validate/main.go`: configuration, lifecycle, orchestration, and CLI;
- `validation/internal/candidate/`: `PropertyEngine`, capability manifest, metadata, unit contract, and panic/timeout containment;
- `validation/internal/reference/`: persistent pinned Python workers using batched JSONL requests and restart/retry handling;
- `validation/internal/catalog/`: fluid discovery and canonical metadata;
- `validation/internal/generator/`: anchors, T–D, P–T, quasi-random, saturation, boundary, flash, round-trip, invalid-input, and adaptive cases;
- `validation/internal/compare/`: absolute, relative, normalized metrics, property tolerances, phase comparison, and error normalization;
- `validation/internal/scheduler/`: bounded queues, deterministic ordering, batching, deadlines, retries, checkpoints, and resume;
- `validation/internal/progress/`: text, JSON, and disabled progress modes;
- `validation/internal/report/`: fluid reports, aggregate statistics, hotspots, coverage, and top-level index reconciliation;
- `validation/internal/storage/`: manifests, resolved configuration, atomic result movement, checkpoints, and resume support;
- `validation/reference/coolprop_reference.py`: one persistent worker process per reference worker;
- `validation/config/validation.yaml`: checked-in defaults;
- `validation/docs/validation-method.md`: concise operational method;
- `validation/docs/validation-plan.md`: this complete implementation and qualification plan.

The candidate contract is:

```go
type PropertyEngine interface {
    Evaluate(ctx context.Context, req Request) Result
    Capabilities() CapabilityManifest
    Metadata(fluid string) FluidMetadata
}
```

Reference requests are batched JSONL records containing `request_id`, fluid, input pair, output, and values. Startup metadata must include CoolProp version/revision, Python version, supported fluids, backend, and reference-state policy. A configured version mismatch aborts the run.

## Validation stages

### Stage 0: preflight

Reject incompatible units, aliases, metadata, reference states, missing fluids, missing declared capabilities, nondeterminism, panics, worker failures, and non-finite behavior before expensive testing. Record all preflight outcomes in the fluid report.

### Stage 1: anchors and smoke

Use deterministic anchor states covering gas, liquid, saturation, supercritical, high-temperature/high-pressure, and near-critical regions. Compare every declared direct output, reversed input order, aliases, phase, and supported inverse/round-trip path. Structural failures may stop that fluid; numerical deviations continue to diagnostics.

### Stage 2: rough domain screening

Generate deterministic native T–Dmass/T–Dmolar grids, engineering P–T grids, and a recorded Sobol or Latin-hypercube audit set. Use metadata-derived validity limits, logarithmic density/pressure where appropriate, critical-density bands, dense-liquid/gas bands, and operational hydrogen pressure bands when valid. Filter only by reference validity; dual errors are recorded, not passed.

Compare every capability-manifest output, including mass/molar thermodynamics, phase, transport where defined, compressibility, and derived properties.

### Stage 3: saturation and phase boundaries

Test T–Q and P–Q with 256 temperatures and qualities `[0, .01, .1, .25, .5, .75, .9, .99, 1]`, saturated liquid/vapor, latent differences, and both sides of saturation. Use recorded offsets `±1e-2, ±1e-3, ±1e-4, ±1e-5` around the boundary. Treat canonical ambiguity as a phase/error-category comparison, not automatically as a numeric defect.

### Stage 4: critical and validity boundaries

Generate clouds around triple, critical, reducing, Tmin/Tmax, Pmin/Pmax, and critical density using relative offsets from `±1e-8` through `±1e-2`. Distinguish valid, inaccurate, expected rejection, inconsistent rejection, panic/crash, and non-finite outcomes. Apply critical-region multipliers only when reported separately.

### Stage 5: input pairs and flashes

Exercise every manifest-declared pair in both directions, including P–T, T–Dmass, T–Dmolar, density/pressure inverses, enthalpy/entropy/internal-energy flashes, and T–Q/P–Q. Start from stable reference states, reconstruct with each pair, and compare state variables and phase.

### Stage 6: round trips

Run P,T→H→P,H→T; P,T→S→P,S→T; T,D→P→P,T→D; T,Q→P→P,Q→T; and supported H,S round trips. Report candidate-vs-reference error and candidate self-consistency independently.

### Stage 7: invalid inputs and failures

Test unknown fluids/properties, duplicate or insufficient inputs, invalid pairs, NaN, infinity, negative/zero density, invalid pressure/quality, out-of-range states, phase ambiguity, unsupported transport, panics, worker crashes, corrupt responses, and timeouts. Normalize categories to `invalid_fluid`, `invalid_property`, `invalid_input_pair`, `out_of_range`, `non_finite_input`, `no_convergence`, `phase_ambiguity`, `not_implemented`, `internal_error`, `panic`, and `timeout`.

### Stage 8: adaptive deviation analysis

Refine tolerance failures, phase mismatches, one-sided errors, non-finite values, panics, timeouts, steep gradients, alternating pass/fail cells, empty-validity cells, and mixed-validity cells. Run local 5×5 then 21×21 grids, subdividing to the configured level/spacing/point budget. Classify evidence as constant offset, unit conversion, mass/molar confusion, reference-state offset, scale error, EOS, saturation, critical, phase selection, inverse convergence, discontinuity, transport, contract, runtime, or unknown.

## Metrics and acceptance

For each numeric comparison:

```text
absolute_error = |candidate - reference|
relative_error = absolute_error / |reference|  when |reference| > zero_floor[property]
normalized_error = absolute_error /
    (absolute_tolerance[property] + relative_tolerance[property] * |reference|)
```

Pass requires `normalized_error <= 1`. Near zero, report absolute and normalized error and mark relative percentage as not meaningful. Initial relative tolerances are metadata `1e-11`, direct EOS `1e-9`, P–T `1e-8`, inverse flashes `1e-7`, saturation `1e-7`, transport `1e-6`, and critical diagnostics `1e-5`; all are configurable and recorded before the run.

For each fluid, stage, pair, property, and phase, report attempted, valid, dual-error, mismatch, passed, failed, failure rate, mean, RMS, median, P95, P99, maximum relative/absolute/normalized error, worst point, and signed bias. Failure rate and numerical error percentages are separate metrics.

A fluid passes only when preflight, mandatory suites, phase/error behavior, adaptive analysis, capabilities, completeness, and report reconciliation pass with no unresolved panic, crash, timeout, non-finite result, or mandatory unsupported capability. The final claim is limited to agreement with official CoolProp for the recorded configuration and domain measure; it is not experimental validation.

For independent quasi-random audit points, report sample count, domain measure, confidence level, observed failure prevalence, one-sided upper confidence bound, and statistical acceptance threshold. Deterministic grids and low-discrepancy sequences are coverage tools, not by themselves statistical confidence evidence.

## Progress and lifecycle

Support `--progress=text`, `--progress=json`, and `--progress=none`. Each active fluid reports stage, planned/completed or adaptive level/queue/budget, passed, failed, error mismatches, active regions, current maximum error, elapsed time, and status.

Use the atomic lifecycle:

```text
results/<run-id>/
├── index.md
├── manifest.json
├── config.resolved.yaml
├── _running/
├── passed/
└── failed/
```

Write each fluid under `_running/<fluid>/`, finish all reports, then rename atomically to `passed/<fluid>/` or `failed/<fluid>/`. Interrupted work remains under `_running/` and is resumed from checkpoints or explicitly classified `incomplete_run`.

The validator writes `manifest.json` and `config.resolved.yaml` immediately after the reference handshake and before any fluid is scheduled. These files are immutable run inputs: they record the selected reference metadata, candidate/runtime metadata, fluid list, seed, worker counts, batch size, generator dimensions, and effective configuration.

## Reports

Each fluid package contains `summary.md`, `summary.json`, sorted compressed test matrix and failure records, `hotspots.csv`, `coverage.json`, `progress.jsonl`, candidate/reference error records, and adaptive region reports. Matrix rows contain stable case ID, fluid, stage, generator, pair, inputs, output, candidate/reference values and phases, absolute/relative/normalized errors, tolerance, error categories, outcome, and duration. Failure inputs retain at least 17 significant decimal digits.

Regenerate `index.md` after each completed fluid and at run completion. It must include run metadata, configuration hash, candidate/reference versions, worktree state, global totals, failure rate, numerical statistics, confidence claims, unsupported capabilities, incomplete fluids, worst deviations, suite totals, and links to every fluid report. Reconcile index totals against per-fluid summaries before marking the run complete.

## Reproducibility and containment

Record candidate commit and dirty state, Go/OS/architecture, Python/CoolProp version and revision, backend, reference state, resolved configuration and hash, seed, coordinate transforms, fluid limits, capability manifest, worker counts, batch size, timestamps, case IDs, and failure inputs. Concurrency must not alter cases, statistics, or classification.

Contain candidate panics, reference crashes/protocol corruption, timeouts, failed writes, and incomplete adaptive analysis per case/fluid. Restart a failed reference worker and retry its batch only within a configured limit. Abort the run only for unusable reference/configuration/result infrastructure.

## Configuration defaults

The checked-in YAML defaults are: CoolProp 7.2.0, HEOS, DEF, batch size 512, seed 20260719, 64×64 T–D and P–T screening, 2,048 quasi-random points, 256 saturation temperatures, qualities listed above, adaptive 5×5 then 21×21 grids, maximum four levels, and a two-million-point fluid budget. The resolved configuration is copied into every run directory.

## Implementation milestones and acceptance

1. **Reference and contract:** persistent worker, handshake, candidate adapter, manifest, unit/alias preflight, panic recovery. Water, Nitrogen, and Hydrogen pass preflight and one known point; an injected density-unit defect fails.
2. **Comparison core:** outcomes, metrics, tolerances, stable IDs, CSV/JSON output. Injected offset, scale, near-zero, mass/molar, phase, and timeout defects are detected/classified.
3. **Screening:** anchors, T–D, P–T, quasi-random, bounded concurrency, progress. Same seed produces the same cases and all configured audit quotas are reached.
4. **Difficult regions:** saturation, phase boundaries, critical/validity boundaries, flashes, and round trips. Phase and numerical failures remain distinct.
5. **Adaptive analysis:** hotspots, local grids, recursive subdivision, classification, and budgets. A localized defect refines only its neighborhood.
6. **Lifecycle/reporting:** per-fluid Markdown/JSON, index reconciliation, atomic movement, checkpoints, resume, and incomplete-run handling. Every completed fluid is in exactly one final folder and no mandatory case is omitted.

The definition of done is an independent command that produces `index.md`, `manifest.json`, `config.resolved.yaml`, `passed/`, and `failed/`, with all fluids, stages, coverage, outcomes, statistics, unsupported capabilities, incomplete work, and report links visible.
