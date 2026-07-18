from __future__ import annotations

import json
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DATA_DIR = ROOT / "data"
OUT_DIR = ROOT / "snippets" / "testcode"
OUT_DIR.mkdir(parents=True, exist_ok=True)

SUPPORTED_ALPHA0 = {
    "IdealGasHelmholtzLead",
    "IdealGasHelmholtzLogTau",
    "IdealGasHelmholtzPlanckEinstein",
    "IdealGasHelmholtzPower",
    "IdealGasHelmholtzPlanckEinsteinFunctionT",
}

SUPPORTED_ALPHAR = {
    "ResidualHelmholtzPower",
    "ResidualHelmholtzGaussian",
    "ResidualHelmholtzNonAnalytic",
}


def safe_ident(name: str) -> str:
    ident = re.sub(r"[^A-Za-z0-9]+", "_", name)
    if ident and ident[0].isdigit():
        ident = "F_" + ident
    return ident


def rel_tol(fluid_name: str) -> float:
    if fluid_name in {"Water", "Nitrogen", "Hydrogen"}:
        return 1e-3
    return 5e-3


def choose_state(fluid_data: dict) -> tuple[str, dict]:
    eos_state = fluid_data["EOS"][0]["STATES"]
    if "reducing" in eos_state:
        return "reducing", eos_state["reducing"]
    return "critical", fluid_data["STATES"]["critical"]


def supported_status(fluid_data: dict) -> tuple[bool, list[str]]:
    missing: list[str] = []
    for term in fluid_data["EOS"][0].get("alpha0", []):
        term_type = term["type"]
        if term_type not in SUPPORTED_ALPHA0:
            missing.append(term_type)
    for term in fluid_data["EOS"][0].get("alphar", []):
        term_type = term["type"]
        if term_type not in SUPPORTED_ALPHAR:
            missing.append(term_type)
    return len(missing) == 0, sorted(set(missing))


def make_supported_file(fluid_name: str, filename: str, fluid_data: dict, missing: list[str]) -> str:
    test_name = safe_ident(fluid_name)
    state_name, state = choose_state(fluid_data)
    t = state["T"]
    rho = state["rhomolar"]
    p = state["p"]
    h = state["hmolar"]
    s = state["smolar"]
    tol = rel_tol(fluid_name)

    lines = [
        "package snippets",
        "",
        "import (",
        '\t"GOcoolprop/pkg/core"',
        '\t"GOcoolprop/pkg/fluid"',
        '\t"GOcoolprop/pkg/saturation"',
        '\t"math"',
        '\t"testing"',
        ")",
        "",
        "func snippetRelErr(a, b float64) float64 {",
        "\tscale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))",
        "\treturn math.Abs(a-b) / scale",
        "}",
        "",
        f"func Test{test_name}_LoadAndReferenceState(t *testing.T) {{",
        f'\tf, err := fluid.LoadFluid("../../data/{filename}")',
        "\tif err != nil {",
        f'\t\tt.Fatalf("load {fluid_name}: %v", err)',
        "\t}",
        "\tstate, err := core.NewState(f)",
        "\tif err != nil {",
        f'\t\tt.Fatalf("new state {fluid_name}: %v", err)',
        "\t}",
        f"\tstate.Update({t}, {rho})",
        f"\tif snippetRelErr(state.Pressure(), {p}) > {tol} {{",
        f'\t\tt.Fatalf("{fluid_name} pressure mismatch: got=%v want=%v", state.Pressure(), {p})',
        "\t}",
        f"\tif snippetRelErr(state.MolarEnthalpy(), {h}) > {tol} {{",
        f'\t\tt.Fatalf("{fluid_name} enthalpy mismatch: got=%v want=%v", state.MolarEnthalpy(), {h})',
        "\t}",
        f"\tif snippetRelErr(state.MolarEntropy(), {s}) > {tol} {{",
        f'\t\tt.Fatalf("{fluid_name} entropy mismatch: got=%v want=%v", state.MolarEntropy(), {s})',
        "\t}",
        "}",
        "",
    ]

    anc = fluid_data.get("ANCILLARIES", {})
    ps = anc.get("pS")
    if ps and "Tmin" in ps and "Tmax" in ps:
        sat_t = round((ps["Tmin"] + ps["Tmax"]) / 2.0, 6)
        lines.extend(
            [
                f"func Test{test_name}_SaturationRoundTrip(t *testing.T) {{",
                f'\tf, err := fluid.LoadFluid("../../data/{filename}")',
                "\tif err != nil {",
                f'\t\tt.Fatalf("load {fluid_name}: %v", err)',
                "\t}",
                f"\tpSat, err := saturation.Psat(f, {sat_t})",
                "\tif err != nil {",
                f'\t\tt.Fatalf("{fluid_name} Psat: %v", err)',
                "\t}",
                "\ttSat, err := saturation.Tsat(f, pSat)",
                "\tif err != nil {",
                f'\t\tt.Fatalf("{fluid_name} Tsat: %v", err)',
                "\t}",
                "\tif math.Abs(tSat-"+
                f"{sat_t}) > 1e-4 {{",
                f'\t\tt.Fatalf("{fluid_name} saturation round-trip mismatch: got=%v want=%v", tSat, {sat_t})',
                "\t}",
                "\tif _, err := saturation.RhoL(f, tSat); err != nil {",
                f'\t\tt.Fatalf("{fluid_name} rhoL: %v", err)',
                "\t}",
                "\tif _, err := saturation.RhoV(f, tSat); err != nil {",
                f'\t\tt.Fatalf("{fluid_name} rhoV: %v", err)',
                "\t}",
                "}",
                "",
            ]
        )

    return "\n".join(lines)


