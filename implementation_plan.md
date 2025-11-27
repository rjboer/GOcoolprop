# Implementation Plan - PropSI Port to Go

## Goal
Port the `PropSI` function from CoolProp to a pure Go implementation, supporting Water, Nitrogen, and Hydrogen initially.

## Proposed Changes

### Data Structures (`pkg/fluid`)
- Define structs to match the CoolProp JSON schema.
- `Fluid` struct containing `EOS`, `CriticalRegion`, etc.
- `Alpha0` and `AlphaR` term structs.

### Core Logic (`pkg/core`)
- `Helmholtz` interface/structs to calculate energy terms.
- Implementation of standard terms (Ideal Gas, Residual).
- `State` struct to hold current T, Rho, Fluid, and calculated properties.

### Solvers (`pkg/solver`)
- `Brent` method for 1D root finding (needed for P-T flash).
- `Newton` method if needed.

### Interface (`pkg/props`)
- `PropSI(output, name1, val1, name2, val2, fluidName)`
- Dispatcher to load fluid and call appropriate solver.

## Verification Plan
### Automated Tests
- Create a test file `props_test.go`.
- Compare results for Water (T=300K, P=101325Pa) against known values.
- Compare results for Nitrogen and Hydrogen.
