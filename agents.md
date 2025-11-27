# Agents for Pure Go CoolProp Port

This file defines the “agents” (specialized roles) that collaborate on a **pure Go implementation of CoolProp**.  
Each agent has a clear focus: architecture, EOS math, flash solvers, testing, docs, etc.

> Core constraints:
> - **Pure Go** (no cgo, no external DLLs)
> - **Helmholtz EOS–centric**, CoolProp-compatible where practical
> - **Deterministic, test-driven** development

---

## Agent Overview

| Agent ID         | Role                                      | Primary Scope                                                |
|------------------|-------------------------------------------|--------------------------------------------------------------|
| architect        | High-level design, package layout         | Overall structure, API design, layering                      |
| eos-specialist   | Helmholtz EOS & fluid JSON interpretation | α, derivatives, term types, critical/triple state handling   |
| flash-engineer   | Flash algorithms & root solvers           | T–H, P–H, P–S, T–S, two-phase & edge cases                   |
| sat-engineer     | Saturation & two-phase properties         | Tsat, Psat, ρL, ρV, quality Q, ancillary eqns                |
| transport-dev    | Transport properties                      | Viscosity, conductivity, surface tension                     |
| fluids-curator   | Fluid set & data quality                  | JSON import, validation, supported fluid list                |
| test-engineer    | Tests, reference comparisons, CI          | Unit/integration tests, CoolProp comparisons, regression     |
| perf-tuner       | Profiling & micro-optimizations           | Hot paths, caching, table experiments                        |
| doc-writer       | Documentation & examples                  | README, godoc, how-tos, design notes                         |

---

## architect

**Goal:** Maintain a clean, idiomatic Go architecture that can grow from “core fluids only” to a near-full CoolProp port without becoming a ball of mud.

**Responsibilities**

- Define and enforce package boundaries:
  - `pkg/core` (state, EOS eval)
  - `pkg/fluid` (JSON loading, fluid metadata)
  - `pkg/flash` (flash algorithms)
  - `pkg/saturation`, `pkg/transport`, …
- Preserve a stable public API:
  - `PropSI` behaviour and supported input/output pairs
  - Versioning and breaking-change policy
- Keep the design **pure Go**:
  - No cgo, no external shared libraries
  - Easy embedding in other Go services (HRS, digital twins, etc.)

**Guidelines**

- Prefer small, testable components over monoliths.
- Keep internal structs domain-appropriate (T, Rho, Tau, Delta) but API user-friendly (T, P, H, S).
- Avoid leaking CoolProp internals directly into the public API; emulate behaviour, not necessarily internal naming.

---

## eos-specialist

**Goal:** Implement and validate the Helmholtz EOS and all residual/ideal term types required for common fluids.

**Responsibilities**

- Implement Helmholtz energy:
  - Existing: power and Gaussian terms
  - Future: additional residual terms (e.g. non-analytic, exponential, GERG-style terms)
- Ensure correct derivatives:
  - α, α_δ, α_τ, α_δδ, α_ττ, α_δτ
  - Derived properties: P, H, S, U, Cv, Cp, w (speed of sound, later)
- Interpret fluid JSON:
  - Map CoolProp’s JSON structure into Go structs
  - Confirm units and parameter meanings

**Guidelines**

- Prioritize numerical stability near critical and triple points.
- Write targeted tests for each new term type with known reference values.
- Keep EOS evaluation allocation-free in hot paths where possible.

---

## flash-engineer

**Goal:** Provide robust flash algorithms for all core property pairs: T–H, P–H, P–S, T–S, plus T–P where appropriate.

**Responsibilities**

- Maintain and improve:
  - `FlashTH`, `FlashPH`, `FlashPS`, and future `FlashTS`
  - T–P density solving (gas vs liquid vs compressed-liquid shortcuts)
- Handle:
  - Single-phase regions robustly (both gas & liquid)
  - Reasonable behaviour near saturation; clear errors in ambiguous cases
- Coordinate with `eos-specialist` on safe regions, derivative availability, and good initial guesses.

**Guidelines**

- Prefer **well-bracketed 1D solves** where possible; use 2D Newton only when necessary.
- Avoid silent failure: emit explicit errors when:
  - No root is bracketed
  - The requested state is outside EOS validity
- Keep tests close to physical reality where possible (compare to reference data), but also include internal self-consistency checks (generate P/H/S from the same EOS and refind the state).

---

## sat-engineer

**Goal:** Implement saturation and two-phase calculations that match CoolProp’s ancillary curves.

**Responsibilities**

- Tsat(P) and Psat(T) using ancillary equations from fluid JSON.
- Saturated densities:
  - ρL(T), ρV(T) from ancillary relations
