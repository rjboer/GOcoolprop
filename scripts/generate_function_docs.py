from pathlib import Path
import re


root = Path(__file__).resolve().parents[1]
out_dir = root / "snippets" / "functions"
out_dir.mkdir(parents=True, exist_ok=True)


def infer_test_conditions(item):
    ident = item["id"]
    if ident.startswith("core.State.") and ident not in {"core.State.Update"}:
        return [
            "Evaluate the function at three state classes: a dilute-gas point where ideal-gas behavior is dominant, a dense-liquid point where residual derivatives dominate, and a near-critical but still stable single-phase point where the derivatives are numerically sensitive.",
            "For Water, Nitrogen, and Hydrogen, use one fixed benchmark state per class and record the exact input pair, returned value, and relative error against the chosen reference.",
            "When the function represents a thermodynamic derivative, compare the analytic result against a central finite-difference estimate computed from the corresponding primary property function using a documented perturbation size.",
            "When the function represents a primary property, also verify at least one thermodynamic identity that depends on it, for example h = u + p/rho, ds/dT|rho = cv/T, or cp >= cv in stable single phase.",
            "Repeat the same test twice on the same state object without changing inputs and confirm the second call returns the same value, to guard against stale-cache or order-of-call issues.",
        ]
    if ident == "core.State.Update":
        return [
            "Run a repeated-update sequence on the same state object using a dilute-gas point, a dense-liquid point, and a near-critical stable single-phase point, then verify that Pressure, H, S, U, Cv, Cp, and all cached reduced variables correspond to the latest update only.",
            "For Water, Nitrogen, and Hydrogen, compute tau and delta manually from the EOS reducing state and compare them to the cached values after each update step.",
            "After each update, compare the cached alpha0, alphar, and total derivative fields against a fresh recomputation from the loaded term lists to verify cache coherence.",
            "Run one test where the object is updated from a liquid-like point directly to a gas-like point and back again, to verify no branch-specific state leaks between calls.",
            "Run invalid-input cases such as T <= 0 and rho <= 0 and verify the resulting state becomes explicitly unusable or fails in the documented way rather than returning plausible-looking values.",
        ]
    if ident.startswith("core.IdealGasHelmholtz") or ident.startswith("core.ResidualHelmholtz"):
        return [
            "Evaluate the base term value at a representative interior point where tau and delta are both comfortably away from singular behavior, and record the exact expected numeric value.",
            "Check first derivatives, second derivatives, and the mixed derivative against central finite differences of the term value using documented step sizes for tau and delta.",
            "Run at least one second test point chosen to stress the family numerically, for example small tau for Planck-Einstein terms, delta near one for reduced-density residual terms, or a Gaussian-center point for Gaussian residual terms.",
            "For derivative methods that should be identically zero for the normalized form, assert exact zero if the implementation is analytic rather than only approximate agreement.",
            "Verify mixed-derivative symmetry explicitly by comparing d/delta(d/dtau(term)) and d/dtau(d/ddelta(term)) at the same point.",
        ]
    if ident == "core.HelmholtzEnergy.Update":
        return [
            "Build a synthetic HelmholtzEnergy instance with multiple alpha0 and alphar terms and verify the returned totals equal the manual sum of each individual term contribution.",
            "At a fixed tau and delta, compare the total first derivatives, second derivatives, and mixed derivative against finite-difference estimates of the aggregated total alpha value.",
            "Run one case with only ideal terms, one with only residual terms, and one with both, to confirm no term list is accidentally skipped.",
            "Verify mixed-derivative symmetry numerically and verify repeated calls at the same input are deterministic.",
        ]
    if ident == "core.NewState":
        return [
            "Construct state objects for fully supported fluids and verify term counts, reducing-state values, gas constant, and cached model structure all match what is expected from the source JSON.",
            "Verify parser-normalized ideal-gas families such as Hydrogen FunctionT and Nitrogen ideal-gas power terms are converted into the intended internal term structures and can be evaluated immediately.",
            "Construct state objects for deliberately unsupported fluids and assert that the returned error names the exact missing term family rather than failing later during property evaluation.",
            "Run one constructor test on a fluid with transport-rich JSON such as R32 to verify schema loading success is cleanly separated from thermodynamic term support failure.",
        ]
    if ident.startswith("fluid.LoadFluid") or ident == "fluid.LoadFluid":
        return [
            "Load representative JSON files from different schema families, including a simple fluid, a fluid with non-analytic residual terms, and a transport-rich fluid with schema variants, and verify all required top-level sections are present after decoding.",
            "Verify exact scalar fields such as fluid name, molar mass, gas constant, critical state, and reducing state against known values from the JSON source for at least one reference fluid.",
            "Attempt to load malformed JSON, truncated JSON, and JSON missing required sections and verify the returned error clearly indicates whether the failure came from file I/O or schema decoding.",
            "Include a transport-rich fluid such as R32 and verify that the decoded structure preserves the original shape well enough for later transport-model dispatch tests.",
        ]
    if ident.startswith("fluid.GetFluidFilename") or ident.startswith("fluid.ListAvailableFluids"):
        return [
            "Verify canonical names, case variants, and common aliases all resolve to the expected JSON filename for several fluid families including gases, water, refrigerants, and hydrocarbons.",
            "Verify unknown names, malformed names, and names that only differ by spacing or dash placement either normalize correctly or fail cleanly in the documented way.",
            "Cross-check the resolved filenames or listed fluids against the actual contents of the data directory so registry drift is detected early.",
        ]
    if ident == "fluid.AncillaryCurve.Evaluate":
        return [
            "Evaluate at least one pressure ancillary, one saturated-liquid-density ancillary, and one saturated-vapor-density ancillary using temperatures near the low end, midpoint, and upper end of the valid range.",
            "For each ancillary type, compare the returned value against either a known reference point or a direct expected value computed from the correlation formula and source coefficients.",
            "Verify that the same ancillary evaluated twice at the same temperature returns exactly the same result and that unsupported ancillary types fail explicitly rather than returning zero silently.",
        ]
    if ident.startswith("saturation."):
        return [
            "Check the function at temperatures or pressures well inside the ancillary valid range and record exact benchmark values for Water and Nitrogen first.",
            "Run a round-trip test where Psat(T) is fed into Tsat(P) and verify the original temperature is recovered within the documented tolerance.",
            "For density helpers, verify the returned saturated liquid and vapor densities are physically ordered and lie in a plausible range for the selected fluid.",
            "Add at least one out-of-range test per function and verify the function returns a clear range error rather than extrapolating silently.",
        ]
    if ident.startswith("solver."):
        return [
            "Run analytic root problems with known solutions, including one easy polynomial case and one non-linear case, and verify both convergence and final error.",
            "Run monotonic thermodynamic-style objectives similar to saturation inversion or density solving to verify the solver behaves well on functions shaped like the real library use cases.",
            "Verify expected failures such as singular Jacobians, unbracketed roots, invalid starting points, or iteration exhaustion produce explicit and distinguishable errors.",
            "Record the number of iterations or at least the fact that the solver converges without oscillation for the nominal cases.",
        ]
    if ident.startswith("flash."):
        return [
            "Use benchmark states derived from known T,rho reference points to generate target P,H,S values for both a gas-like and a liquid-like case, then solve the flash and compare the recovered T and rho to the original state.",
            "After the flash returns, recompute the defining residual equations using the returned state and verify both residuals are within tolerance, not just that the solved variables look plausible.",
            "Add a near-saturation case where the phase branch matters and verify the solver either selects the documented single-phase or endpoint branch correctly or returns a clear unsupported-state error.",
            "Add at least one unsupported two-phase interior test and verify the solver fails explicitly rather than converging to a misleading pseudo-solution.",
        ]
    if ident == "props.PropSI":
        return [
            "Test each supported input pair for Water, Nitrogen, and Hydrogen using benchmark states that cover dilute gas, dense liquid, and saturation-endpoint usage where applicable.",
            "For each input pair, verify both value accuracy and branch correctness, for example T,P returning the correct density branch and T,Q or P,Q returning the correct saturation endpoint only for Q=0 and Q=1.",
            "For output keys such as H, S, U, Cv, Cp, P, D, P_SAT, and T_SAT, verify that the same physical state gives self-consistent results across multiple valid input pairs, and verify Q only on saturation endpoints.",
            "Verify unsupported input pairs, unsupported fluids, interior-quality requests, supercritical quality requests, and out-of-range saturation requests all fail with explicit errors that include enough context to diagnose the call.",
        ]
    if ident.startswith("transport."):
        return [
            "Check the function at a dilute-gas benchmark and a dense-state benchmark where the selected model family supports both regions, and record expected values from the chosen reference.",
            "Use the primary validation fluid for the model family first, for example Nitrogen for the currently supported conductivity and viscosity paths, then add one second fluid if the same model family is shared.",
            "Verify that unsupported transport model families, hardcoded-only paths, or ECS paths not yet implemented fail explicitly and do not return misleading placeholder values.",
            "For composite transport functions, verify the total equals the sum of the supported sub-contributions when the model is defined that way.",
        ]
    return [
        "Verify the normal success path with representative in-range inputs.",
        "Verify at least one failure path with invalid or unsupported inputs.",
        "Cross-check the result against a trusted reference or internal identity where applicable.",
    ]


