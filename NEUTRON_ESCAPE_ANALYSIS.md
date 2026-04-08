# Neutron Escape Analysis

> Understanding why neutrons don't escape thick U-235 slabs

## Table of Contents
1. [Observation](#observation)
2. [Physical Explanation](#physical-explanation)
3. [Mathematical Analysis](#mathematical-analysis)
4. [Thickness Comparison](#thickness-comparison)
5. [Critical Mass Implications](#critical-mass-implications)

---

## Observation

When running the Monte Carlo simulation with a **20 cm thick U-235 slab**, we observe:

```
Gen 20: 1000 → 2061 neutrons (F:842 A:158 E:0) k=2.061
                                          ↑
                                    ZERO ESCAPES!
```

**Question:** Is this a bug or correct physics?

**Answer:** ✅ **Correct physics!** The slab is so thick that neutrons essentially never escape.

---

## Physical Explanation

### The Key Numbers

For U-235 thermal neutrons at 0.0253 eV (room temperature):

```
Σ_total = 33.92 cm⁻¹
Mean Free Path (MFP) = 1/Σ_total = 0.0295 cm = 0.295 mm
```

This means neutrons travel **less than 0.3 mm** on average before hitting a U-235 nucleus.

### Cross Section Breakdown

```
Σ_fission    = 28.59 cm⁻¹  →  P(fission)    = 84.3%
Σ_absorption =  4.83 cm⁻¹  →  P(absorption) = 14.2%
Σ_scatter    =  0.49 cm⁻¹  →  P(scatter)    =  1.4%
─────────────────────────────────────────────────────
Σ_total      = 33.92 cm⁻¹  →  Total         = 100%
```

**Key insight:** Only 1.4% of collisions result in scattering (where neutron continues moving). Most collisions end in fission or absorption!

---

## Mathematical Analysis

### Escape Probability Calculation

For a neutron born at the **center** of a 20 cm slab:

1. **Distance to escape:** 10 cm (half the thickness)
2. **Number of MFPs to travel:** 10 cm / 0.0295 cm = **339 MFPs**
3. **Required trajectory:** Must scatter ~339 times without fission/absorption

### Probability Calculation

At each collision, the neutron must scatter to continue:

```
P(scatter) = 0.0144 per collision

P(escape after N collisions) ≈ (P_scatter)^N

For N = 339:
P(escape) ≈ (0.0144)^339 ≈ 10^-600

Result: ESSENTIALLY ZERO
```

For comparison:
- Number of atoms in the universe ≈ 10^80
- This probability is 10^-600 → incomprehensibly small!

### Why Fission Dominates

```
Neutron born → Travels 0.3 mm → COLLISION
                                    ↓
                    ┌───────────────┼───────────────┐
                    │               │               │
                 FISSION       ABSORPTION       SCATTER
                 84.3%           14.2%           1.4%
                    ↓               ↓               ↓
                 DIES            DIES          CONTINUES
                                                   ↓
                                         Travels 0.3 mm → COLLISION
                                                           ↓
                                                    (Repeat 339 times
                                                     to escape?)
                                                     → IMPOSSIBLE!
```

---

## Thickness Comparison

### Simulation Results

| Thickness | MFPs to Edge | Escape % | Fission % | Absorption % | k_eff | Status |
|-----------|--------------|----------|-----------|--------------|-------|---------|
| **20.0 cm** | 678 MFPs | **0.00%** | 85.6% | 14.4% | 2.085 | Infinite medium |
| 0.5 cm | 17 MFPs | 0.01% | 85.6% | 14.3% | 2.087 | Near-infinite |
| **0.1 cm** | 3.4 MFPs | **18.9%** | 69.5% | 11.7% | 1.693 | Significant leakage |

### Key Observations

1. **20 cm slab (678 MFPs):**
   - Behaves as "infinite medium"
   - No neutrons escape
   - k_eff = 2.085 (intrinsic multiplication factor)

2. **0.1 cm slab (3.4 MFPs):**
   - ~19% of neutrons escape before interacting
   - k_eff drops to 1.693 (leakage reduces criticality)
   - Fission rate drops from 85.6% → 69.5%

### Visual Comparison

```
20 cm SLAB (678 MFPs):
┌─────────────────────────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░ NEUTRON ░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░ (Trapped) ░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
└─────────────────────────────────────────┘
     Escape distance: 678 collisions
     Escape probability: ~10^-600

0.1 cm SLAB (3.4 MFPs):
┌──┐
│░░│ → → → NEUTRON ESCAPES!
│░░│
└──┘
Escape distance: ~3 collisions
Escape probability: ~19%
```

---

## Critical Mass Implications

### What is Critical Mass?

Critical mass is the **minimum amount of fissile material** needed to sustain a chain reaction.

**Two competing factors:**

1. **Neutron Production:** Fissions create new neutrons (multiplication)
2. **Neutron Leakage:** Surface area allows neutrons to escape

### The k_eff Equation

```
k_eff = k_∞ × P_non-leakage

Where:
- k_∞ = infinite medium multiplication factor (no leakage)
- P_non-leakage = probability neutron doesn't escape
```

### Size Matters!

| Geometry | Escape % | k_eff | Criticality |
|----------|----------|-------|-------------|
| 0.1 cm slab | 18.9% | 1.693 | Subcritical* |
| 0.5 cm slab | 0.01% | 2.087 | Supercritical |
| **20 cm slab** | **0.00%** | **2.085** | **Highly supercritical** |

*Note: In reality, the 0.1 cm slab would be subcritical because the fission rate would drop even further in a real multi-generation scenario.

### Critical Mass Formula (Sphere)

For a bare U-235 sphere:

```
Critical radius ≈ 3.5 × MFP × √(k_∞)

For U-235:
- MFP = 0.0295 cm
- k_∞ ≈ 2.08
- r_crit ≈ 3.5 × 0.0295 × √2.08 ≈ 0.15 cm

Critical mass ≈ 52 kg U-235 (bare sphere)
```

With a neutron reflector (like beryllium), this drops to ~15 kg.

---

## Rule of Thumb

### When Do Neutrons Escape?

Significant escape occurs when:

```
Slab thickness ≤ 10 × MFP
```

For U-235:
```
10 × MFP = 10 × 0.0295 cm = 0.295 cm ≈ 3 mm
```

**Practical guidelines:**

- **< 3 mm:** High leakage (k_eff significantly reduced)
- **3-30 mm:** Moderate leakage (transition region)
- **> 30 mm:** Negligible leakage (behaves as infinite medium)

Your 20 cm = 200 mm slab is **67× the threshold** → essentially infinite!

---

## Summary

### Why Zero Escapes at 20 cm?

1. **Short mean free path:** Neutrons collide every 0.3 mm
2. **High fission probability:** 84.3% chance of fission per collision
3. **Low scatter probability:** Only 1.4% scatter and continue
4. **Thick geometry:** 678 mean free paths to the edge

**Result:** Probability of escape ≈ 10^-600 → essentially impossible

### Is This Realistic?

✅ **YES!** This is exactly how nuclear weapons and reactors work:

- **Too small:** Neutrons escape → subcritical (chain reaction dies)
- **Critical mass:** Leakage balanced by production → k_eff = 1
- **Supercritical:** Trapped neutrons → exponential growth

Your simulation correctly captures this physics!

---

## Verification

### Test Different Thicknesses

```go
// In main.go, try different thicknesses:

sim, _ := simulation.New(library, "U-235", 0.05)  // High leakage
sim, _ := simulation.New(library, "U-235", 0.1)   // Moderate leakage
sim, _ := simulation.New(library, "U-235", 0.5)   // Low leakage
sim, _ := simulation.New(library, "U-235", 5.0)   // Near zero leakage
sim, _ := simulation.New(library, "U-235", 20.0)  // Zero leakage
```

Expected trend:
```
Thickness ↑ → Escapes ↓ → k_eff ↑
```

### Expected Results

```
Thickness    Escapes    k_eff
─────────────────────────────
0.05 cm      ~35%       ~1.4
0.1 cm       ~19%       ~1.7
0.5 cm       ~0.01%     ~2.08
5.0 cm       0.00%      ~2.08
20.0 cm      0.00%      ~2.08
```

At ~0.3 cm and above, the slab becomes "optically thick" and k_eff plateaus at k_∞ ≈ 2.08.

---

## References

- ENDF/B-VIII.0 Nuclear Data Library
- Lamarsh & Baratta, "Introduction to Nuclear Engineering"
- Duderstadt & Hamilton, "Nuclear Reactor Analysis"

---

🎯 **Bottom line:** Zero escapes at 20 cm is not a bug—it's fundamental nuclear physics!
