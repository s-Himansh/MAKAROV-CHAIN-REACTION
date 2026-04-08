# Monte Carlo Neutron Transport - Implementation Guide

> Step-by-step guide to building the core Monte Carlo simulation

## Table of Contents
1. [Overview](#overview)
2. [Mathematical Foundation](#mathematical-foundation)
3. [Implementation Breakdown](#implementation-breakdown)
4. [Step-by-Step Code](#step-by-step-code)
5. [Complete Example](#complete-example)
6. [Testing & Validation](#testing--validation)

---

## Overview

### What Are We Building?

A **generation-based Monte Carlo neutron transport code** that:
1. Tracks individual neutrons through material
2. Samples random interactions (fission/absorption/scatter/escape)
3. Runs multiple generations (Markov chain)
4. Calculates k_effective (criticality measure)

### The Big Picture

```
┌─────────────────────────────────────────────────────────────┐
│                    GENERATION N                             │
│                                                             │
│  Start with 1000 neutrons                                  │
│         │                                                   │
│         ├─→ Neutron 1: Track → Fission → Create 3 neutrons │
│         ├─→ Neutron 2: Track → Absorbed → 0 neutrons       │
│         ├─→ Neutron 3: Track → Escaped → 0 neutrons        │
│         ├─→ Neutron 4: Track → Fission → Create 2 neutrons │
│         └─→ ... (996 more)                                 │
│                                                             │
│  Result: 2400 new neutrons for Generation N+1              │
│  k = 2400/1000 = 2.4 (SUPERCRITICAL!)                     │
└─────────────────────────────────────────────────────────────┘
         │
         ├─→ GENERATION N+1 (start with 2400 neutrons)
         └─→ Repeat...
```

### Why "Monte Carlo"?

We use **random sampling** to simulate neutron behavior:
- Random distance to next collision
- Random interaction type (fission vs absorption vs scatter)
- Random number of neutrons from fission
- Random direction after scatter

**Law of Large Numbers:** Track enough neutrons → statistical average → accurate answer

---

## Mathematical Foundation

### 1. Exponential Free Path Sampling

**Question:** How far does a neutron travel before hitting an atom?

**Physics:** Neutrons travel through matter with probability of collision per unit distance = Σ_total

**Probability density:**
```
P(x) = Σ_total × e^(-Σ_total × x)
```

This is an **exponential distribution** with mean λ = 1/Σ_total (mean free path).

**Sampling method:** Inverse transform sampling

```
Given: Random number ξ ∈ [0, 1)
Find: Distance d such that P(travel < d) = ξ

Solution:
ξ = 1 - e^(-Σ_total × d)
e^(-Σ_total × d) = 1 - ξ
-Σ_total × d = ln(1 - ξ)
d = -ln(1 - ξ) / Σ_total

Simplification: Since (1-ξ) is uniformly distributed if ξ is,
d = -ln(ξ) / Σ_total
```

**Implementation:**
```go
func sampleCollisionDistance(Σ_total float64) float64 {
    ξ := rand.Float64()      // Random [0, 1)
    λ := 1.0 / Σ_total       // Mean free path
    return -λ * math.Log(ξ)  // Exponential sampling
}
```

**Example:**
```
Σ_total = 33.91 cm⁻¹
λ = 0.0295 cm
ξ = 0.3

d = -(0.0295) × ln(0.3)
d = -(0.0295) × (-1.204)
d = 0.0355 cm = 0.355 mm

→ Neutron travels 0.355 mm before collision
```

---

### 2. Discrete Event Sampling

**Question:** Which type of interaction occurs?

**Given:** Three possible outcomes with probabilities:
```
P(fission)    = Σ_fission / Σ_total
P(absorption) = Σ_absorption / Σ_total
P(scatter)    = Σ_scatter / Σ_total

Where: Σ_total = Σ_fission + Σ_absorption + Σ_scatter
```

**Probability line:**
```
0.0                                                    1.0
├──────────┼──────────┼──────────────────────────────┤
   FISSION   ABSORPTION        SCATTER

   P_f = 0.84
   P_a = 0.14
   P_s = 0.02
```

**Sampling method:**
```
Generate ξ ∈ [0, 1)

If ξ < P_f:
    → FISSION
Else if ξ < P_f + P_a:
    → ABSORPTION
Else:
    → SCATTER
```

**Implementation:**
```go
func sampleInteractionType(Σ_f, Σ_a, Σ_s, Σ_total float64) InteractionType {
    ξ := rand.Float64()

    P_f := Σ_f / Σ_total
    P_a := Σ_a / Σ_total

    if ξ < P_f {
        return FISSION
    } else if ξ < P_f + P_a {
        return ABSORPTION
    } else {
        return SCATTER
    }
}
```

**Example:**
```
Σ_fission = 28.59 cm⁻¹
Σ_absorption = 4.83 cm⁻¹
Σ_scatter = 0.49 cm⁻¹
Σ_total = 33.91 cm⁻¹

P_fission = 28.59/33.91 = 0.843 (84.3%)
P_absorption = 4.83/33.91 = 0.142 (14.2%)
P_scatter = 0.49/33.91 = 0.014 (1.4%)

Random ξ = 0.75
→ 0.75 < 0.843 → FISSION!
```

---

### 3. Neutron Multiplicity Sampling

**Question:** How many neutrons are produced in fission?

**Given:** Average nu (ν) = 2.44 neutrons/fission

**Distribution:** For U-235, the actual distribution is:
```
P(2 neutrons) ≈ 56%
P(3 neutrons) ≈ 44%
Average = 2.44
```

**Simple method:** Sample from discrete distribution
```
nu = 2.44
base = floor(nu) = 2
fraction = nu - base = 0.44

If random() < fraction:
    return base + 1  // 3 neutrons (44% chance)
Else:
    return base      // 2 neutrons (56% chance)
```

**Implementation:**
```go
func sampleNeutronEmission(nu float64) int {
    base := int(nu)                    // 2
    fraction := nu - float64(base)     // 0.44

    if rand.Float64() < fraction {
        return base + 1                // 3 neutrons
    }
    return base                        // 2 neutrons
}
```

**Advanced method:** Use actual probability distribution from ENDF
```go
// Actual U-235 fission neutron distribution
var nuDistribution = map[int]float64{
    0: 0.0,
    1: 0.0,
    2: 0.5604,  // 56.04%
    3: 0.4396,  // 43.96%
    4: 0.0,
}
```

For now, stick with the simple method!

---

### 4. Isotropic Scattering

**Question:** What direction does neutron go after scattering?

**Assumption:** Isotropic scattering = equal probability in all directions

**Math:** Sample point uniformly on unit sphere

**Method:**
```
θ = polar angle from z-axis
φ = azimuthal angle around z-axis

cos(θ) = 2ξ₁ - 1        (uniform in [-1, 1])
φ = 2π × ξ₂             (uniform in [0, 2π))

Direction vector:
dx = sin(θ) × cos(φ)
dy = sin(θ) × sin(φ)
dz = cos(θ)
```

**Implementation:**
```go
func sampleIsotropicDirection() (float64, float64, float64) {
    // Sample cos(θ) uniformly in [-1, 1]
    cosTheta := 2*rand.Float64() - 1
    sinTheta := math.Sqrt(1 - cosTheta*cosTheta)

    // Sample φ uniformly in [0, 2π)
    phi := 2 * math.Pi * rand.Float64()

    // Convert to Cartesian
    dx := sinTheta * math.Cos(phi)
    dy := sinTheta * math.Sin(phi)
    dz := cosTheta

    return dx, dy, dz
}
```

---

## Implementation Breakdown

### Architecture Overview

```
main.go
├─ Simulation struct
│  ├─ Material (from nuclear_data.json)
│  ├─ Geometry (sphere/slab)
│  ├─ SimParams (# particles, # generations)
│  └─ Tally (statistics)
│
├─ Physics Functions
│  ├─ sampleCollisionDistance()
│  ├─ sampleInteractionType()
│  ├─ sampleNeutronEmission()
│  └─ sampleIsotropicDirection()
│
├─ Tracking Function
│  └─ trackNeutron() → follows one neutron to completion
│
├─ Generation Loop
│  └─ Run() → tracks all neutrons, generation by generation
│
└─ Analysis
   ├─ calculateKeff()
   └─ PrintResults()
```

---

## Step-by-Step Code

### Step 1: Define Data Types

```go
package main

import (
    "fmt"
    "log"
    "math"
    "math/rand"
    "time"

    "makarov-chains/models"
)

// InteractionType represents what happened to a neutron
type InteractionType int

const (
    Fission InteractionType = iota
    Absorption
    Scatter
    Escape
)

// String method for nice printing
func (it InteractionType) String() string {
    switch it {
    case Fission:
        return "FISSION"
    case Absorption:
        return "ABSORPTION"
    case Scatter:
        return "SCATTER"
    case Escape:
        return "ESCAPE"
    default:
        return "UNKNOWN"
    }
}

// NeutronOutcome holds the result of tracking one neutron
type NeutronOutcome struct {
    Type        InteractionType
    NewNeutrons int  // Only non-zero for fission
}
```

**Explanation:**
- `InteractionType` is an enum (0, 1, 2, 3)
- `iota` auto-increments: Fission=0, Absorption=1, etc.
- `String()` method lets us print it nicely
- `NeutronOutcome` packages the result of tracking one neutron

---

### Step 2: Create Simulation Struct

```go
// Simulation holds all data and methods for Monte Carlo
type Simulation struct {
    // Configuration
    Material models.Material
    Geometry models.Geometry
    Params   models.SimParams

    // Results
    Tally models.Tally

    // Pre-calculated cross sections (for speed)
    ΣFission    float64  // cm⁻¹
    ΣAbsorption float64  // cm⁻¹
    ΣScatter    float64  // cm⁻¹
    ΣTotal      float64  // cm⁻¹
}
```

**Explanation:**
- Groups all simulation data in one place
- Pre-calculates macroscopic cross sections (don't recalculate every time)
- Stores material, geometry, parameters, results

---

### Step 3: Initialize Simulation

```go
// NewSimulation creates and initializes a simulation
func NewSimulation(library *models.NuclearDataLibrary, isotope string, thickness float64) (*Simulation, error) {
    // Load material from nuclear data
    material, err := library.CreateMetalMaterial(isotope)
    if err != nil {
        return nil, err
    }

    sim := &Simulation{
        Material: material,
        Geometry: models.Geometry{
            Volume: thickness,  // For 1D slab, "volume" = thickness
        },
        Params: models.SimParams{
            NumParticles:   1000,
            NumGenerations: 50,
            Seed:           time.Now().UnixNano(),
        },
        Tally: models.Tally{
            NeutronsPerGen: make([]int, 0),
        },
    }

    // Pre-calculate macroscopic cross sections
    sim.ΣFission = material.MacroscopicCrossSection(material.Nuclide.FissionCrossSection)
    sim.ΣAbsorption = material.MacroscopicCrossSection(material.Nuclide.AbsorptionCrossSection)
    sim.ΣScatter = material.MacroscopicCrossSection(material.Nuclide.ScatterCrossSection)
    sim.ΣTotal = material.TotalCrossSections()

    return sim, nil
}
```

**Explanation:**
- Creates simulation from isotope name
- Sets default parameters (1000 neutrons, 50 generations)
- Pre-calculates Σ values for efficiency
- Returns pointer so we can call methods on it

**Example usage:**
```go
library, _ := models.LoadNuclearData("nuclear_data.json")
sim, _ := NewSimulation(library, "U-235", 20.0)  // 20 cm slab
```

---

### Step 4: Implement Physics Sampling Functions

#### 4a. Sample Collision Distance

```go
// sampleCollisionDistance samples the distance to next collision
// using exponential distribution: P(d) = Σ × e^(-Σd)
func (s *Simulation) sampleCollisionDistance() float64 {
    ξ := rand.Float64()        // Random [0, 1)
    λ := 1.0 / s.ΣTotal        // Mean free path [cm]
    return -λ * math.Log(ξ)    // Exponential sampling
}
```

**Why it works:**
```
P(distance < d) = ξ
∫₀ᵈ Σe^(-Σx) dx = ξ
1 - e^(-Σd) = ξ
e^(-Σd) = 1 - ξ
d = -ln(1-ξ)/Σ = -ln(ξ)/Σ  (since 1-ξ ~ uniform)
```

---

#### 4b. Sample Interaction Type

```go
// sampleInteractionType determines which reaction occurs
// Based on relative cross section magnitudes
func (s *Simulation) sampleInteractionType() InteractionType {
    ξ := rand.Float64()  // Random [0, 1)

    // Calculate probabilities
    P_fission := s.ΣFission / s.ΣTotal
    P_absorption := s.ΣAbsorption / s.ΣTotal
    // P_scatter = s.ΣScatter / s.ΣTotal (implicit)

    // Sample from discrete distribution
    if ξ < P_fission {
        return Fission
    } else if ξ < P_fission + P_absorption {
        return Absorption
    } else {
        return Scatter
    }
}
```

**Visual:**
```
0.0                                                    1.0
├──────────────────┼────┼──┤
    FISSION         ABS  SC
    84.3%          14.2% 1.4%
```

---

#### 4c. Sample Neutron Emission

```go
// sampleNeutronEmission samples number of neutrons from fission
// Uses simple interpolation between floor and ceiling of nu
func (s *Simulation) sampleNeutronEmission() int {
    nu := s.Material.Nuclide.Nu  // e.g., 2.44

    base := int(nu)                   // 2
    fraction := nu - float64(base)    // 0.44

    // 44% chance of 3, 56% chance of 2
    if rand.Float64() < fraction {
        return base + 1
    }
    return base
}
```

**Distribution:**
```
nu = 2.44

P(2) = 56%  ─────┐
P(3) = 44%  ──┐  │
              │  │
             ┌┴──┴┐
             2    3  neutrons
```

---

#### 4d. Sample Isotropic Direction (For Scattering)

```go
// sampleIsotropicDirection samples a uniformly random direction
// Returns unit vector (dx, dy, dz)
func (s *Simulation) sampleIsotropicDirection() (float64, float64, float64) {
    // Sample cos(θ) uniformly in [-1, 1]
    cosTheta := 2*rand.Float64() - 1
    sinTheta := math.Sqrt(1 - cosTheta*cosTheta)

    // Sample φ uniformly in [0, 2π)
    phi := 2 * math.Pi * rand.Float64()

    // Convert to Cartesian coordinates
    dx := sinTheta * math.Cos(phi)
    dy := sinTheta * math.Sin(phi)
    dz := cosTheta

    return dx, dy, dz
}
```

**Why this works:**
- cos(θ) uniform in [-1,1] → uniform on sphere surface
- φ uniform in [0, 2π) → covers all azimuths
- Result: equal probability in all directions

---

### Step 5: Track Single Neutron

This is the **core function** that follows one neutron from birth to death.

```go
// trackNeutron follows one neutron through material until it's absorbed, fissions, or escapes
func (s *Simulation) trackNeutron() NeutronOutcome {
    // Start at center of slab
    position := s.Geometry.Volume / 2.0
    slabThickness := s.Geometry.Volume

    // Neutron tracking loop
    for {
        // 1. Sample distance to next collision
        distance := s.sampleCollisionDistance()

        // 2. Move neutron
        // In 1D, randomly choose direction (left or right)
        if rand.Float64() < 0.5 {
            position += distance
        } else {
            position -= distance
        }

        // 3. Check if neutron escaped geometry
        if position < 0 || position > slabThickness {
            return NeutronOutcome{Type: Escape}
        }

        // 4. Neutron collided - sample interaction type
        interactionType := s.sampleInteractionType()

        // 5. Process interaction
        switch interactionType {
        case Fission:
            // Neutron caused fission
            newNeutrons := s.sampleNeutronEmission()
            return NeutronOutcome{
                Type:        Fission,
                NewNeutrons: newNeutrons,
            }

        case Absorption:
            // Neutron absorbed, stops here
            return NeutronOutcome{Type: Absorption}

        case Scatter:
            // Neutron scattered, continues flying
            // In 1D: just continues in random direction
            // In 3D: would call sampleIsotropicDirection()
            continue
        }
    }
}
```

**What this does:**

1. **Start** at center of geometry
2. **Loop** until neutron dies or escapes:
   - Sample random distance to collision
   - Move neutron
   - Check if escaped
   - If not, sample what interaction happened
   - Process interaction:
     - **Fission** → return with new neutron count
     - **Absorption** → return (neutron dies)
     - **Scatter** → continue loop (neutron keeps going)

**Example trace:**
```
Neutron born at position 10 cm

Step 1: Sample distance = 0.3 cm, move left → pos = 9.7 cm
        Sample interaction → SCATTER
        Continue...

Step 2: Sample distance = 0.4 cm, move right → pos = 10.1 cm
        Sample interaction → SCATTER
        Continue...

Step 3: Sample distance = 0.2 cm, move right → pos = 10.3 cm
        Sample interaction → FISSION
        Produce 3 neutrons
        DONE!
```

---

### Step 6: Run Generation Loop

This runs multiple generations, building the Markov chain.

```go
// Run executes the Monte Carlo simulation
func (s *Simulation) Run() {
    // Initialize random seed
    rand.Seed(s.Params.Seed)

    // Start with initial neutron population
    neutronBank := s.Params.NumParticles

    fmt.Printf("Starting simulation: %s, %.1f cm thick\n",
        s.Material.Nuclide.Name, s.Geometry.Volume)
    fmt.Printf("Σ_total = %.2f cm⁻¹, MFP = %.4f cm\n\n",
        s.ΣTotal, 1.0/s.ΣTotal)

    // Generation loop
    for gen := 0; gen < s.Params.NumGenerations; gen++ {
        // Check if chain reaction died
        if neutronBank == 0 {
            fmt.Println("Chain reaction died (subcritical)")
            break
        }

        // Counters for this generation
        nextBank := 0
        fissions := 0
        absorptions := 0
        escapes := 0

        // Track each neutron in current generation
        for n := 0; n < neutronBank; n++ {
            outcome := s.trackNeutron()

            switch outcome.Type {
            case Fission:
                fissions++
                nextBank += outcome.NewNeutrons
                s.Tally.TotalFissions++

            case Absorption:
                absorptions++
                s.Tally.TotalAbsorptions++

            case Escape:
                escapes++
                s.Tally.TotalEscapes++
            }
        }

        // Record statistics
        s.Tally.NeutronsPerGen = append(s.Tally.NeutronsPerGen, nextBank)

        // Calculate k for this generation
        k_gen := float64(nextBank) / float64(neutronBank)

        // Print generation summary
        fmt.Printf("Gen %2d: %5d → %5d neutrons (F:%4d A:%4d E:%4d) k=%.3f\n",
            gen+1, neutronBank, nextBank, fissions, absorptions, escapes, k_gen)

        // Update neutron bank for next generation
        neutronBank = nextBank
    }

    // Calculate final k_eff
    s.calculateKeff()
}
```

**What this does:**

```
Generation 0: Start with 1000 neutrons
   │
   ├─ Track neutron 1 → Fission → 3 new neutrons
   ├─ Track neutron 2 → Absorbed → 0 neutrons
   ├─ Track neutron 3 → Escaped → 0 neutrons
   └─ ... (997 more)

   Result: 2400 neutrons produced
   k = 2400 / 1000 = 2.4

Generation 1: Start with 2400 neutrons
   └─ Repeat...
```

---

### Step 7: Calculate k_effective

```go
// calculateKeff computes k_eff from generation history
func (s *Simulation) calculateKeff() {
    if len(s.Tally.NeutronsPerGen) < 2 {
        s.Tally.KEff = 0.0
        return
    }

    // Skip first few generations (burn-in period)
    // These are biased by initial conditions
    skipGens := 10
    if len(s.Tally.NeutronsPerGen) < 20 {
        skipGens = len(s.Tally.NeutronsPerGen) / 2
    }

    sum := 0.0
    count := 0

    // Average k over all generations (after burn-in)
    for i := skipGens; i < len(s.Tally.NeutronsPerGen); i++ {
        // k_i = neutrons_out[i] / neutrons_in[i]
        neutronsIn := s.Params.NumParticles
        if i > 0 {
            neutronsIn = s.Tally.NeutronsPerGen[i-1]
        }

        if neutronsIn > 0 {
            k := float64(s.Tally.NeutronsPerGen[i]) / float64(neutronsIn)
            sum += k
            count++
        }
    }

    if count > 0 {
        s.Tally.KEff = sum / float64(count)
    }
}
```

**Why skip generations?**

```
Generation:  1    2    3    4    5    6    7    8    9   10  ...
k:          3.2  2.8  2.6  2.5  2.45 2.44 2.43 2.44 2.43 2.44
            ↑─────── BURN-IN ──────↑   ↑───── STABLE ─────↑

Initial transient (sensitive to starting conditions)
Skip these when averaging!
```

---

### Step 8: Print Results

```go
// PrintResults displays simulation results
func (s *Simulation) PrintResults() {
    fmt.Println("\n" + "=".repeat(50))
    fmt.Println("SIMULATION RESULTS")
    fmt.Println("=".repeat(50))
    fmt.Printf("Material: %s\n", s.Material.Nuclide.Name)
    fmt.Printf("Geometry: 1D slab, %.2f cm thick\n", s.Geometry.Volume)
    fmt.Printf("\nTotal events:\n")
    fmt.Printf("  Fissions:    %d\n", s.Tally.TotalFissions)
    fmt.Printf("  Absorptions: %d\n", s.Tally.TotalAbsorptions)
    fmt.Printf("  Escapes:     %d\n", s.Tally.TotalEscapes)
    fmt.Printf("\nk_effective: %.4f\n", s.Tally.KEff)

    // Interpret k_eff
    if s.Tally.KEff > 1.0 {
        fmt.Println("→ SUPERCRITICAL (growing reaction)")
    } else if s.Tally.KEff < 1.0 {
        fmt.Println("→ SUBCRITICAL (dying reaction)")
    } else {
        fmt.Println("→ CRITICAL (sustained reaction)")
    }
}
```

---

## Complete Example

Here's how to put it all together in `main.go`:

```go
package main

import (
    "fmt"
    "log"
    "math"
    "math/rand"
    "time"

    "makarov-chains/models"
)

// [Include all the code from steps 1-8 above]

func main() {
    // Load nuclear data
    library, err := models.LoadNuclearData("nuclear_data.json")
    if err != nil {
        log.Fatal("Failed to load nuclear data:", err)
    }

    fmt.Println("Loaded:", library.GetMetadata())
    fmt.Println()

    // Create simulation
    // 20 cm thick U-235 slab
    sim, err := NewSimulation(library, "U-235", 20.0)
    if err != nil {
        log.Fatal("Failed to create simulation:", err)
    }

    // Run simulation
    sim.Run()

    // Print results
    sim.PrintResults()
}
```

---

## Testing & Validation

### Test 1: Subcritical System

**Thin slab → lots of escapes → subcritical**

```go
sim, _ := NewSimulation(library, "U-235", 1.0)  // 1 cm thin
sim.Run()
// Expected: k_eff < 1.0
```

**Why?** Neutrons escape before causing many fissions.

---

### Test 2: Supercritical System

**Thick slab → few escapes → supercritical**

```go
sim, _ := NewSimulation(library, "U-235", 50.0)  // 50 cm thick
sim.Run()
// Expected: k_eff > 1.0
```

**Why?** Neutrons can't escape, high fission probability.

---

### Test 3: Non-Fissile Material

**U-238 → no fission → k_eff ≈ 0**

```go
sim, _ := NewSimulation(library, "U-238", 20.0)
sim.Run()
// Expected: k_eff ≈ 0.0 (chain dies immediately)
```

**Why?** U-238 doesn't fission at thermal energies.

---

### Test 4: Statistical Convergence

Run same simulation 10 times:

```go
for i := 0; i < 10; i++ {
    sim, _ := NewSimulation(library, "U-235", 20.0)
    sim.Params.Seed = int64(i)  // Different random seed
    sim.Run()
    fmt.Printf("Run %d: k_eff = %.4f\n", i+1, sim.Tally.KEff)
}
```

Expected:
```
Run 1: k_eff = 2.4521
Run 2: k_eff = 2.4489
Run 3: k_eff = 2.4556
...
Average ≈ 2.45 ± 0.01
```

**Standard deviation should decrease with more neutrons!**

---

## Common Issues & Solutions

### Issue 1: k_eff Keeps Growing

```
Gen 1: k = 2.5
Gen 2: k = 6.2
Gen 3: k = 15.8
...
```

**Problem:** Supercritical system with no population control

**Solution:** Add population control (weight adjustment) or stop at max neutrons:
```go
if neutronBank > 100000 {
    fmt.Println("Population exploded! System highly supercritical.")
    break
}
```

---

### Issue 2: Chain Dies Immediately

```
Gen 1: k = 0.0
Chain reaction died
```

**Problem:** Non-fissile material or very thin geometry

**Check:**
- Is σ_fission > 0? (check material data)
- Is geometry thick enough?
- Are neutrons all escaping?

---

### Issue 3: k_eff is Noisy

```
Gen 10: k = 2.5
Gen 11: k = 1.8
Gen 12: k = 3.2
```

**Problem:** Not enough neutrons (high statistical variance)

**Solution:** Increase `NumParticles`:
```go
sim.Params.NumParticles = 10000  // More neutrons = less noise
```

---

## Summary

**What you've built:**

1. ✅ **Physics sampling** - Exponential, discrete, etc.
2. ✅ **Neutron tracking** - Follow one neutron to completion
3. ✅ **Generation loop** - Build Markov chain
4. ✅ **k_eff calculation** - Measure criticality

**Flow:**
```
Load Data → Create Simulation → Run Generations → Calculate k_eff
```

**Next steps:**
- Extend to 3D geometry (spheres)
- Add energy groups (thermal + fast)
- Implement variance reduction techniques
- Add more physics (scattering angles, etc.)

🎉 **You now have a working Monte Carlo neutron transport code!**