def infer_boundary_conditions(item):
    ident = item["id"]
    if ident.startswith("core.State.") and ident not in {"core.State.Update"}:
        return [
            "Very low density approaching the ideal-gas limit.",
            "High density in a stable liquid region before spinodal behavior.",
            "Near-critical stable single-phase states where derivatives become sensitive.",
        ]
    if ident == "core.State.Update":
        return [
            "Temperature approaching zero or negative input.",
            "Density approaching zero or negative input.",
            "Reduced variables near tau=1 or delta=1 where critical behavior is sensitive.",
        ]
    if ident.startswith("core.IdealGasHelmholtzLead") or ident.startswith("core.IdealGasHelmholtzLogTau"):
        return [
            "Small positive tau and delta values that remain physically valid.",
            "Moderately large tau values to test derivative scaling.",
            "Points close to singular log arguments, while still remaining valid.",
        ]
    if ident.startswith("core.IdealGasHelmholtzPlanckEinstein"):
        return [
            "Small tau where exponential cancellation can occur.",
            "Large tau where exponential terms become very small.",
            "Parameter combinations that keep logarithm arguments strictly positive.",
        ]
    if ident.startswith("core.ResidualHelmholtzPower"):
        return [
            "delta near zero for low-density behavior.",
            "delta near one where reduced-density scaling changes regime.",
            "Large delta values where exponential damping or growth terms dominate.",
        ]
    if ident.startswith("core.ResidualHelmholtzGaussian"):
        return [
            "delta and tau near the Gaussian center.",
            "Points far from the Gaussian center where the exponential term is small.",
            "Second-derivative checks where cancellation can occur.",
        ]
    if ident.startswith("fluid.LoadFluid"):
        return [
            "Missing file path.",
            "Unreadable or malformed JSON.",
            "Schema variants with nested objects versus arrays in transport content.",
        ]
    if ident == "fluid.AncillaryCurve.Evaluate" or ident.startswith("saturation."):
        return [
            "Tmin and Tmax of the ancillary correlation.",
            "Just inside and just outside the valid ancillary range.",
            "Temperatures approaching the critical limit where ancillary accuracy degrades.",
        ]
    if ident.startswith("solver.Brent"):
        return [
            "Exact root at an interval endpoint.",
            "Very tight bracket with small residuals.",
            "Intervals with no sign change.",
        ]
    if ident.startswith("solver.Newton2D"):
        return [
            "Initial guesses close to the solution.",
            "Initial guesses far from the solution but still physically meaningful.",
            "Jacobian determinant approaching zero.",
        ]
    if ident.startswith("flash."):
        return [
            "Gas-side initial conditions.",
            "Liquid-side initial conditions.",
            "Near-saturation and near-critical states where branch selection is difficult.",
        ]
    if ident == "props.PropSI":
        return [
            "Input pairs near saturation boundaries.",
            "Supercritical states where Q is undefined.",
            "Unsupported input/output combinations and missing-fluid cases.",
        ]
    if ident.startswith("transport."):
        return [
            "Low-density dilute limit.",
            "High-density state where residual transport terms dominate.",
            "Out-of-scope model families such as ECS or hardcoded paths that are not yet implemented.",
        ]
    return [
        "Lower valid input bound.",
        "Upper valid input bound.",
        "One clearly invalid input case.",
    ]


def infer_validation_gaps(item):
    ident = item["id"]
    gaps = []

    if ident == "core.NewState":
        gaps.extend(
            [
                "The markdown signature must match the current code: `func NewState(f *fluid.FluidData) (*State, error)` rather than the old non-error-returning form.",
                "The current repo validates constructor success and explicit unsupported-term failure, but it still does not compare constructor output against a live CoolProp backend object.",
            ]
        )
    if ident == "fluid.AncillaryCurve.Evaluate":
        gaps.extend(
            [
                "The markdown signature must match the current code: `func (ac *AncillaryCurve) Evaluate(T float64) (float64, error)` rather than the old value-only form.",
                "Current validation checks local correlation behavior, but it does not prove parity against a live CoolProp ancillary evaluation call for every supported fluid.",
            ]
        )
    if ident == "transport.Conductivity":
        gaps.extend(
            [
                "The old note that the file is structurally broken is no longer true; the code now builds and has basic tests.",
                "Current validation still falls short of full CoolProp transport parity because ECS and critical-enhancement paths are not implemented.",
            ]
        )
    if ident.startswith("transport."):
        gaps.append(
            "Current tests cover only the implemented local transport paths; they do not yet match the full transport model coverage of regular CoolProp."
        )
    if ident.startswith("core.State.") or ident in {"core.HelmholtzEnergy.Update", "core.NewState"}:
        gaps.append(
            "Current validation uses internal identities, JSON reference states, and finite differences; it does not yet call regular CoolProp HEOS directly at runtime."
        )
    if ident.startswith("flash.") or ident == "props.PropSI":
        gaps.append(
            "Current validation covers only a small benchmark subset and does not yet prove branch parity with regular CoolProp across the full supported input space."
        )
    if ident.startswith("saturation.") or ident == "fluid.AncillaryCurve.Evaluate":
        gaps.append(
            "Current saturation validation is ancillary-based and benchmark-based; it is not yet a direct product-to-product comparison against live CoolProp calls for all fluids."
        )
    if ident.startswith("solver."):
        gaps.append(
            "Current solver validation proves local numerical behavior, not CoolProp parity directly, because CoolProp does not expose the same solver API as the public reference surface."
        )
    if ident.startswith("fluid.LoadFluid") or ident.startswith("fluid.GetFluidFilename") or ident.startswith("fluid.ListAvailableFluids"):
        gaps.append(
            "These functions can be aligned with CoolProp data and alias behavior, but they are GOcoolprop-specific APIs rather than one-to-one public CoolProp functions."
        )

    if ident == "props.PropSI":
        gaps.extend(
            [
                "The current repo tests do not yet enforce the tighter parity tolerances described in the project plan; they are still looser benchmark-style checks.",
                "Current tests do not yet compare against live `CoolProp HEOS` results for every supported output and input-pair combination.",
            ]
        )
    if ident.startswith("core.State.") and ident != "core.State.Update":
        gaps.append(
            "The current derivative and reference-point tolerances in `pkg/validation` are looser than the desired final parity tolerances against regular CoolProp."
        )

    if not gaps:
        gaps.append(
            "No function-specific CoolProp mismatch is currently documented beyond the general scope gap that the repo still relies mostly on local benchmark validation rather than live CoolProp runtime comparison."
        )
    return gaps


