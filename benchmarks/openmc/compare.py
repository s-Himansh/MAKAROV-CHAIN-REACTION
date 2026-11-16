#!/usr/bin/env python3
"""Compare this repo's k_eff with OpenMC in multi-group mode.

OpenMC can run with a user-supplied one-group library instead of its
continuous-energy data, so we feed it exactly the macroscopic cross sections
the Go engine uses (nuclear_data.json) and compare bare spheres of the same
radii. No external nuclear data download is needed.

    python3 benchmarks/openmc/compare.py --isotope U-235 --radii 0.02 0.05 0.1 0.3

Run from the repository root with `openmc` importable and `go` on PATH.
Tested with OpenMC 0.16.0 (conda-forge: `micromamba create -n openmc -c conda-forge openmc`).
"""
import argparse
import json
import subprocess
import tempfile
from pathlib import Path

AVOGADRO = 6.022e23  # same constant as models/constants.go
ROOT = Path(__file__).resolve().parents[2]


def macro_xs(isotope):
    """Return (sigma_t, sigma_a, sigma_f, nu_sigma_f, sigma_s) in 1/cm."""
    data = json.loads((ROOT / "nuclear_data.json").read_text())["isotopes"][isotope]
    n = data["density_metal"] * AVOGADRO / data["molar_mass"] * 1e-24  # atoms/(b*cm)
    fis, cap, sca = (n * data[k] for k in ("xs_fission", "xs_capture", "xs_elastic"))
    # The Go engine's total is fission + capture + elastic. OpenMC's
    # "absorption" includes fission.
    return fis + cap + sca, fis + cap, fis, data["nu"] * fis, sca


def go_keff(isotope, radius, particles, generations, inactive, seed):
    with tempfile.TemporaryDirectory() as tmp:
        out = Path(tmp) / "r.json"
        subprocess.run(
            ["go", "run", ".", "--quiet", "--isotope", isotope, "--geometry", "sphere",
             "--thickness", str(radius), "--particles", str(particles),
             "--generations", str(generations), "--inactive", str(inactive), "--seed", str(seed), "--output", str(out)],
            cwd=ROOT, check=True, capture_output=True)
        return json.loads(out.read_text())["k_effective"]


def openmc_keff(isotope, radius, particles, generations, inactive):
    import numpy as np
    import openmc

    st, sa, sf, nusf, ss = macro_xs(isotope)
    with tempfile.TemporaryDirectory() as tmp:
        tmp = Path(tmp)
        groups = openmc.mgxs.EnergyGroups([0.0, 20.0e6])
        xs = openmc.XSdata("fuel", groups)
        xs.order = 0
        xs.set_total([st])
        xs.set_absorption([sa])
        xs.set_fission([sf])
        xs.set_nu_fission([nusf])
        xs.set_chi([1.0])
        xs.set_scatter_matrix(np.array([[[ss]]]))
        lib = openmc.MGXSLibrary(groups)
        lib.add_xsdata(xs)
        lib.export_to_hdf5(str(tmp / "mgxs.h5"))

        mat = openmc.Material(name="fuel")
        mat.set_density("macro", 1.0)
        mat.add_macroscopic(openmc.Macroscopic("fuel"))
        materials = openmc.Materials([mat])
        materials.cross_sections = str(tmp / "mgxs.h5")

        sphere = openmc.Sphere(r=radius, boundary_type="vacuum")
        geometry = openmc.Geometry([openmc.Cell(fill=mat, region=-sphere)])

        settings = openmc.Settings()
        settings.energy_mode = "multi-group"
        settings.batches = generations
        settings.inactive = inactive
        settings.particles = particles
        settings.source = openmc.IndependentSource(space=openmc.stats.Point((0, 0, 0)))
        settings.output = {"tallies": False, "summary": False}

        model = openmc.Model(geometry, materials, settings)
        sp_path = model.run(cwd=str(tmp), output=False)
        with openmc.StatePoint(sp_path) as sp:
            return sp.keff.n, sp.keff.s


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--isotope", default="U-235")
    ap.add_argument("--radii", type=float, nargs="+", default=[0.02, 0.05, 0.1, 0.3])
    ap.add_argument("--particles", type=int, default=20000)
    ap.add_argument("--generations", type=int, default=400)
    ap.add_argument("--inactive", type=int, default=100,
                    help="burn-in generations; thick systems converge slowly, too few biases k")
    ap.add_argument("--seed", type=int, default=1)
    ap.add_argument("--skip-openmc", action="store_true", help="only print the Go results")
    a = ap.parse_args()

    st, sa, sf, nusf, _ = macro_xs(a.isotope)
    print(f"{a.isotope}: Sigma_t={st:.3f}/cm  analytic k_inf = nuSigma_f/Sigma_a = {nusf / sa:.4f}")
    print(f"{'radius cm':>10} {'go k_eff':>10} {'openmc k_eff':>16} {'diff':>8}")
    for r in a.radii:
        g = go_keff(a.isotope, r, a.particles, a.generations, a.inactive, a.seed)
        if a.skip_openmc:
            print(f"{r:>10} {g:>10.4f}")
            continue
        k, s = openmc_keff(a.isotope, r, a.particles, a.generations, a.inactive)
        print(f"{r:>10} {g:>10.4f} {k:>9.4f}±{s:.4f} {g - k:>+8.4f}")


if __name__ == "__main__":
    main()
