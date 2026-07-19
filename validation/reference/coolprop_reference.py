#!/usr/bin/env python3
import json, os, platform, sys

try:
    import CoolProp
    import CoolProp.CoolProp as CP
except Exception as exc:
    print(json.dumps({"startup_error": str(exc)}), flush=True)
    raise

def evaluate(case):
    try:
        value = CP.PropsSI(case["output"], case["input1"], case["value1"], case["input2"], case["value2"], case["fluid"])
        phase = CP.PhaseSI(case["input1"], case["value1"], case["input2"], case["value2"], case["fluid"])
        return {"value": value, "error": None, "phase": phase}
    except Exception as exc:
        return {"value": 0.0, "error": str(exc), "phase": ""}

print(json.dumps({"ready": True, "coolprop_version": getattr(CoolProp, "__version__", "unknown"), "python_version": platform.python_version(), "backend": "HEOS", "reference_state": "DEF"}), flush=True)
for line in sys.stdin:
    try:
        request = json.loads(line)
        print(json.dumps({"request_id": request["request_id"], "results": [evaluate(c) for c in request["cases"]]}), flush=True)
    except Exception as exc:
        print(json.dumps({"request_id": request.get("request_id", "") if isinstance(request, dict) else "", "error": str(exc), "results": []}), flush=True)