def infer_gap_analysis(item):
    ident = item["id"]
    gaps = []

    gaps.append(
        "The functional description in this file states the intended contract, but regular CoolProp should be treated as the authoritative behavioral reference whenever the two differ."
    )

    if ident.startswith("core.State.") or ident in {"core.HelmholtzEnergy.Update", "core.NewState"}:
        gaps.extend(
            [
                "CoolProp evaluates Helmholtz-based state behavior through a mature backend that keeps ideal and residual contributions, reducing-state logic, and derivative usage aligned across all downstream properties and flashes.",
                "GOcoolprop is closer than before, but its parity evidence still comes mainly from internal consistency checks, reference JSON states, and local finite-difference validation rather than direct product-to-product backend comparison.",
                "Where this description asks for canonical formulas or cache consistency, the remaining gap is not only implementation detail but also proof that the implementation matches CoolProp across the supported state space.",
            ]
        )
    if ident.startswith("core.IdealGasHelmholtz") or ident.startswith("core.ResidualHelmholtz"):
        gaps.extend(
            [
                "CoolProp term-family behavior is defined not only by formulas but also by parser normalization, coefficient conventions, and interactions with backend-wide derivative usage.",
                "GOcoolprop term implementations may be mathematically correct in isolation while still differing from CoolProp if parser-level transformations, coefficient interpretation, or edge-case handling are incomplete.",
            ]
        )
    if ident.startswith("fluid.LoadFluid") or ident.startswith("fluid.GetFluidFilename") or ident.startswith("fluid.ListAvailableFluids"):
        gaps.extend(
            [
                "These functions are local GOcoolprop APIs, so parity with CoolProp means matching data interpretation, alias behavior, and failure semantics rather than matching a public CoolProp function one-to-one.",
                "The main gap here is registry and schema behavior: CoolProp has broader fluid coverage and mature alias handling, while GOcoolprop still documents the target behavior more completely than it proves it.",
            ]
        )
    if ident == "fluid.AncillaryCurve.Evaluate" or ident.startswith("saturation."):
        gaps.extend(
            [
                "CoolProp saturation behavior is backend-aware and can combine ancillary and EOS logic depending on the path and fluid.",
                "GOcoolprop currently relies on ancillary-centric saturation handling for the supported scope, so this description is necessarily narrower than full CoolProp behavior even when the formulas are locally correct.",
            ]
        )
    if ident.startswith("solver."):
        gaps.extend(
            [
                "CoolProp does not expose these exact solver functions as public contracts, so parity here means numerical robustness sufficient to support CoolProp-like thermodynamic behavior rather than strict API equivalence.",
                "The remaining gap is therefore indirect: if these solvers fail in hard regions or have weaker branch logic, higher-level thermodynamic parity will drift even if the solver unit tests pass.",
            ]
        )
    if ident == "flash.DensityTP":
        gaps.extend(
            [
                "CoolProp treats T,P flashes as backend-level phase-aware operations with mature region logic and broader edge-case handling.",
                "GOcoolprop now documents and implements a stricter single-phase contract with explicit saturation-boundary rejection, which is closer operationally but still narrower than full CoolProp flash coverage.",
                "The remaining gap is strongest near saturation, near the critical region, and in proof of branch correctness across more than a small benchmark set.",
            ]
        )
    if ident.startswith("flash."):
        gaps.extend(
            [
                "CoolProp flash functions effectively cover richer phase behavior, stronger initialization strategies, and more mature branch handling than the current GOcoolprop implementation.",
                "GOcoolprop now describes the supported scope more honestly, but the functional description is still broader than the currently proven parity because unsupported interior two-phase states are rejected instead of being solved in a CoolProp-equivalent way.",
                "The main difference is no longer only solver choice; it is scope. CoolProp supports more physical cases, while GOcoolprop documents a narrower, explicit contract.",
            ]
        )
    if ident == "props.PropSI":
        gaps.extend(
            [
                "CoolProp `PropSI` is the product-level contract, so any mismatch in phase semantics, supported pairs, output meaning, or error behavior is user-visible immediately.",
                "GOcoolprop is closer because the supported pairs now route through stricter flash logic and endpoint-only quality handling, but it still exposes a smaller supported region than regular CoolProp.",
                "The largest remaining parity gaps are live-output comparison, broader phase coverage, tolerance tightening, and exact alignment of edge-case semantics around saturation and supercritical states.",
            ]
        )
    if ident.startswith("transport."):
        gaps.extend(
            [
                "CoolProp transport support includes broader model-family coverage, more fluid-specific paths, and more mature critical-region handling than the current GOcoolprop implementation.",
                "The functional description here is therefore partly aspirational: it names the contract GOcoolprop should meet, while current parity is limited to the implemented model families and validation points.",
            ]
        )

    gaps.extend(infer_validation_gaps(item))
    return gaps


def infer_current_scope(item):
    ident = item["id"]
    scope = []

    if ident == "core.State.Update":
        return [
            "The current implementation updates reduced variables from the reducing state and keeps separated ideal and residual Helmholtz caches.",
            "Invalid inputs currently poison the state with `NaN`-like unusable values rather than returning an explicit error from `Update` itself.",
            "Parity is supported mainly for the validated core fluids and still depends on local tests rather than live CoolProp backend comparison.",
        ]
    if ident.startswith("core.State."):
        return [
            "The current implementation exposes analytic Helmholtz-based property and derivative methods for supported fluids.",
            "These methods are validated with internal identities, benchmark states, and finite differences, not yet with full live CoolProp runtime comparison.",
            "Behavior outside the supported EOS term set remains an explicit construction-time failure rather than degraded best-effort evaluation.",
        ]
    if ident == "core.NewState":
        return [
            "The current constructor returns `(*State, error)` and fails explicitly on unsupported EOS content.",
            "Supported term families for the current core fluids are constructed into separated ideal and residual term lists.",
            "Constructor parity with CoolProp parser behavior is partial: supported families are normalized locally, but full parser-equivalence is not yet proven for all fluids.",
        ]
    if ident.startswith("core.IdealGasHelmholtz") or ident.startswith("core.ResidualHelmholtz"):
        return [
            "The current implementation provides direct analytic term evaluation and derivative methods for the implemented term families.",
            "Term math is locally validated by finite-difference checks, but parser-level equivalence with all CoolProp term conventions is not yet proven globally.",
        ]
    if ident == "core.HelmholtzEnergy.Update":
        return [
            "The current implementation aggregates ideal and residual term contributions separately and as totals.",
            "Correctness is supported by local derivative checks and by downstream state-math tests for the core fluids.",
        ]
    if ident.startswith("fluid.LoadFluid"):
        return [
            "The current loader can decode the supported CoolProp-style JSON structures used by the repo, including several transport-schema variants.",
            "Schema handling is broad enough for the current core fluids and selected transport-rich files, but not yet demonstrated as complete for the full CoolProp data set.",
        ]
    if ident.startswith("fluid.GetFluidFilename") or ident.startswith("fluid.ListAvailableFluids"):
        return [
            "The current implementation provides a local registry convenience layer over the repo data directory.",
            "This is not a one-to-one CoolProp public API, so parity here means matching naming and lookup expectations rather than API identity.",
        ]
    if ident == "fluid.AncillaryCurve.Evaluate":
        return [
            "The current implementation evaluates supported ancillary families and returns explicit errors for unsupported types.",
            "Ancillary behavior is used directly by the current saturation helpers and therefore defines the supported saturation scope of GOcoolprop.",
        ]
    if ident.startswith("saturation."):
        return [
            "The current saturation helpers are ancillary-driven rather than full backend equilibrium solvers.",
            "They support the current endpoint-style saturation workflow but remain narrower than full CoolProp saturation behavior near difficult regions and unsupported fluids.",
        ]
    if ident.startswith("solver.Brent"):
        return [
            "The current scalar solver is part of the active thermodynamic path and is used in density and saturation inversion logic.",
            "Its correctness target is operational robustness for GOcoolprop flashes, not API parity with a public CoolProp solver function.",
        ]
    if ident.startswith("solver.Newton2D"):
        return [
            "The current two-variable Newton solver remains available as a utility, but it is no longer the primary path for the corrected supported flash behavior.",
            "Its current role is secondary relative to the newer bracketing-based flash strategy.",
        ]
    if ident == "flash.DensityTP":
        return [
            "The current implementation is a phase-aware single-phase `T,P -> rho` solve with explicit saturation-boundary rejection.",
            "Subcritical exact saturation-boundary `T,P` requests are intentionally rejected so callers must use `Q=0` or `Q=1`.",
            "This is closer to honest CoolProp-like behavior for the supported scope, but still narrower than full backend flash coverage.",
        ]
    if ident.startswith("flash.FlashTH"):
        return [
            "The current implementation supports single-phase and saturation-endpoint `T,H` behavior for the validated scope.",
            "Interior two-phase targets are explicit errors rather than pseudo-solutions.",
            "Endpoint enthalpy matches can return the ancillary saturation density directly.",
        ]
    if ident.startswith("flash.FlashPH") or ident.startswith("flash.FlashPS"):
        quantity = "enthalpy" if ident.endswith("PH") else "entropy"
        return [
            f"The current implementation solves supported `{quantity}` flashes by phase-aware temperature bracketing plus repeated single-phase density solves.",
            "Interior two-phase targets are explicit errors rather than approximate branch picks.",
            "This removes the older weak Newton-guess path from the main supported flash route, but full CoolProp phase coverage is still not present.",
        ]
    if ident == "props.PropSI":
        return [
            "The current implementation supports the documented core input pairs through the corrected state, saturation, and flash paths.",
            "Quality handling is currently endpoint-only: `Q=0` and `Q=1` are supported, while interior-quality inputs or outputs are explicit errors.",
            "Product-level behavior is closer to CoolProp for the validated core scope, but still narrower than full CoolProp coverage in unsupported fluids and interior two-phase regions.",
        ]
    if ident.startswith("transport."):
        return [
            "The current implementation supports only the transport model families that are explicitly implemented in the repo.",
            "Unsupported model families are expected to fail explicitly rather than returning silent placeholders.",
            "Transport parity remains narrower than thermodynamic parity and is still behind full CoolProp model coverage.",
        ]

    return [
        "The current implementation covers the supported local scope documented in the codebase.",
        "Any broader parity claim still depends on adding direct comparison evidence against regular CoolProp.",
    ]


