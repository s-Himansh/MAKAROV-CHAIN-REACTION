# OpenMC cross-check

`compare.py` runs the same bare-sphere problem in this repo and in OpenMC's
multi-group mode, using one energy group built from `nuclear_data.json`, and
prints both k_eff values side by side.

```bash
python3 benchmarks/openmc/compare.py --radii 0.02 0.05 0.1 0.3   # needs openmc + go
python3 benchmarks/openmc/compare.py --skip-openmc                # Go side only
```

Multi-group mode needs no continuous-energy data library, so the comparison
isolates transport and fission-source handling from cross-section differences.

What to expect, and what a mismatch means:

- Large radii should approach the analytic k_inf = nu*Sigma_f / Sigma_a (printed
  at the top), about 2.08 for U-235.
- Small radii depend on leakage. A systematic gap there points at the geometry,
  source sampling or scattering treatment.
- The Go engine is isotropic and one-group, as is the OpenMC input, so
  differences beyond a few standard errors are worth investigating.

The OpenMC half has not been run yet (no OpenMC install was available when it
was written); treat the first run as a check of the script as well.
