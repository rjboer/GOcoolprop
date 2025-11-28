# Port PropSI to Pure Go

- [x] **Planning**
    - [x] Create implementation plan <!-- id: 0 -->
    - [x] Define Go structs for Fluid JSON data <!-- id: 1 -->
- [x] **Data Layer**
    - [x] Implement JSON loader for fluid files <!-- id: 2 -->
    - [x] Load Water, Nitrogen, Hydrogen data <!-- id: 3 -->
- [x] **Thermodynamics Core**
    - [x] Implement Ideal Gas Helmholtz Energy (alpha0) <!-- id: 4 -->
    - [x] Implement Residual Helmholtz Energy (alphar) <!-- id: 5 -->
    - [x] Implement Property calculations (P, S, H, U, Cv, Cp) from Helmholtz energy <!-- id: 6 -->
- [x] **Solvers**
    - [x] Implement 1D solver (Brent/Newton) for T-D inputs <!-- id: 7 -->
    - [x] Implement P-T flash (solve for Density given P, T) <!-- id: 8 -->
- [x] **Interface**
    - [x] Implement `PropSI` function <!-- id: 9 -->

---

## Verification

- [/] Verify against sample values for Water <!-- id: 10 -->
  - [ ] Verify single-phase (gas) region
  - [ ] Verify compressed liquid region (e.g. 300 K, 1 atm and 10 MPa)
  - [ ] Verify saturation curve (Tsat, Psat, rhoL, rhoV)
- [/] Verify against sample values for Nitrogen <!-- id: 11 -->
  - [ ] Compare against reference tables / CoolProp for (T,P), (P,H), (P,S)
- [/] Verify against sample values for Hydrogen <!-- id: 12 -->
  - [ ] Compare against reference tables / CoolProp for (T,P), (P,H), (P,S)

---

## Fluids

**Goal:** Support ~50–100 common fluids via JSON, matching CoolProp’s coverage for typical engineering use.

- [x] Load initial fluids
  - [x] Water
  - [x] Nitrogen
  - [x] Hydrogen
- [ ] Add additional “core” pure fluids (first batch)
  - [ ] CO2
  - [ ] Methane
  - [ ] Oxygen
  - [ ] Argon
  - [ ] DryAir
- [ ] Extend to broader CoolProp set (target 50–100)
  - [ ] Copy additional JSON fluid definitions from CoolProp
  - [ ] Sanity-check each fluid loads and basic properties evaluate
  - [ ] Document supported fluid list

---

## Flash Algorithms

**Goal:** Robust 2-property flashes similar to CoolProp.

- [x] T–H flash (`FlashTH`)
  - [x] Implement core solver
  - [x] Basic unit tests (gas + liquid)
  - [ ] Robustness tests near saturation / critical
- [x] P–H flash (`FlashPH`)
  - [x] Implement solver
  - [x] Basic unit tests for gas and dense liquid
  - [ ] Improve initial guesses and convergence in two-phase regions
- [x] P–S flash (`FlashPS`)
  - [x] Implement solver
  - [x] Basic unit tests for gas and dense liquid
  - [ ] Improve stability near phase boundaries
- [ ] T–S flash (`FlashTS`)
  - [ ] Implement solver
  - [ ] Add unit tests (gas, liquid, near saturation)
- [ ] Integrate flashes into `PropSI`
  - [ ] Map all supported (input1, input2) pairs to flash routines
  - [ ] Return clear errors for unsupported pairs
- [ ] Add regression tests
  - [ ] Cross-check flash results vs CoolProp for representative states

---

## Saturation & Two-Phase Properties

**Current status:** basic saturation helpers and quality handling exist; need to formalize and verify.

- [x] Implement saturation temperature from pressure
  - [x] Use ancillary Ps(T) / Ts(P) equations from JSON
  - [x] Validate range and accuracy per fluid
- [x] Implement saturation pressure from temperature
  - [x] Direct evaluation of Ps(T) ancillary equation
  - [x] Validate against reference data
- [x] Calculate saturated liquid/vapor densities
  - [x] Use rhoL(T) and rhoV(T) ancillary equations
  - [x] Add tests across temperature range
- [x] Two-phase quality calculations
  - [x] Given T,Q or P,Q, calculate rho via v-mixing
  - [x] Compute Q from (T, rho) using v, vL, vV
  - [x] Verify behaviour near Q→0 and Q→1
- [x] Add saturation property outputs to PropSI
  - [x] `"T_SAT"`, `"P_SAT"`, `"Q"` (quality)
  - [x] Add tests for saturation outputs for each core fluid

---

## Transport Properties

- [ ] **Viscosity**
  - [ ] Parse viscosity correlation data from JSON
  - [ ] Implement dilute gas contribution
  - [ ] Implement residual / high-density contribution
  - [ ] Validate against CoolProp for key fluids