def infer_task_list(item):
    ident = item["id"]
    tasks = [
        "Review this function against the current code and confirm the markdown signature, behavior summary, and supported-scope wording still match the implementation exactly.",
        "Add or refresh focused regression tests that exercise the documented nominal path, boundary path, and at least one explicit failure path.",
        "Add direct comparison points against regular CoolProp HEOS where that comparison is meaningful for this function.",
        "Tighten tolerances or expand benchmark coverage until this function has explicit evidence, not only intent, for the documented behavior.",
        "Update this markdown file again whenever the implementation scope changes so the spec does not drift ahead of or behind the code.",
    ]

    if ident == "core.State.Update":
        tasks.extend(
            [
                "Add direct backend-comparison checks for tau, delta, pressure, enthalpy, entropy, internal energy, Cv, and Cp at benchmark Water, Nitrogen, and Hydrogen points.",
                "Add cache-coherence tests that compare cached derivatives against fresh recomputation after repeated gas-to-liquid-to-gas update sequences.",
                "Add negative-input tests that prove invalid states cannot leak plausible-looking downstream property values.",
            ]
        )
    elif ident.startswith("core.State."):
        tasks.extend(
            [
                "Add central-difference validation for this property or derivative across dilute, dense, and near-critical single-phase states.",
                "Add identity-based checks tying this function to related thermodynamic quantities so parity is proven through multiple independent relationships.",
            ]
        )
    elif ident.startswith("core.IdealGasHelmholtz") or ident.startswith("core.ResidualHelmholtz"):
        tasks.extend(
            [
                "Add term-level reference cases derived from CoolProp-compatible coefficients and verify term value, first derivatives, second derivatives, and mixed derivative.",
                "Document parser assumptions for this term family so any coefficient normalization mismatch with CoolProp is visible immediately.",
            ]
        )
    elif ident == "core.NewState":
        tasks.extend(
            [
                "Audit every term family used by the current core fluids and confirm constructor coverage is explicit: implemented or hard error, never silent skip.",
                "Add constructor tests proving the built term inventory matches the intended CoolProp-style parser normalization for supported fluids.",
            ]
        )
    elif ident.startswith("fluid.LoadFluid"):
        tasks.extend(
            [
                "Add schema-fixture tests covering simple fluids, non-analytic fluids, and transport-rich fluids with variant JSON shapes.",
                "Document any known differences between GOcoolprop schema assumptions and CoolProp data conventions.",
            ]
        )
    elif ident.startswith("fluid.GetFluidFilename") or ident.startswith("fluid.ListAvailableFluids"):
        tasks.extend(
            [
                "Expand alias and normalization tests to include case changes, hyphen and spacing differences, and unsupported names.",
                "Cross-check the generated list or filename mapping against the current data directory contents.",
            ]
        )
    elif ident == "fluid.AncillaryCurve.Evaluate":
        tasks.extend(
            [
                "Add direct benchmark tables for pressure, saturated-liquid-density, and saturated-vapor-density ancillary evaluations.",
                "Add explicit unsupported-type tests so ancillary failures are honest and traceable.",
            ]
        )
    elif ident.startswith("saturation."):
        tasks.extend(
            [
                "Add round-trip saturation tests such as Psat(T)->Tsat(P) and endpoint density plausibility checks.",
                "Add out-of-range and near-critical tests to document where the ancillary-based approximation diverges from full CoolProp behavior.",
            ]
        )
    elif ident.startswith("solver."):
        tasks.extend(
            [
                "Add convergence-path tests that record iterations or residual history for representative thermodynamic objective shapes.",
                "Add difficult-case tests that prove explicit failure is returned for unbracketed, singular, or non-convergent problems.",
            ]
        )
    elif ident == "flash.DensityTP":
        tasks.extend(
            [
                "Add product-comparison cases for vapor, liquid, near-saturation single-phase, and exact saturation-boundary rejection behavior.",
                "Verify that branch selection follows the documented saturation-pressure logic for subcritical states.",
            ]
        )
    elif ident.startswith("flash."):
        tasks.extend(
            [
                "Add benchmark flash cases for gas-like, liquid-like, saturation-endpoint, and unsupported interior two-phase targets.",
                "Verify that returned states satisfy the defining residual equations and that wrong-branch convergence does not occur silently.",
                "Document clearly which flash regions still fall short of regular CoolProp coverage and must remain explicit errors.",
            ]
        )
    elif ident == "props.PropSI":
        tasks.extend(
            [
                "Build a per-pair benchmark matrix for Water, Nitrogen, and Hydrogen covering T,D; T,P; T,H; P,H; P,S; T,Q; and P,Q.",
                "Add explicit negative-path tests for unsupported output keys, unsupported pairs, unsupported fluids, interior-quality requests, and supercritical quality requests.",
                "Compare error semantics as well as numeric results so the high-level API behaves honestly when GOcoolprop cannot match CoolProp scope.",
            ]
        )
    elif ident.startswith("transport."):
        tasks.extend(
            [
                "Add low-density and dense-state benchmark points for each supported model family.",
                "Add explicit unsupported-model tests for ECS, hardcoded, or critical-enhancement paths that are still missing.",
                "Document the exact fluid/model combinations that can and cannot currently claim parity with CoolProp.",
            ]
        )

    return tasks