def make_blocked_file(fluid_name: str, filename: str, missing: list[str]) -> str:
    test_name = safe_ident(fluid_name)
    missing_literal = ", ".join(missing)
    missing_go = ", ".join(f'"{m}"' for m in missing)
    return "\n".join(
        [
            "package snippets",
            "",
            "import (",
            '\t"GOcoolprop/pkg/core"',
            '\t"GOcoolprop/pkg/fluid"',
            '\t"strings"',
            '\t"testing"',
            ")",
            "",
            f"func Test{test_name}_UnsupportedToday(t *testing.T) {{",
            f'\tf, err := fluid.LoadFluid("../../data/{filename}")',
            "\tif err != nil {",
            f'\t\tt.Fatalf("load {fluid_name}: %v", err)',
            "\t}",
            "\t_, err = core.NewState(f)",
            "\tif err == nil {",
            f'\t\tt.Fatalf("expected {fluid_name} to fail until missing terms are implemented")',
            "\t}",
            f"\tmissing := []string{{{missing_go}}}",
            "\tfor _, term := range missing {",
            "\t\tif strings.Contains(err.Error(), term) {",
            "\t\t\treturn",
            "\t\t}",
            "\t}",
            f'\t\tt.Fatalf("unexpected error for {fluid_name}: %v; expected one of [{missing_literal}]", err)',
            "}",
            "",
        ]
    )


def main() -> None:
    files = sorted(DATA_DIR.glob("*.json"))
    readme_lines = [
        "# Fluid Testcode Snippets",
        "",
        "This folder contains one Go `_test.go` snippet per fluid in `data/`.",
        "",
        "Generation rules:",
        "",
        "- If the fluid is supported by the current `core.NewState` term coverage, the snippet contains a reference-state test and a saturation round-trip test when ancillaries are present.",
        "- If the fluid is not yet supported, the snippet contains a negative-path test that asserts `core.NewState` fails explicitly with the missing term family.",
        "- These snippets are meant as copyable starting points for real package tests.",
        "",
    ]

    supported_count = 0
    blocked_count = 0

    for path in files:
        fluid_data = json.loads(path.read_text(encoding="utf-8"))
        fluid_name = fluid_data["INFO"]["NAME"]
        ok, missing = supported_status(fluid_data)
        out_path = OUT_DIR / f"{path.stem}_test.go"
        if ok:
            supported_count += 1
            content = make_supported_file(fluid_name, path.name, fluid_data, missing)
        else:
            blocked_count += 1
            content = make_blocked_file(fluid_name, path.name, missing)
        out_path.write_text(content + "\n", encoding="utf-8")
        readme_lines.append(f"- [{out_path.name}]({out_path.name})")

    readme_lines.extend(
        [
            "",
            f"Supported snippet files: `{supported_count}`",
            f"Blocked snippet files: `{blocked_count}`",
        ]
    )
    (OUT_DIR / "README.md").write_text("\n".join(readme_lines) + "\n", encoding="utf-8")
    print(f"Generated {len(files)} fluid test snippets in {OUT_DIR}")


if __name__ == "__main__":
    main()
