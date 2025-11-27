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
- [/] **Verification**
    - [/] Verify against sample values for Water <!-- id: 10 -->
    - [/] Verify against sample values for Nitrogen <!-- id: 11 -->
    - [/] Verify against sample values for Hydrogen <!-- id: 12 -->
- [ ] Implement saturation temperature from pressure
  - [ ] Use ancillary pS equation from JSON
  - [ ] Inverse solve if needed
- [ ] Implement saturation pressure from temperature
  - [ ] Direct evaluation of pS equation
- [ ] Calculate saturated liquid/vapor densities
  - [ ] Use rhoL and rhoV ancillary equations
- [ ] Two-phase quality calculations
  - [ ] Given P,Q or T,Q, calculate properties
- [ ] Add saturation property outputs to PropSI
  - [ ] "T_sat", "P_sat", "Q" (quality)
 
### Transport Properties
- [ ] **Viscosity**
  - [ ] Implement dilute gas contribution
  - [ ] Implement residual contribution
  - [ ] Parse transport data from JSON
- [ ] **Thermal Conductivity**
  - [ ] Similar structure to viscosity
  - [ ] Critical enhancement if present
- [ ] **Surface Tension**
  - [ ] Use ancillary equation from JSON
  - [ ] Only valid at saturation
- [ ] Add to PropSI outputs
  - [ ] "V" (viscosity), "L" (conductivity), "I" (surface tension)

### 5. Additional Helmholtz Terms
- [ ] Identify which terms are used by common fluids
- [ ] Implement missing term types:
  - [ ] IdealGasHelmholtzPower (if not already done)
  - [ ] ResidualHelmholtzExponential
  - [ ] ResidualHelmholtzNonAnalytic
  - [ ] Others as needed
- [ ] Test with fluids that use these terms

### 6. Improved Error Handling
- [ ] Better error messages for unsupported input pairs
- [ ] Validation of input ranges (T, P within valid bounds)
- [ ] Handle edge cases (critical point, triple point)
- [ ] Graceful degradation when data is missing

### 7. Performance Optimization
- [ ] Cache fluid data after first load
- [ ] Optimize Helmholtz energy calculations
- [ ] Profile hot paths
- [ ] Consider pre-computing common derivatives

### 8. Documentation & Examples
- [ ] Add godoc comments to all exported functions
- [ ] Create examples for common use cases
- [ ] Document supported fluids and properties
- [ ] Add usage guide to README

## Priority Order
1. **Additional Flash Algorithms** (Quick win, high value)
2. **Saturation Properties** (Essential for many applications)
3. **P-H Flash** (Most commonly needed after P-T)
4. **Transport Properties** (Viscosity and conductivity)
5. **Additional Helmholtz Terms** (As needed for specific fluids)
6. **P-S and T-H Flash** (Nice to have)
7. **Error Handling & Optimization** (Polish)
8. **Documentation** (Throughout)