functions = [
    {
        "id": "fluid.LoadFluid",
        "file": "pkg/fluid/loader.go",
        "signature": "func LoadFluid(path string) (*FluidData, error)",
        "role": "Read one CoolProp JSON fluid file and materialize a validated FluidData model.",
        "must": [
            "Open the requested file and decode the JSON into FluidData.",
            "Validate required sections before returning success: fluid identity, EOS block, state data, and any ancillary or transport blocks referenced later.",
            "Reject malformed or unsupported schema content with explicit errors that name the missing field or unsupported structure.",
            "Preserve the original numeric values from the JSON without applying hidden normalization that belongs in EOS-term construction.",
        ],
        "validation": [
            "Unit test successful load for representative fluids such as Water, Nitrogen, and Hydrogen.",
            "Unit test schema failures for truncated JSON, missing EOS arrays, and transport variants known to differ across fluids.",
            "Compare loaded scalar metadata against CoolProp JSON source values for at least one fluid per schema family.",
        ],
        "coolprop": [
            "Regular CoolProp loads the fluid schema through its FluidLibrary parser and fails loudly on invalid structure.",
            "GOcoolprop should match the source-of-truth JSON content, not invent defaults that hide missing data.",
        ],
    },
    {
        "id": "fluid.LoadFluidByName",
        "file": "pkg/fluid/loader.go",
        "signature": "func LoadFluidByName(name, dataDir string) (*FluidData, error)",
        "role": "Resolve a user-facing fluid name or alias to the correct CoolProp JSON file and load it.",
        "must": [
            "Normalize aliases consistently through the registry before attempting direct filename fallback.",
            "Support canonical names, common aliases, and exact JSON filenames where appropriate.",
            "Return an explicit error when a name is unknown instead of silently mapping to the wrong fluid.",
            "Delegate all file parsing guarantees to LoadFluid so alias resolution and schema loading stay separate.",
        ],
        "validation": [
            "Test aliases such as water, H2O, nitrogen, N2, hydrogen, and refrigerant variants.",
            "Test unknown names and verify the error is actionable.",
            "Cross-check the resolved filename against the embedded CoolProp data filename for each alias family.",
        ],
        "coolprop": [
            "Regular CoolProp accepts fluid strings and aliases at the API layer; this function is the GOcoolprop equivalent of that name resolution step.",
            "Alias coverage does not need to be byte-for-byte identical to CoolProp, but incorrect aliasing is unacceptable.",
        ],
    },
    {
        "id": "fluid.GetFluidFilename",
        "file": "pkg/fluid/registry.go",
        "signature": "func GetFluidFilename(name string) (string, error)",
        "role": "Map a normalized fluid identifier to the exact JSON filename in the data directory.",
        "must": [
            "Normalize case, spaces, and dashes consistently.",
            "Return the exact CoolProp JSON filename for supported names and aliases.",
            "Reject unknown names without guessing.",
            "Stay synchronized with the actual contents of the data directory.",
        ],
        "validation": [
            "Table-driven test covering every alias family in the registry.",
            "Diff the registry values against the filenames present in data/.",
            "Test that invalid names do not resolve accidentally after normalization.",
        ],
        "coolprop": [
            "Regular CoolProp has broader name handling across backends and mixtures; this function only needs pure-fluid file resolution for now.",
        ],
    },
    {
        "id": "fluid.ListAvailableFluids",
        "file": "pkg/fluid/registry.go",
        "signature": "func ListAvailableFluids() []string",
        "role": "Expose the set of resolvable fluids that the library currently knows about.",
        "must": [
            "Return one representative identifier per fluid file rather than duplicate aliases.",
            "Prefer stable canonical names over arbitrary alias order if a canonical list is introduced.",
            "Remain consistent with the registry and the data directory.",
        ],
        "validation": [
            "Test uniqueness of the returned list.",
            "Test that common fluids such as Water, Nitrogen, and Hydrogen are present.",
            "Optionally compare count and filenames against data/.",
        ],
        "coolprop": [
            "Regular CoolProp exposes fluid listings through different APIs; this function is a local convenience contract.",
        ],
    },
    {
        "id": "fluid.AncillaryCurve.Evaluate",
        "file": "pkg/fluid/ancillary.go",
        "signature": "func (ac *AncillaryCurve) Evaluate(T float64) (float64, error)",
        "role": "Evaluate one ancillary correlation exactly as encoded in the fluid JSON.",
        "must": [
            "Use the ancillary reducing temperature from the correlation when provided.",
            "Evaluate the theta-series with the correct formula for the declared ancillary type.",
            "Support the currently used CoolProp ancillary families and reject unknown types explicitly rather than returning a silent zero.",
            "Preserve the units implied by the JSON reducing value and type.",
        ],
        "validation": [
            "Compare Psat, rhoL, and rhoV ancillary evaluations against CoolProp reference values at multiple temperatures per fluid.",
            "Verify behavior near Tmin, Tmax, and the critical approach.",
            "Test unknown ancillary types and confirm the function returns a hard failure through its `(float64, error)` signature.",
        ],
        "coolprop": [
            "Regular CoolProp evaluates ancillaries from the same model data but does not silently convert unknown types to zero.",
            "GOcoolprop should move to explicit failure semantics for unsupported ancillary types.",
        ],
    },
    {
        "id": "core.HelmholtzEnergy.Update",
        "file": "pkg/core/types.go",
        "signature": "func (h *HelmholtzEnergy) Update(tau, delta float64) (a, da_ddelta, da_dtau, d2a_ddelta2, d2a_dtau2, d2a_ddelta_dtau float64)",
        "role": "Aggregate total Helmholtz energy and derivatives from the alpha0 and alphar term lists.",
        "must": [
            "Evaluate every loaded term exactly once for the supplied tau and delta.",
            "Return mathematically consistent totals for alpha, first derivatives, second derivatives, and the mixed derivative.",
            "Keep alpha0 and alphar separable at the state level even if this aggregator also returns totals.",
            "Never hide unsupported terms by omission during construction of the term list.",
        ],
        "validation": [
            "Finite-difference test total derivatives against perturbations in tau and delta.",
            "Cross-check the sum of per-term values against a manually accumulated reference for sample fluids.",
            "Verify mixed-derivative symmetry numerically.",
        ],
        "coolprop": [
            "Regular CoolProp preserves ideal and residual contributions separately throughout the backend.",
            "GOcoolprop should expose enough structure that downstream properties can use canonical alpha0 and alphar formulas, not just totals.",
        ],
    },
    {
        "id": "core.NewState",
        "file": "pkg/core/state.go",
        "signature": "func NewState(f *fluid.FluidData) (*State, error)",
        "role": "Construct a thermodynamic state object and build the Helmholtz term model from the fluid JSON.",
        "must": [
            "Parse every supported alpha0 and alphar term family from EOS[0].",
            "Fail loudly when the fluid contains unsupported term families instead of silently skipping them.",
            "Apply the same term normalization and conversion rules that regular CoolProp applies during fluid parsing.",
            "Store enough structure to keep reducing-state data and separate ideal/residual derivatives available during property evaluation.",
        ],
        "validation": [
            "Test constructor success for fluids fully covered by implemented term families.",
            "Test constructor failure for fluids that still use unsupported term families, for example CarbonDioxide or any fluid that requires an unimplemented EOS family.",
            "Diff the constructed term inventory against the source JSON and CoolProp parser expectations.",
        ],
        "coolprop": [
            "Regular CoolProp transforms some JSON term families during parsing, for example FunctionT and enthalpy-entropy offset ideal-gas blocks.",
            "GOcoolprop cannot claim parity until NewState reproduces those parser-level semantics.",
        ],
    },
    {
        "id": "core.State.Update",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) Update(T, Rho float64)",
        "role": "Update the thermodynamic state for one temperature and molar density pair and cache Helmholtz derivatives.",
        "must": [
            "Use the EOS reducing state, not a critical-state fallback, to compute tau and delta.",
            "Evaluate and cache both total and separated alpha0 and alphar derivatives needed by all property methods.",
            "Reject non-physical inputs such as non-positive temperature or density.",
            "Keep cached pressure and derivative values internally consistent with the latest state.",
        ],
        "validation": [
            "Compare tau and delta against CoolProp backend values for sample fluids.",
            "Verify cached derivatives reproduce direct recomputation.",
            "Run property identity checks after repeated updates over dilute, dense, and near-critical single-phase states.",
        ],
        "coolprop": [
            "Regular CoolProp bases reduced variables on the reducing state from the EOS metadata.",
            "Any critical-state fallback should remain a last-resort compatibility path only when reducing data is genuinely absent.",
        ],
    },
    {
        "id": "core.State.Pressure",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) Pressure() float64",
        "role": "Return pressure from the canonical Helmholtz EOS relation.",
        "must": [
            "Use p = rho*R*T*(1 + delta*alphar_delta).",
            "Derive pressure from residual derivatives only, not from a shortcut that assumes a specific ideal-gas delta derivative form.",
            "Return the cached value only if it is guaranteed consistent with the current state.",
        ],
        "validation": [
            "Compare direct T,rho pressure against CoolProp HEOS values for benchmark points.",
            "Verify equivalence with the explicit alphar formula from cached separated derivatives.",
            "Check low-density ideal-gas limit and high-density single-phase states.",
        ],
        "coolprop": [
            "Regular CoolProp derives pressure from alphar only in its Helmholtz backend.",
            "GOcoolprop should avoid depending on the assumption that alpha0_delta is always 1/delta across all future ideal-gas term support.",
        ],
    },
    {
        "id": "core.State.MolarEntropy",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) MolarEntropy() float64",
        "role": "Return molar entropy from the Helmholtz EOS.",
        "must": [
            "Use s = R*(tau*alpha_tau - alpha).",
            "Ensure alpha and alpha_tau correspond to the same state and reducing basis.",
            "Remain consistent with the chosen reference state encoded by the fluid model.",
        ],
        "validation": [
            "Compare against CoolProp HEOS entropy at benchmark states.",
            "Verify thermodynamic identities such as ds/dT at constant rho = cv/T.",
        ],
        "coolprop": [
            "Regular CoolProp uses the same canonical formula but with full term coverage and correct reference-state handling.",
        ],
    },
    {
        "id": "core.State.MolarEnthalpy",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) MolarEnthalpy() float64",
        "role": "Return molar enthalpy from the Helmholtz EOS.",
        "must": [
            "Use h = R*T*(1 + tau*alphar_tau + tau*alpha0_tau + delta*alphar_delta).",
            "Avoid a collapsed-total shortcut that hides the explicit ideal and residual contributions.",
            "Stay consistent with the fluid reference state and reducing basis.",
        ],
        "validation": [
            "Compare against CoolProp HEOS enthalpy for gas, liquid, and near-critical single-phase points.",
            "Verify h = u + p/rho.",
        ],
        "coolprop": [
            "Regular CoolProp computes enthalpy with separated ideal and residual terms.",
            "GOcoolprop currently uses a total-alpha form that should be checked carefully against the canonical expression as term support broadens.",
        ],
    },
    {
        "id": "core.State.MolarInternalEnergy",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) MolarInternalEnergy() float64",
        "role": "Return molar internal energy from the Helmholtz EOS.",
        "must": [
            "Use u = R*T*tau*alpha_tau in a form consistent with separated alpha0 and alphar derivatives.",
            "Stay reference-state consistent with the fluid model.",
        ],
        "validation": [
            "Compare against CoolProp HEOS at representative states.",
            "Verify u = h - p/rho.",
        ],
        "coolprop": [
            "Regular CoolProp uses the same identity but with full EOS coverage.",
        ],
    },
    {
        "id": "core.State.Cv",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) Cv() float64",
        "role": "Return molar isochoric heat capacity from Helmholtz derivatives.",
        "must": [
            "Use cv = -R*tau^2*(alpha0_tautau + alphar_tautau).",
            "Handle all ideal and residual term families with analytic second derivatives.",
            "Remain stable across low-density and dense-fluid states.",
        ],
        "validation": [
            "Compare against CoolProp HEOS Cv values.",
            "Verify cv from finite-difference dU/dT at constant rho.",
        ],
        "coolprop": [
            "Regular CoolProp depends on complete second-derivative support; missing or inaccurate DTau2 implementations will show up here immediately.",
        ],
    },
    {
        "id": "core.State.Cp",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) Cp() float64",
        "role": "Return molar isobaric heat capacity from Helmholtz derivatives.",
        "must": [
            "Use the canonical CoolProp Helmholtz expression with residual delta derivatives and total tau second derivatives.",
            "Derive alphar_delta, alphar_deltadelta, and alphar_deltatau from separated caches rather than subtracting assumptions where possible.",
            "Guard against singular denominators near spinodal or critical behavior and return explicit errors upstream where necessary.",
        ],
        "validation": [
            "Compare against CoolProp HEOS Cp values in stable single-phase regions.",
            "Verify cp >= cv in stable regions and finite-difference dH/dT at constant P where flash support is reliable.",
        ],
        "coolprop": [
            "Regular CoolProp uses the same formula but with backend logic that is robust near difficult states.",
            "Any derivative mistake in alphar immediately contaminates Cp and flash Jacobians.",
        ],
    },
    {
        "id": "core.State.DPdT",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) DPdT() float64",
        "role": "Return the pressure derivative with respect to temperature at constant density for flash Jacobians.",
        "must": [
            "Derive from the canonical pressure relation and the reducing-state definition of tau.",
            "Use alphar_deltatau explicitly.",
            "Stay algebraically consistent with Pressure().",
        ],
        "validation": [
            "Compare analytic DPdT against central finite differences of Pressure().",
            "Use these checks across multiple stable states for each benchmark fluid.",
        ],
        "coolprop": [
            "Regular CoolProp exposes derivative APIs built from the same backend derivatives; GOcoolprop should match those numerically.",
        ],
    },
    {
        "id": "core.State.DPdRho",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) DPdRho() float64",
        "role": "Return the pressure derivative with respect to density at constant temperature for flash Jacobians and stability checks.",
        "must": [
            "Differentiate the canonical pressure relation with respect to density using reducing density from the EOS reducing state.",
            "Use residual delta derivatives explicitly.",
            "Be reliable enough to detect loss of monotonicity near problematic density regions.",
        ],
        "validation": [
            "Compare against central finite differences of Pressure().",
            "Check sign and magnitude in dilute-gas and compressed-liquid regions.",
        ],
        "coolprop": [
            "Regular CoolProp uses density derivatives in many solvers; inaccuracies here propagate directly into root finding.",
        ],
    },
    {
        "id": "core.State.DHdT",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) DHdT() float64",
        "role": "Return the enthalpy derivative with respect to temperature at constant density for flash Jacobians.",
        "must": [
            "Compute the actual derivative at constant density, not substitute Cp unless the algebra proves equivalence under the same constraint.",
            "Use separated derivatives consistently.",
        ],
        "validation": [
            "Compare against central finite differences of MolarEnthalpy() at constant rho.",
            "Compare against the expression used in flash Jacobians and document the distinction from Cp.",
        ],
        "coolprop": [
            "Regular CoolProp distinguishes different derivatives by variable constraints.",
            "The current GOcoolprop implementation returning Cp here is a likely mathematical mismatch.",
        ],
    },
    {
        "id": "core.State.DHdRho",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) DHdRho() float64",
        "role": "Return the enthalpy derivative with respect to density at constant temperature for flash Jacobians.",
        "must": [
            "Differentiate the canonical enthalpy expression with respect to density using reducing density from the EOS reducing state.",
            "Include all required tau-delta coupling terms if present in the exact derivative.",
        ],
        "validation": [
            "Compare against central finite differences of MolarEnthalpy() at constant temperature.",
            "Validate impact on P-H flash convergence versus CoolProp benchmarks.",
        ],
        "coolprop": [
            "Regular CoolProp flash solvers rely on exact thermodynamic derivatives; shortcut expressions must be proven before use.",
        ],
    },
    {
        "id": "core.State.DSdT",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) DSdT() float64",
        "role": "Return the entropy derivative with respect to temperature at constant density for flash Jacobians.",
        "must": [
            "Use the exact derivative implied by the Helmholtz entropy relation.",
            "Remain consistent with Cv/T only if the derivation holds under constant density for the implemented formulas.",
        ],
        "validation": [
            "Compare against central finite differences of MolarEntropy() at constant rho.",
            "Verify consistency with Cv/T numerically.",
        ],
        "coolprop": [
            "Regular CoolProp derivative consistency is critical for P-S flash robustness.",
        ],
    },
    {
        "id": "core.State.DSdRho",
        "file": "pkg/core/state.go",
        "signature": "func (s *State) DSdRho() float64",
        "role": "Return the entropy derivative with respect to density at constant temperature for flash Jacobians.",
        "must": [
            "Differentiate the entropy relation exactly using reducing density from the EOS reducing state.",
            "Use mixed derivatives consistently and preserve sign correctness.",
        ],
        "validation": [
            "Compare against central finite differences of MolarEntropy() at constant temperature.",
            "Verify effect on P-S flash root quality.",
        ],
        "coolprop": [
            "Regular CoolProp exposes the same thermodynamic information through backend derivatives; this function should match numerically.",
        ],
    },
]


