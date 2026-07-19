# Reproducible Validation Run Metadata Implementation Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Make every validation run self-describing and reproducible by writing a manifest and resolved configuration at run start, including candidate/reference versions, run inputs, generator settings, and environment metadata.

**Architecture:** Extend the existing storage package with typed run metadata and atomic JSON/YAML artifact writers. Start the reference worker before creating the run so the startup handshake is captured in the manifest, then write immutable run artifacts before scheduling fluids. Preserve the existing screening and atomic per-fluid lifecycle.

**Tech Stack:** Go 1.23.6 module, standard library JSON/filesystem/process APIs, existing validation CLI and Python JSONL worker.

## Global Constraints

- Keep the validator separate from production CoolProp code.
- Preserve deterministic seeds and generator settings in every run.
- Do not change the candidate/reference numerical comparison behavior in this step.
- Keep existing user modifications and generated validation artifacts intact.

---

### Task 1: Define reproducibility artifacts

**Files:**
- Create: validation/internal/storage/metadata.go
- Test: validation/internal/storage/metadata_test.go

**Interfaces:**
- Produce storage.Manifest, storage.ConfigSnapshot, storage.WriteManifest, and storage.WriteResolvedConfig.
- Manifest fields include run ID, timestamps, Go version, Python version, CoolProp version, backend, reference state, Python executable, seed, generator counts, worker counts, batch size, fluid list, OS, architecture, and Git revision/dirty state when available.

- [ ] Step 1: Write failing tests for JSON manifest fields, deterministic config serialization, and resolved YAML output.
- [ ] Step 2: Run go test ./validation/internal/storage -run TestManifest -run TestResolvedConfig; expect failure because the types/functions do not exist.
- [ ] Step 3: Implement typed metadata and writers using only the standard library.
- [ ] Step 4: Run the focused storage tests and expect PASS.
- [ ] Step 5: Run git diff --check.

### Task 2: Create the manifest before fluid execution

**Files:**
- Modify: validation/cmd/coolprop-validate/main.go
- Modify: validation/cmd/coolprop-validate/main_test.go

**Interfaces:**
- run(ctx, opts) starts the reference worker, creates the run, writes manifest.json and config.resolved.yaml, then executes fluids.
- Reference startup metadata is captured from reference.Worker.Startup.

- [ ] Step 1: Add a failing integration-level test for run using a temporary result root and a fake executable/reference startup fixture; assert both artifacts exist and contain the configured seed/fluid list.
- [ ] Step 2: Run the focused CLI test and expect failure because run metadata is not written.
- [ ] Step 3: Refactor run so reference startup occurs before storage initialization, then write manifest/config with resolved generator values and runtime worker counts.
- [ ] Step 4: Ensure reference startup failure returns an error for a normal run rather than silently continuing without a canonical reference.
- [ ] Step 5: Run the focused CLI tests and expect PASS.

### Task 3: Verify the executable and documentation

**Files:**
- Modify: validation/docs/validation-plan.md
- Modify: validation/docs/validation-method.md

- [ ] Step 1: Document that manifest.json and config.resolved.yaml are written before fluid scheduling and are immutable run inputs.
- [ ] Step 2: Run go test ./....
- [ ] Step 3: Run go vet ./....
- [ ] Step 4: Build with go build -o validation/coolprop-validate.exe ./validation/cmd/coolprop-validate.
- [ ] Step 5: Run a small Water/Nitrogen/Hydrogen smoke validation with the Python 3.13 CoolProp executable and confirm both run-level artifacts exist and contain the recorded reference metadata.