- [ ] **Thermal Conductivity**
  - [ ] Parse conductivity correlation data from JSON
  - [ ] Implement conductivity model
  - [ ] Implement critical enhancement if present
  - [ ] Validate against CoolProp
- [ ] **Surface Tension**
  - [ ] Use ancillary equation from JSON
  - [ ] Restrict use to saturation conditions
  - [ ] Validate against CoolProp / reference data
- [ ] Add to PropSI outputs
  - [ ] `"V"` (viscosity)
  - [ ] `"L"` (thermal conductivity)
  - [ ] `"I"` (surface tension)

---

## Additional Helmholtz Terms

**Current status:** Power & Gaussian residual terms are implemented; more term types needed for full CoolProp compatibility.

- [ ] Survey which term types are used by common fluids
- [ ] Implement missing term types:
  - [ ] ResidualHelmholtzExponential
  - [ ] ResidualHelmholtzNonAnalytic (Lemmon2005-style)
  - [ ] GERG-style mixture terms (for later mixture support)
  - [ ] Any remaining ideal-gas Helmholtz forms not yet covered
- [ ] Add unit tests for each new term type
- [ ] Enable fluids that depend on these terms and verify basic properties

---

## Improved Error Handling

- [ ] Better error messages for unsupported input pairs
  - [ ] Include input names/values and fluid in error text
- [ ] Validation of input ranges
  - [ ] Check T, P, rho against fluid’s valid range
  - [ ] Return explicit errors when outside EOS limits
- [ ] Handle edge cases
  - [ ] Critical point region
  - [ ] Triple point vicinity
  - [ ] Supercritical region (Q undefined)
- [ ] Graceful degradation when data is missing
  - [ ] Detect missing ancillary / transport correlations
  - [ ] Fallback to partial functionality with clear warnings

---

## Performance Optimization

- [ ] Cache fluid data after first load
- [ ] Optimize Helmholtz energy and derivative calculations
- [ ] Profile hot paths (flash routines, PropSI)
- [ ] Consider pre-computing common derivatives or using small lookup tables
- [ ] Investigate simple tabular backends for frequent states (optional)

---

## Documentation & Examples

- [ ] Add godoc comments to all exported functions
- [ ] Create examples for common use cases:
  - [ ] Simple (T,P) → (H,S,ρ) queries
  - [ ] P–H and P–S flashes
  - [ ] Saturation and two-phase examples (Q, Tsat, Psat)
- [ ] Document supported fluids and properties
- [ ] Add usage guide to README
- [ ] Add “full-port roadmap” section linking to major future features

---

## Major Future Features (Full CoolProp Port Roadmap)

These are large, multi-step projects.

### Humid Air

- [ ] Design humid-air backend (dry air + water mixture model)
- [ ] Implement psychrometric properties:
  - [ ] Relative humidity, dew point, wet-bulb temperature
  - [ ] Humidity ratio, moist air enthalpy
- [ ] Add dedicated API (e.g. `HumidAirSI`)

### Incompressible Fluids

- [ ] Define EOS approach for incompressible fluids (brines, glycols, oils)
- [ ] Implement property correlations / tables for selected fluids
- [ ] Add separate backend and integration with PropSI-like API

### Mixtures

- [ ] Design mixture data structures (composition, mixing rules)
- [ ] Implement GERG-2008 or similar for natural gas
- [ ] Implement departure functions and mixture Helmholtz terms
- [ ] Add mixture flash algorithms (T,P,z → H,S,ρ, etc.)

### Tabular Backends

- [ ] Implement table generation for selected fluids and property sets
- [ ] Implement bicubic (or similar) interpolation
- [ ] Integrate as an optional backend for performance-critical use cases

### Advanced CoolProp-Style Features

- [ ] Phase envelope tracing
- [ ] Critical point calculation / verification
- [ ] Thermodynamic consistency checks
- [ ] Reducing state calculations and derivatives
- [ ] Additional utility functions as needed

---

## Priority Order

1. **Flash Algorithms & Saturation Robustness**
   - Finish and harden T–H, P–H, P–S, add T–S
   - Verify saturation & two-phase behaviour for core fluids
2. **More Fluids**
   - Extend fluid set to common engineering fluids via JSON
3. **Transport Properties**
   - Viscosity, thermal conductivity, surface tension for key fluids
4. **Additional Helmholtz Terms**
   - Unlock more complex fluids that depend on extra term types
5. **Error Handling & Optimization**
   - Better messages, range checks, and performance improvements
6. **Documentation & Examples**
   - Keep docs in sync as features land
7. **Major Future Features**
   - Humid air, incompressibles, mixtures, tabular backends, and full CoolProp-style extras
