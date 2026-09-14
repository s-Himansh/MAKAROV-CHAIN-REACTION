# MAKAROV-CHAIN-REACTION

A Monte Carlo neutron transport simulator that models nuclear chain reactions using Markov chains. Track individual neutrons through fissile material, sampling random interactions (fission, absorption, scattering, escape) to compute the effective multiplication factor (k_effective).

## Quick Start

```bash
# Default: U-235 slab, 20cm, 1000 neutrons, 50 generations
go run main.go

# Custom simulation
go run main.go --isotope Pu-239 --thickness 10 --geometry sphere --particles 5000 --generations 100

# Export results to JSON
go run main.go --isotope U-235 --thickness 15 --output results.json

# Quiet mode (summary only)
go run main.go --quiet --output results.json
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--isotope` | `U-235` | Isotope to simulate |
| `--thickness` | `20.0` | Slab thickness or sphere radius (cm) |
| `--particles` | `1000` | Neutrons per generation |
| `--generations` | `50` | Number of generations |
| `--geometry` | `slab` | `slab` or `sphere` |
| `--seed` | `0` (auto) | Random seed for reproducibility |
| `--output` | (none) | Export to `.json` or `.csv` |
| `--quiet` | `false` | Suppress per-generation output |

## Available Isotopes

| Isotope | Role | Fission XS (b) | ν |
|---------|------|----------------|---|
| U-235 | Primary fissile fuel | 584.4 | 2.44 |
| U-238 | Fertile material | 0.0 | 2.50 |
| Pu-239 | Weapons-grade fissile | 747.4 | 2.88 |
| Pu-241 | Highly fissile plutonium | 1011.0 | 2.92 |
| H-1 | Water moderator | 0.0 | — |
| C-12 | Graphite moderator | 0.0 | — |
| B-10 | Control rod absorber | 0.0 | — |
| Cd-113 | Control rod absorber | 0.0 | — |

## Physics

The simulation uses Monte Carlo methods to solve the neutron transport equation:

1. **Free path sampling**: Distance to next collision sampled from exponential distribution `P(d) = Σ × e^(-Σd)`
2. **Interaction sampling**: Discrete sampling based on relative macroscopic cross sections (fission vs absorption vs scatter)
3. **Neutron multiplicity**: Number of neutrons per fission sampled from fractional ν
4. **Population control**: Supercritical banks capped to prevent explosion (standard practice in criticality calculations)
5. **k_effective**: Computed as average generation-wise neutron ratio after burn-in period

## Architecture

```
main.go                          CLI entry point
models/
  constants.go                   Avogadro's number
  nuclide.go                     Microscopic cross sections
  nuclear_data.go                JSON deserialization
  material.go                    Macroscopic cross section math
  geometry.go                    Slab/Sphere geometry
  interaction_type.go            Fission/Absorption/Scatter/Escape
  tally.go                       Aggregated statistics
  sim_params.go                  Simulation parameters
loading_nuclear_data/
  nuclear_data.go                Nuclear data library loader
simulation/
  simulation.go                  Core Monte Carlo engine
nuclear_data.json                ENDF/B-VIII.0 cross-section data
```

## Output

```
Starting simulation: U-235, 1D slab, 20.0 cm thick
Σ_total = 33.92 cm⁻¹, MFP = 0.0295 cm

Gen  1:  1000 →  2092 neutrons (F: 861 A: 139 E:   0) k=2.092
Gen  2:  1000 →  2060 neutrons (F: 852 A: 148 E:   0) k=2.060
...

==================================================
SIMULATION RESULTS
==================================================
Material: U-235
Geometry: 1D slab, 20.00 cm thick

Total events:
  Fissions:    42630
  Absorptions: 7370
  Escapes:     0

k_effective: 2.0726
→ SUPERCRITICAL (growing reaction)
```

## JSON Export

```json
{
  "material": "U-235",
  "geometry": "slab",
  "thickness": 20,
  "num_particles": 1000,
  "num_generations": 50,
  "seed": 1789397526413582000,
  "total_fissions": 42630,
  "total_absorptions": 7370,
  "total_escapes": 0,
  "k_effective": 2.0726,
  "status": "supercritical",
  "neutrons_per_gen": [2092, 2060, ...]
}
```

## Testing

```bash
go test ./...
```

## Cross-Section Data

Thermal neutron cross sections at 0.0253 eV (293.6 K) from ENDF/B-VIII.0.
See `DATA_SOURCES.md` for full attribution and verification instructions.