def add_term_family(prefix, file_name, family, role_stub, notes):
    methods = [
        ("Term", "term value"),
        ("DDelta", "first delta derivative"),
        ("DTau", "first tau derivative"),
        ("DDelta2", "second delta derivative"),
        ("DTau2", "second tau derivative"),
        ("DDeltaTau", "mixed derivative"),
    ]
    for method, label in methods:
        functions.append(
            {
                "id": f"{prefix}.{method}",
                "file": file_name,
                "signature": f"func (t *{family}) {method}(tau, delta float64) float64",
                "role": f"Return the {label} for the {role_stub}.",
                "must": notes["must"](method),
                "validation": notes["validation"](method),
                "coolprop": notes["coolprop"](method),
            }
        )


add_term_family(
    "core.IdealGasHelmholtzLead",
    "pkg/core/alpha0.go",
    "IdealGasHelmholtzLead",
    "lead ideal-gas Helmholtz term",
    {
        "must": lambda method: [
            "Implement the exact analytic expression for the normalized lead ideal-gas term.",
            "Assume parser-level normalization from the CoolProp JSON representation has already been applied.",
        ],
        "validation": lambda method: [
            "Check against finite differences of the adjacent derivative level.",
            "Verify mixed-derivative symmetry where applicable.",
        ],
        "coolprop": lambda method: [
            "Regular CoolProp builds equivalent lead behavior during fluid parsing.",
        ],
    },
)

add_term_family(
    "core.IdealGasHelmholtzLogTau",
    "pkg/core/alpha0.go",
    "IdealGasHelmholtzLogTau",
    "log-tau ideal-gas Helmholtz term",
    {
        "must": lambda method: [
            "Implement the exact analytic expression for the normalized log-tau ideal-gas term.",
            "Preserve the expected zero delta dependence for this family.",
        ],
        "validation": lambda method: [
            "Check against finite differences of the adjacent derivative level.",
            "Verify exact zero behavior for delta derivatives and mixed derivative.",
        ],
        "coolprop": lambda method: [
            "Regular CoolProp uses equivalent ideal-gas log-tau behavior where present.",
        ],
    },
)

add_term_family(
    "core.IdealGasHelmholtzPlanckEinstein",
    "pkg/core/alpha0.go",
    "IdealGasHelmholtzPlanckEinstein",
    "normalized Planck-Einstein ideal-gas Helmholtz term",
    {
        "must": lambda method: [
            "Implement the exact analytic expression for the normalized Planck-Einstein form.",
            "Use numerically stable exponential handling at small and large tau.",
            "Only use this after parser-level conversion from the original CoolProp JSON representation where required.",
        ],
        "validation": lambda method: [
            "Check against finite differences of the adjacent derivative level.",
            "Compare term values against converted CoolProp reference calculations.",
        ],
        "coolprop": lambda method: [
            "Regular CoolProp supports multiple related Planck-Einstein ideal-gas families; GOcoolprop must not confuse them.",
        ],
    },
)

add_term_family(
    "core.ResidualHelmholtzPower",
    "pkg/core/alphar.go",
    "ResidualHelmholtzPower",
    "residual power or damped-power Helmholtz term",
    {
        "must": lambda method: [
            "Implement the exact analytic expression for both plain and damped forms.",
            "Preserve exponent signs and normalization exactly as defined by the model.",
        ],
        "validation": lambda method: [
            "Check against finite differences of the adjacent derivative level.",
            "Compare selected term values against CoolProp-style reference calculations.",
        ],
        "coolprop": lambda method: [
            "Regular CoolProp supports this family plus additional residual families not yet present in GOcoolprop.",
        ],
    },
)

add_term_family(
    "core.ResidualHelmholtzGaussian",
    "pkg/core/alphar.go",
    "ResidualHelmholtzGaussian",
    "Gaussian residual Helmholtz term",
    {
        "must": lambda method: [
            "Implement the exact analytic expression for the Gaussian form.",
            "Document and prove the algebra for second and mixed derivatives because Cp and flash stability are sensitive to them.",
        ],
        "validation": lambda method: [
            "Check against finite differences of the adjacent derivative level.",
            "Run focused tests near the Gaussian center where cancellation risk is higher.",
        ],
        "coolprop": lambda method: [
            "Regular CoolProp supports this family plus additional residual families not yet present in GOcoolprop.",
            "The current GOcoolprop file itself shows uncertainty around some Gaussian derivative algebra, so this family needs extra scrutiny.",
        ],
    },
)

