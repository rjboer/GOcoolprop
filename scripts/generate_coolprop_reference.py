import json
from pathlib import Path

from CoolProp.CoolProp import PropsSI


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "pkg" / "validation" / "testdata" / "coolprop_core_reference.json"


STATE_POINTS = [
    {"name": "water_liquid_300K", "fluid": "Water", "T": 300.0, "rho": 55317.3},
    {"name": "nitrogen_gas_300K", "fluid": "Nitrogen", "T": 300.0, "rho": 40.6},
    {"name": "hydrogen_gas_300K", "fluid": "Hydrogen", "T": 300.0, "rho": 40.6},
    {"name": "water_liquid_450K", "fluid": "Water", "T": 450.0, "rho": 50000.0},
    {"name": "water_gas_500K", "fluid": "Water", "T": 500.0, "rho": 100.0},
    {"name": "nitrogen_dense_150K", "fluid": "Nitrogen", "T": 150.0, "rho": 1000.0},
    {"name": "hydrogen_dense_60K", "fluid": "Hydrogen", "T": 60.0, "rho": 2000.0},
    {"name": "hydrogen_gas_80K", "fluid": "Hydrogen", "T": 80.0, "rho": 1000.0},
]

TP_POINTS = [
    {"name": "water_tp_300K_1atm", "fluid": "Water", "T": 300.0, "P": 101325.0},
    {"name": "nitrogen_tp_300K_1atm", "fluid": "Nitrogen", "T": 300.0, "P": 101325.0},
    {"name": "hydrogen_tp_300K_1atm", "fluid": "Hydrogen", "T": 300.0, "P": 101325.0},
    {"name": "water_tp_450K_dense", "fluid": "Water", "T": 450.0, "P": 17459072.21246576},
    {"name": "water_tp_500K_gas", "fluid": "Water", "T": 500.0, "P": 408586.4461412298},
    {"name": "nitrogen_tp_150K_dense", "fluid": "Nitrogen", "T": 150.0, "P": 1160978.1210144467},
    {"name": "hydrogen_tp_60K_dense", "fluid": "Hydrogen", "T": 60.0, "P": 956171.8440181161},
    {"name": "hydrogen_tp_80K_gas", "fluid": "Hydrogen", "T": 80.0, "P": 659401.856064851},
]

SAT_POINTS = [
    {"name": "water_sat_300K_q0", "fluid": "Water", "T": 300.0, "Q": 0.0},
    {"name": "water_sat_300K_q1", "fluid": "Water", "T": 300.0, "Q": 1.0},
    {"name": "water_sat_1atm_q0", "fluid": "Water", "P": 101325.0, "Q": 0.0},
    {"name": "water_sat_1atm_q1", "fluid": "Water", "P": 101325.0, "Q": 1.0},
]


def heos(fluid: str) -> str:
    return f"HEOS::{fluid}"


def add_td_outputs(entry: dict) -> dict:
    fluid = heos(entry["fluid"])
    T = entry["T"]
    rho = entry["rho"]
    entry.update(
        {
            "P": PropsSI("P", "T", T, "DMOLAR", rho, fluid),
            "H": PropsSI("HMOLAR", "T", T, "DMOLAR", rho, fluid),
            "S": PropsSI("SMOLAR", "T", T, "DMOLAR", rho, fluid),
            "U": PropsSI("UMOLAR", "T", T, "DMOLAR", rho, fluid),
            "Cv": PropsSI("CVMOLAR", "T", T, "DMOLAR", rho, fluid),
            "Cp": PropsSI("CPMOLAR", "T", T, "DMOLAR", rho, fluid),
        }
    )
    return entry


def add_tp_outputs(entry: dict) -> dict:
    fluid = heos(entry["fluid"])
    T = entry["T"]
    P = entry["P"]
    entry.update(
        {
            "rho": PropsSI("DMOLAR", "T", T, "P", P, fluid),
            "H": PropsSI("HMOLAR", "T", T, "P", P, fluid),
            "S": PropsSI("SMOLAR", "T", T, "P", P, fluid),
            "U": PropsSI("UMOLAR", "T", T, "P", P, fluid),
            "Cv": PropsSI("CVMOLAR", "T", T, "P", P, fluid),
            "Cp": PropsSI("CPMOLAR", "T", T, "P", P, fluid),
        }
    )
    return entry


def add_sat_outputs(entry: dict) -> dict:
    fluid = heos(entry["fluid"])
    q = entry["Q"]
    if "T" in entry:
        T = entry["T"]
        entry.update(
            {
                "P_ref": PropsSI("P", "T", T, "Q", q, fluid),
                "rho": PropsSI("DMOLAR", "T", T, "Q", q, fluid),
                "H": PropsSI("HMOLAR", "T", T, "Q", q, fluid),
                "S": PropsSI("SMOLAR", "T", T, "Q", q, fluid),
            }
        )
    else:
        P = entry["P"]
        entry.update(
            {
                "T_ref": PropsSI("T", "P", P, "Q", q, fluid),
                "rho": PropsSI("DMOLAR", "P", P, "Q", q, fluid),
                "H": PropsSI("HMOLAR", "P", P, "Q", q, fluid),
                "S": PropsSI("SMOLAR", "P", P, "Q", q, fluid),
            }
        )
    return entry


def main() -> None:
    dataset = {
        "state_points": [add_td_outputs(dict(point)) for point in STATE_POINTS],
        "tp_points": [add_tp_outputs(dict(point)) for point in TP_POINTS],
        "saturation_points": [add_sat_outputs(dict(point)) for point in SAT_POINTS],
    }
    OUT.write_text(json.dumps(dataset, indent=2) + "\n", encoding="utf-8")
    print(f"Wrote {OUT}")


if __name__ == "__main__":
    main()
