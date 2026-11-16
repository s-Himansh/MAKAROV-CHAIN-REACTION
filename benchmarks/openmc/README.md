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

## Results

OpenMC 0.16.0, U-235 bare sphere, one group built from `nuclear_data.json`.
OpenMC's own k_inf check: at r = 1 cm it returns 2.08425, equal to the analytic
2.0843, so the cross-section mapping is exact.

| radius (cm) | Go k_eff | OpenMC k_eff |
|---|---|---|
| 0.02 | 0.75-0.88 (6 seeds, noisy: near-subcritical) | 0.803 +/- 0.001 |
| 0.05 | 1.382 | 1.385 +/- 0.001 |
| 0.1 | 1.768 | 1.765 +/- 0.002 |
| 0.3 | 2.0255 +/- 0.001 (6 seeds) | 2.037-2.044 (two runs) |
| 1.0 | about k_inf | 2.0843 |

The engines agree to about 0.2% at 0.05 and 0.1 cm. At 0.3 cm Go is roughly
0.5-1% lower, more than its seed-to-seed spread (0.001), so there is a small
systematic difference in leakage that is not yet explained. Two OpenMC runs
at 0.3 cm differ by about 5 of their own reported standard deviations, so
OpenMC's quoted errors are optimistic here as well. Candidates to check next:
the bank and downsampling scheme versus OpenMC's per-collision fission-site
banking, and the k estimator (generation ratio versus OpenMC's combined
estimator).