functions.extend(
    [
        {
            "id": "saturation.Psat",
            "file": "pkg/saturation/saturation.go",
            "signature": "func Psat(f *fluid.FluidData, T float64) (float64, error)",
            "role": "Return saturation pressure from the fluid ancillary correlation.",
            "must": [
                "Respect the ancillary valid temperature range.",
                "Use the exact ancillary formula for the fluid.",
                "Return explicit errors outside supported range instead of extrapolating invisibly.",
            ],
            "validation": [
                "Compare against CoolProp saturation pressure at multiple temperatures per fluid.",
                "Check edge behavior near triple and critical temperatures.",
            ],
            "coolprop": [
                "Regular CoolProp can use ancillaries as helpers but final saturation behavior is integrated with phase logic; GOcoolprop should not overstate ancillary-only accuracy.",
            ],
        },
        {
            "id": "saturation.Tsat",
            "file": "pkg/saturation/saturation.go",
            "signature": "func Tsat(f *fluid.FluidData, P float64) (float64, error)",
            "role": "Invert the saturation pressure ancillary to find saturation temperature.",
            "must": [
                "Bracket the valid ancillary temperature range correctly.",
                "Verify pressure range before solving.",
                "Use a robust scalar solver and return explicit errors when inversion fails.",
            ],
            "validation": [
                "Compare against CoolProp Tsat values at multiple pressures.",
                "Verify Psat(Tsat(P)) reproduces the target pressure within tolerance.",
            ],
            "coolprop": [
                "Regular CoolProp saturation calls are backend-aware; ancillary inversion is only acceptable here when documented as the chosen approximation path.",
            ],
        },
        {
            "id": "saturation.RhoL",
            "file": "pkg/saturation/saturation.go",
            "signature": "func RhoL(f *fluid.FluidData, T float64) (float64, error)",
            "role": "Return saturated-liquid molar density from the fluid ancillary correlation.",
            "must": [
                "Use the liquid-density ancillary exactly and enforce its validity range.",
            ],
            "validation": [
                "Compare against CoolProp saturated liquid density at multiple temperatures.",
            ],
            "coolprop": [
                "Regular CoolProp may use more integrated saturation logic; this function is ancillary-based and should be documented as such.",
            ],
        },
        {
            "id": "saturation.RhoV",
            "file": "pkg/saturation/saturation.go",
            "signature": "func RhoV(f *fluid.FluidData, T float64) (float64, error)",
            "role": "Return saturated-vapor molar density from the fluid ancillary correlation.",
            "must": [
                "Use the vapor-density ancillary exactly and enforce its validity range.",
            ],
            "validation": [
                "Compare against CoolProp saturated vapor density at multiple temperatures.",
            ],
            "coolprop": [
                "Regular CoolProp may use more integrated saturation logic; this function is ancillary-based and should be documented as such.",
            ],
        },
        {
            "id": "solver.Brent",
            "file": "pkg/solver/brent.go",
            "signature": "func Brent(f func(float64) float64, a, b float64, tol float64) (float64, error)",
            "role": "Provide a robust scalar root finder for saturation inversion and one-dimensional flash searches.",
            "must": [
                "Require a valid bracket unless an endpoint is already a root.",
                "Converge reliably for well-behaved monotonic thermodynamic objectives.",
                "Return explicit failure on unbracketed roots or iteration exhaustion.",
            ],
            "validation": [
                "Unit test on analytic roots and monotonic nonlinear functions.",
                "Use thermodynamic objectives from Psat inversion and density solves as integration tests.",
            ],
            "coolprop": [
                "Regular CoolProp uses robust numeric solvers internally; this local implementation must be held to the same stability standard even if not the same exact algorithm.",
            ],
        },
        {
            "id": "solver.Newton2D",
            "file": "pkg/solver/newton2d.go",
            "signature": "func Newton2D(funcJS func(x, y float64) (f1, f2, J11, J12, J21, J22 float64), x0, y0 float64, tol float64, maxIter int) (x, y float64, err error)",
            "role": "Provide a two-variable Newton solver for coupled flash equations.",
            "must": [
                "Handle residual and Jacobian evaluation robustly.",
                "Report singular Jacobians and divergence clearly.",
                "Gain damping or line-search support if raw Newton is not reliable enough for thermodynamic flashes.",
            ],
            "validation": [
                "Unit test on analytic systems.",
                "Integration test on P-H and P-S flashes against CoolProp benchmarks.",
                "Record convergence rates and failure modes near difficult regions.",
            ],
            "coolprop": [
                "Regular CoolProp flash logic is more phase-aware than the current GOcoolprop solver stack.",
                "A plain undamped Newton method is unlikely to be sufficient for full parity.",
            ],
        },
        {
            "id": "flash.DensityTP",
            "file": "pkg/flash/single_phase.go",
            "signature": "func DensityTP(fluidData *fluid.FluidData, T, P float64) (float64, error)",
            "role": "Solve for molar density at fixed temperature and pressure within the supported single-phase and saturation-boundary contract.",
            "must": [
                "Use a robust EOS pressure root solve over density rather than ideal-gas or compressed-liquid shortcuts.",
                "Infer vapor versus liquid branch from saturation pressure when the state is subcritical and off the saturation line.",
                "Reject exact saturation-boundary T,P requests and require the caller to specify Q=0 or Q=1 instead.",
                "Return explicit errors when no stable single-phase density root can be bracketed.",
            ],
            "validation": [
                "Benchmark against CoolProp HEOS for gas-like and liquid-like T,P states for Water, Nitrogen, and Hydrogen.",
                "Verify that the returned density reproduces the target pressure when passed back through State.Update and Pressure().",
                "Verify subcritical saturation-boundary T,P requests fail explicitly rather than picking an arbitrary branch.",
            ],
            "coolprop": [
                "Regular CoolProp performs phase-aware T,P flashes internally and does not silently invent a branch at exact saturation without additional information.",
                "GOcoolprop now follows that stricter contract for the supported scope, but still relies on local EOS root scanning rather than the full CoolProp backend logic.",
            ],
        },
        {
            "id": "flash.FlashTH",
            "file": "pkg/flash/th_flash.go",
            "signature": "func FlashTH(fluidData *fluid.FluidData, T, H_target float64) (float64, error)",
            "role": "Solve for density at fixed temperature and molar enthalpy.",
            "must": [
                "Search for all physically possible density roots at the supplied temperature and choose the branch with explicit phase-aware logic.",
                "Return exact saturation-endpoint densities when the target enthalpy matches the ancillary liquid or vapor endpoint at the supplied temperature.",
                "Reject unsupported interior two-phase targets explicitly rather than returning a pseudo-single-phase density.",
            ],
            "validation": [
                "Compare density results against CoolProp HEOS for gas, liquid, and saturation-endpoint cases.",
                "Check that H(T, rho_result) reproduces the target enthalpy within tolerance.",
            ],
            "coolprop": [
                "Regular CoolProp flash behavior is phase-aware across single-phase and two-phase states.",
                "GOcoolprop is now closer for the supported scope because it handles endpoint branch selection explicitly, but it still rejects unsupported two-phase interior enthalpy flashes.",
            ],
        },
        {
            "id": "flash.FlashPH",
            "file": "pkg/flash/ph_flash.go",
            "signature": "func FlashPH(fluidData *fluid.FluidData, P_target, H_target float64) (float64, float64, error)",
            "role": "Solve for temperature and density at fixed pressure and molar enthalpy.",
            "must": [
                "Use phase-aware temperature bracketing and a single-phase density solve instead of relying on a fragile coupled Newton guess.",
                "Infer the correct liquid or vapor branch from saturation endpoint enthalpies when the pressure is subcritical.",
                "Detect unsupported two-phase interior targets explicitly and fail rather than returning a misleading state.",
            ],
            "validation": [
                "Benchmark against CoolProp HEOS for gas, compressed liquid, and near-saturation cases.",
                "Verify both pressure and enthalpy residuals at the returned state.",
            ],
            "coolprop": [
                "Regular CoolProp uses a more mature backend-wide flash implementation with broader phase coverage.",
                "GOcoolprop is closer now because it no longer depends on the older weak Newton-guess path for the supported single-phase scope.",
            ],
        },
        {
            "id": "flash.FlashPS",
            "file": "pkg/flash/ps_flash.go",
            "signature": "func FlashPS(fluidData *fluid.FluidData, P_target, S_target float64) (float64, float64, error)",
            "role": "Solve for temperature and density at fixed pressure and molar entropy.",
            "must": [
                "Use phase-aware temperature bracketing and a single-phase density solve rather than a weak coupled Newton guess.",
                "Infer the correct liquid or vapor branch from saturation endpoint entropies when the pressure is subcritical.",
                "Detect unsupported two-phase interior targets explicitly.",
            ],
            "validation": [
                "Benchmark against CoolProp HEOS for gas, compressed liquid, and near-saturation cases.",
                "Verify both pressure and entropy residuals at the returned state.",
            ],
            "coolprop": [
                "Regular CoolProp P-S flashes are sensitive to entropy derivative quality and phase logic.",
                "GOcoolprop is closer now because it no longer relies on the earlier rough Newton path for the supported single-phase scope, but full two-phase parity is still missing.",
            ],
        },
        {
            "id": "props.PropSI",
            "file": "pkg/props/props.go",
            "signature": "func PropSI(output, name1 string, val1 float64, name2 string, val2 float64, fluidName string) (float64, error)",
            "role": "Expose a CoolProp-like high-level API for property, flash, saturation, and transport queries.",
            "must": [
                "Normalize input names and route each supported input pair to the correct thermodynamic path.",
                "Preserve CoolProp-like semantics for T,D; T,P; P,H; P,S; T,H; T,Q; and P,Q where supported.",
                "Return explicit errors for unsupported input pairs, unsupported fluids, unsupported term coverage, undefined outputs such as supercritical quality, and unsupported interior-quality requests.",
                "Use the dedicated phase-aware T,P density solve rather than compressed-liquid or arbitrary branch shortcuts.",
                "Support Q only as a saturation-endpoint concept in the current scope: Q=0 and Q=1 are valid, interior two-phase Q requests are explicit errors.",
            ],
            "validation": [
                "Build an end-to-end benchmark matrix against regular CoolProp HEOS for representative fluids and state regions.",
                "Verify each input pair both for returned value accuracy and for branch correctness near saturation, including explicit endpoint-only Q behavior.",
                "Test transport outputs separately from thermodynamic outputs.",
            ],
            "coolprop": [
                "Regular CoolProp PropSI is the behavioral reference product for this function.",
                "GOcoolprop is closer now because phase selection and T,P/P,H/P,S/T,H behavior use stricter flash rules than before, but it still differs in unsupported EOS coverage and the lack of general interior two-phase flash support.",
            ],
        },
        {
            "id": "transport.Viscosity",
            "file": "pkg/transport/viscosity.go",
            "signature": "func Viscosity(f *fluid.FluidData, T, Rho float64) (float64, error)",
            "role": "Return dynamic viscosity by combining dilute and residual contributions from the fluid transport model.",
            "must": [
                "Dispatch correctly across supported transport model families.",
                "Combine dilute, residual, and later critical-enhancement or hardcoded contributions as required by the fluid data.",
                "Return explicit unsupported-model errors when the fluid requires a transport path that is not yet implemented.",
            ],
            "validation": [
                "Compare against CoolProp viscosity values across dilute gas and dense-fluid states for covered fluids.",
                "Check model-family-specific unit consistency.",
            ],
            "coolprop": [
                "Regular CoolProp transport support is broader and often includes fluid-specific hardcoded models.",
                "GOcoolprop should not claim transport parity until those models are explicitly covered.",
            ],
        },
        {
            "id": "transport.ViscosityDilute",
            "file": "pkg/transport/viscosity.go",
            "signature": "func ViscosityDilute(f *fluid.FluidData, T float64) (float64, error)",
            "role": "Return the dilute-gas viscosity contribution from the declared transport model.",
            "must": [
                "Implement each supported dilute viscosity family exactly with correct unit conversions.",
                "Validate collision-integral evaluation and parameter normalization against CoolProp data conventions.",
            ],
            "validation": [
                "Compare against CoolProp low-density viscosity values for fluids using each supported dilute model.",
                "Cross-check units by reproducing literature or CoolProp test values.",
            ],
            "coolprop": [
                "Regular CoolProp model support is broader; this function should fail loudly on unsupported dilute model types.",
            ],
        },
        {
            "id": "transport.ViscosityResidual",
            "file": "pkg/transport/viscosity.go",
            "signature": "func ViscosityResidual(f *fluid.FluidData, T, Rho float64) (float64, error)",
            "role": "Return the density-dependent viscosity contribution from the declared transport model.",
            "must": [
                "Implement each supported higher-order viscosity family exactly, including reducing-state choices declared by the transport model.",
                "Preserve units and exponent signs exactly as defined in the model.",
            ],
            "validation": [
                "Compare against CoolProp dense-state viscosity values for covered fluids.",
                "Check sensitivity near high density where exponential terms dominate.",
            ],
            "coolprop": [
                "Regular CoolProp often includes additional terms or hardcoded paths not yet present here.",
            ],
        },
        {
            "id": "transport.Conductivity",
            "file": "pkg/transport/conductivity.go",
            "signature": "func Conductivity(f *fluid.FluidData, T, Rho float64) (float64, error)",
            "role": "Return thermal conductivity by combining dilute, residual, and eventually critical contributions.",
            "must": [
                "Compile and execute correctly with explicit model dispatch and explicit unsupported-model failures.",
                "Dispatch across supported conductivity model families.",
                "Combine dilute, residual, and later critical-enhancement terms with correct units.",
            ],
            "validation": [
                "Compare against CoolProp conductivity values for covered fluids now that the file is build-clean again.",
                "Test dilute and dense regions separately.",
            ],
            "coolprop": [
                "Regular CoolProp transport coverage is more complete; GOcoolprop is still not at model-family parity in this area even though the file now builds.",
            ],
        },
        {
            "id": "transport.ConductivityDilute",
            "file": "pkg/transport/conductivity.go",
            "signature": "func ConductivityDilute(f *fluid.FluidData, T float64) (float64, error)",
            "role": "Return the dilute thermal conductivity contribution from the declared transport model.",
            "must": [
                "Implement each supported dilute conductivity family exactly with correct polynomial or rational structure.",
                "Preserve units and model normalization from the fluid JSON.",
            ],
            "validation": [
                "Compare against CoolProp low-density conductivity values for supported fluids.",
                "Verify polynomial and rational branches independently.",
            ],
            "coolprop": [
                "Regular CoolProp supports these model families plus additional fluid-specific paths.",
            ],
        },
        {
            "id": "transport.ConductivityResidual",
            "file": "pkg/transport/conductivity.go",
            "signature": "func ConductivityResidual(f *fluid.FluidData, T, Rho float64) (float64, error)",
            "role": "Return the density-dependent thermal conductivity contribution from the declared transport model.",
            "must": [
                "Implement each supported residual conductivity family exactly.",
                "Use the reducing quantities required by the transport model rather than assuming critical-state scaling where the model defines otherwise.",
            ],
            "validation": [
                "Compare against CoolProp dense-state conductivity values for supported fluids.",
                "Check high-density sensitivity and sign behavior.",
            ],
            "coolprop": [
                "Regular CoolProp also adds critical enhancement for some fluids; GOcoolprop needs explicit scope labeling when that is missing.",
            ],
        },
        {
            "id": "transport.SurfaceTension",
            "file": "pkg/transport/surface_tension.go",
            "signature": "func SurfaceTension(f *fluid.FluidData, T float64) (float64, error)",
            "role": "Return surface tension from the fluid surface-tension correlation at saturation conditions.",
            "must": [
                "Resolve the correct data source when the fluid stores surface tension in transport or ancillary blocks.",
                "Evaluate the declared correlation only below the critical temperature.",
                "Reject missing data or out-of-range states explicitly.",
            ],
            "validation": [
                "Compare against CoolProp surface tension values for supported fluids such as Water.",
                "Verify correct failure above Tc and for fluids with no data.",
            ],
            "coolprop": [
                "Regular CoolProp treats surface tension as a saturation-only property; GOcoolprop should keep the same behavioral contract.",
            ],
        },
    ]
)

