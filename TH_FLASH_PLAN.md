# T-H Flash Implementation Plan

## Goal
Implement T-H flash: Given Temperature (T) and Enthalpy (H), solve for Density (ρ).

## Problem Analysis

**Input**: T (temperature), H (target enthalpy)
**Output**: ρ (density)

**Equation to solve**:
```
H(T, ρ) - H_target = 0
```

This is a **1D problem** because T is known, and we only need to find ρ.

## Approach

### 1. Use Brent's Method (Already Implemented)
- We already have a working Brent solver in `pkg/solver/brent.go`
- Define objective function: `f(ρ) = H(T, ρ) - H_target`
- Find ρ such that f(ρ) = 0

### 2. Initial Bounds Strategy
For a given T and H_target, we need to bracket the solution:

**Gas phase** (low density):
- Lower bound: ρ_min ≈ 1e-8 mol/m³
- Upper bound: ρ_max ≈ ρ_critical / 2

**Liquid phase** (high density):
- Lower bound: ρ_min ≈ ρ_critical * 0.8
- Upper bound: ρ_max ≈ ρ_triple_liquid * 1.2

**Strategy**: Try gas phase first, then liquid phase if no solution found.

### 3. Implementation Location

Create new package: `pkg/flash/`

**Files**:
- `pkg/flash/th_flash.go` - T-H flash implementation
- `pkg/flash/th_flash_test.go` - Tests

## Implementation Steps

1. Create `pkg/flash` package
2. Implement `FlashTH(fluid, T, H)` function
3. Add tests with known T-D pairs
4. Update `PropSI` to support T-H input pairs
5. Verify with test cases

## Test Strategy

For each test:
1. Start with known (T, ρ) pair
2. Calculate H = H(T, ρ) using State
3. Call FlashTH(T, H)
4. Verify returned ρ matches original ρ

Test fluids: Water, Nitrogen, Hydrogen
Test phases: Gas, Liquid

## Integration with PropSI

Add support for:
- "T" + "H" → FlashTH
- "H" + "T" → FlashTH

## Success Criteria

- [ ] FlashTH function implemented
- [ ] Tests pass for gas phase
- [ ] Tests pass for liquid phase  
- [ ] PropSI supports T-H input pairs
- [ ] Accuracy within 0.1% for density
