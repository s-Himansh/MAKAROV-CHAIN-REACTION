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

OpenMC 0.16.0, U-235 bare sphere, one group built from `nuclear_data.json`,
20,000 neutrons per generation, 400 generations with the first 100 discarded
(the script defaults). At r = 1 cm in a larger run OpenMC returns 2.08425,
equal to the analytic k_inf = 2.0843, so the cross-section mapping is exact.

| radius (cm) | Go k_eff | OpenMC k_eff | diff |
|---|---|---|---|
| 0.02 | 0.8019 | 0.8020 +/- 0.0002 | -0.0001 |
| 0.05 | 1.3841 | 1.3851 +/- 0.0003 | -0.0010 |
| 0.1 | 1.7613 | 1.7622 +/- 0.0002 | -0.0009 |
| 0.3 | 2.0281 | 2.0282 +/- 0.0001 | -0.0001 |
| 1.0 | 2.0772 | 2.0811 +/- 0.0001 | -0.0038 |

The engines agree to 0.01-0.07% from 0.02 to 0.3 cm. At 1 cm (about 30 mean
free paths across) the fission source converges slowly, so 100 inactive
generations is not enough for either code; Go with 300 inactive of 800 gives
2.0785, still approaching k_inf.

## What the comparison found and fixed

An earlier run showed Go 0.5-1% below OpenMC at 0.3 cm. Two causes, both fixed:

1. **Burn-in too short.** The fission source of a sphere several mean free
   paths across takes tens of generations to converge. Go started from a
   uniform source (k too low) and OpenMC from a point at the centre (k too
   high), and with 10 inactive generations neither had settled. First-generation
   transport matched (leakage 7.41% in Go versus 7.28% +/- 0.08 in OpenMC for a
   fixed uniform source), which ruled out the transport itself. `--inactive`
   makes the burn-in configurable.
2. **Subcritical systems died out.** The bank was only ever reduced, so a
   subcritical run shrank every generation until no neutrons were left and
   k_eff was never estimated. The bank is now renormalised to `--particles`
   every generation, as OpenMC does.