seen = set()
for item in functions:
    if item["id"] in seen:
        raise RuntimeError(f"duplicate function id: {item['id']}")
    seen.add(item["id"])

for item in functions:
    lines = [
        f"# {item['id']}",
        "",
        f"- Source: `{item['file']}`",
        f"- Signature: `{item['signature']}`",
        "",
        "## What This Function Needs To Do",
        "",
        item["role"],
        "",
    ]
    lines.extend(f"- {value}" for value in item["must"])
    lines.extend(
        [
            "",
            "## Current GOcoolprop Scope",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in infer_current_scope(item))
    lines.extend(
        [
            "",
            "## Validation",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in item["validation"])
    lines.extend(
        [
            "",
            "## Test Conditions",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in infer_test_conditions(item))
    lines.extend(
        [
            "",
            "## Boundary Conditions",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in infer_boundary_conditions(item))
    lines.extend(
        [
            "",
            "## Difference With Regular CoolProp",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in item["coolprop"])
    lines.extend(
        [
            "",
            "## Gap Analysis vs CoolProp",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in infer_gap_analysis(item))
    lines.extend(
        [
            "",
            "## Task List",
            "",
        ]
    )
    lines.extend(f"- {value}" for value in infer_task_list(item))
    lines.extend(
        [
            "",
            "## Implementation Notes",
            "",
            "- This document now distinguishes between the corrected current GOcoolprop scope and the broader target behavior implied by regular CoolProp parity.",
            "- If this function depends on unsupported EOS or transport term families, that dependency should be surfaced explicitly at runtime or construction time.",
        ]
    )
    slug = re.sub(r"[^A-Za-z0-9_.-]+", "_", item["id"]) + ".md"
    (out_dir / slug).write_text("\n".join(lines), encoding="utf-8")

by_pkg = {}
for item in functions:
    pkg = item["id"].split(".")[0]
    by_pkg.setdefault(pkg, []).append(item)

readme = [
    "# Function Specification Matrix",
    "",
    "This folder describes, per library function, what GOcoolprop needs to do to reach a reliable CoolProp-like target state.",
    "",
    "Coverage rules:",
    "",
    "- One document per non-test function or method in `pkg/`.",
    "- Each document describes the target responsibility, validation path, explicit test conditions, explicit boundary conditions, and the relevant difference with regular CoolProp.",
    "- These are implementation targets, not statements that the current code is already correct.",
    "",
    f"Total documented functions: `{len(functions)}`",
    "",
]

for pkg in sorted(by_pkg):
    readme.append(f"## {pkg}")
    readme.append("")
    for item in sorted(by_pkg[pkg], key=lambda x: x["id"]):
        slug = re.sub(r"[^A-Za-z0-9_.-]+", "_", item["id"]) + ".md"
        readme.append(f"- [{item['id']}]({slug})")
    readme.append("")

(out_dir / "README.md").write_text("\n".join(readme), encoding="utf-8")
print(f"Wrote {len(functions)} function docs to {out_dir}")