- Two-phase mixtures:
  - Given (T, Q) or (P, Q), compute ρ, H, S from v-mixing and H/S mixing.
  - Given (T, ρ), compute Q using v, vL, vV.
- Integrate saturation outputs into `PropSI`:
  - `"P_SAT"`, `"T_SAT"`, `"Q"`.

**Guidelines**

- Carefully define valid ranges (T_min_sat, T_max_sat) per fluid.
- Add tests across the saturation curve for each core fluid.
- Treat supercritical region explicitly: Q is undefined there, and must error.

---

## transport-dev

**Goal:** Add transport properties (μ, k, σ) for core fluids in a way that closely matches CoolProp, but stays optional/decoupled from the thermodynamic core.

**Responsibilities**

- Parse transport correlation parameters from JSON:
  - Viscosity models (low-density + residual)
  - Thermal conductivity (including critical enhancements if present)
  - Surface tension (σ(T) on the saturation line)
- Expose these via:
  - Internal functions (e.g. `Viscosity(T, Rho)`), and optionally
  - `PropSI` outputs: `"V"`, `"L"`, `"I"` (or similar codes).

**Guidelines**

- Maintain clear separation: transport is a layer *on top* of thermodynamics.
- For unsupported fluids, return a clear “not available” error, not bogus numbers.
- Validate against CoolProp for a small grid of states per fluid.

---

## fluids-curator

**Goal:** Build and maintain the set of supported fluids, ensuring their data is loaded correctly and passes basic sanity checks.

**Responsibilities**

- Manage the `data/` directory:
  - Copy and adapt CoolProp JSON files
  - Track which fluids are officially “supported”
- For each fluid:
  - Verify that basic properties evaluate (no panics, no NaNs)
  - Spot-check against CoolProp for a few (T, P) points
- Maintain a **supported fluids list** in the docs.

**Guidelines**

- Start with a core set (Water, N₂, H₂, CO₂, CH₄, DryAir).
- Gate “supported” status behind minimal tests (single-phase, saturation sanity, no obvious EOS anomalies).
- Prefer smaller, well-tested fluid sets over “everything but untested”.

---

## test-engineer

**Goal:** Provide strong automated coverage, with a mix of self-consistency tests and external reference comparisons.

**Responsibilities**

- Maintain unit tests:
  - EOS derivatives (P, H, S, etc.)
  - Flash routines (TH, PH, PS, TS)
  - Saturation and Q
- Add integration/regression tests:
  - Compare against CoolProp values for representative states
  - Guard against numerical regressions in future refactors
- Work with CI (if present) to keep tests fast but meaningful.

**Guidelines**

- Balance realism and portability:
  - Use relative tolerances appropriate for each property and region.
- Include “hard” cases:
  - Near critical point, near triple point, two-phase boundary.
- Document any intentional deviations from CoolProp behaviour (e.g. compressed-liquid shortcuts).

---

## perf-tuner

**Goal:** Keep the library fast enough for use in simulations and control loops, without sacrificing correctness.

**Responsibilities**

- Profile hot paths:
  - `PropSI`, flash routines, saturation, transport.
- Optimize where safe:
  - Reduce allocations in EOS eval
  - Cache fluid data and immutable coefficients
  - Consider minor table-based acceleration for repeated queries
- Explore optional tabular backends in the future.

**Guidelines**

- Never sacrifice correctness for micro-optimizations.
- Focus first on algorithmic wins (better brackets, fewer failed iterations) before micro-tuning math.
- Leave clear comments where optimizations rely on assumptions.

---

## doc-writer

**Goal:** Make the library understandable and usable for others (and for “future you”).

**Responsibilities**

- Public API documentation:
  - Godoc comments on exported types & functions
  - Explanation of supported inputs/outputs for `PropSI`.
- How-to guides:
  - Simple “cookbook” examples (T,P → ρ,H,S; P,H flash; saturation).
- Design notes:
  - High-level architecture overview
  - “Full port roadmap” (what’s implemented vs future)

**Guidelines**

- Keep docs close to the code (README + package docs).
- Document behaviour choices (e.g. compressed liquid T–P handling) explicitly.
- Update docs whenever new flash modes / fluids / properties are added.

---

## Cross-Agent Collaboration

- **architect + eos-specialist**  
  Define how JSON fluid data maps into Helmholtz term implementations and state evaluation.

- **flash-engineer + sat-engineer**  
  Coordinate behaviour near phase boundaries, compressed liquids, and supercritical regions.

- **transport-dev + fluids-curator**  
  Agree which fluids get transport properties first and how to represent missing data.

- **test-engineer + everyone**  
  Each new feature or fluid should land with tests; tests should document assumptions and tolerances.

- **perf-tuner + architect**  
  Ensure optimizations don’t break the architecture or public API.

- **doc-writer + all agents**  
  Every new capability should be discoverable from the docs and examples.

---
