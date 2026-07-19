# High-Pressure Subcritical Phase Selection Fix

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Correct T–P density selection for subcritical states whose pressure is at or above critical pressure, and prevent recurrence of the Water/Hydrogen validation defect.

**Root cause:** `pkg/flash/inferPhaseTP` returned `phaseAny` when `P >= Pcrit`, even when `T < Tcrit`. The density solver then selected the root nearest ideal-gas density. CoolProp correctly selects the compressed-liquid state because `P > Psat(T)` at these subcritical temperatures.

**Evidence:** Water at T=273.16172684000003 K and P=22064000.000000026 Pa returned 321.9999998635203 kg/m³ instead of 1010.735995570951 kg/m³. Hydrogen at T=13.957986043 K and P=1296400.000000001 Pa returned 30.4952487928738 kg/m³ instead of 78.04483347628127 kg/m³. Both failures reproduced in the validator and both phase-inference regressions failed before the fix.

## Executed tasks

- [x] Add a focused Water/Hydrogen regression test in `pkg/flash/single_phase_regression_test.go`.
- [x] Run the regression test before the fix and observe both cases classified as `phaseAny`.
- [x] Change `inferPhaseTP` to bypass saturation classification only when `T >= Tcrit`.
- [x] Run the focused regression test and confirm both states select `phaseLiquid` with density above critical density.
- [x] Run `go test ./pkg/flash`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Build `validation/coolprop-validate.exe`.
- [x] Run the three-fluid end-to-end screening smoke against CoolProp 7.2.0/HEOS/DEF.

## Acceptance

The fix is accepted when the focused regression, full Go test suite, vet, build, and the reproducible three-fluid screening run pass. The smoke run must report Water, Nitrogen, and Hydrogen with zero failed cases.
