# Validation Method

The complete implementation and qualification requirements are maintained in [validation-plan.md](validation-plan.md).

The validator is a separate executable. It compares the Go candidate with a pinned CoolProp Python worker over deterministic screening, saturation, boundary, flash, invalid-input, and adaptive cases. Every case receives a stable identifier and is recorded as passed, failed, a consistent dual error, an error mismatch, unsupported, or an execution failure.

The `D` property is mass density in kg/m³; `Dmolar` is molar density in mol/m³. Results record versions, reference state, configuration, seed, worker counts, and platform metadata. A normal run requires the reference worker to start successfully, then writes immutable `manifest.json` and `config.resolved.yaml` files before scheduling fluids.
