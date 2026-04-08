# Nuclear Data Sources Reference

> Where the data in `nuclear_data.json` comes from

## Quick Answer

All data in `nuclear_data.json` comes from **ENDF/B-VIII.0** (Evaluated Nuclear Data File, version 8.0), the official U.S. nuclear data library maintained by Brookhaven National Laboratory.

**Specific values are for thermal neutrons at 0.0253 eV (room temperature).**

---

## Primary Source: ENDF/B-VIII.0

**What is it?**
- Official nuclear data library used by nuclear engineers and physicists
- Maintained by National Nuclear Data Center (NNDC) at Brookhaven National Lab
- Contains experimental measurements compiled from decades of research
- Updated periodically (latest: ENDF/B-VIII.0 released in 2018)

**Website:** https://www.nndc.bnl.gov/endf/

---

## How to Find This Data Yourself

### Method 1: JANIS (Recommended for Beginners)

**JANIS** = Java-based Nuclear Information Software

**Website:** https://www.oecd-nea.org/janisweb/

**Steps:**
1. Go to https://www.oecd-nea.org/janisweb/
2. Select isotope (e.g., "U-235")
3. Select "Cross Section" → "Neutron Cross Section"
4. Select energy = 0.0253 eV (thermal)
5. See values for:
   - Fission (MT=18)
   - Capture (MT=102)
   - Elastic (MT=2)
   - nu (neutrons per fission)

**Example for U-235:**
```
Energy: 0.0253 eV (thermal)
Fission (σf):    584.4 barns
Capture (σγ):    98.81 barns
Elastic (σel):   10.0 barns
Total (σtot):    693.21 barns
nu (ν):          2.4367
```

### Method 2: NNDC Sigma (Cross Section Plotter)

**Website:** https://www.nndc.bnl.gov/sigma/

**Steps:**
1. Go to https://www.nndc.bnl.gov/sigma/
2. Enter isotope (e.g., "U-235")
3. Select "ENDF/B-VIII.0"
4. Select reaction type:
   - MT=18 → Fission
   - MT=102 → Radiative capture
   - MT=2 → Elastic scattering
5. Plot cross section vs energy
6. Read value at 0.0253 eV

### Method 3: IAEA LiveChart

**Website:** https://nds.iaea.org/relnsd/vcharthtml/VChartHTML.html

**Steps:**
1. Go to https://nds.iaea.org/relnsd/vcharthtml/VChartHTML.html
2. Find element (e.g., Uranium → U-235)
3. Click on isotope box
4. View "Neutron Cross Section" tab
5. Read thermal (0.0253 eV) values

### Method 4: Wikipedia (Quick Reference)

Many isotopes have thermal cross sections listed on Wikipedia:

**Example:** https://en.wikipedia.org/wiki/Uranium-235

Search for "cross section" on the page to find thermal values.

⚠️ **Warning:** Wikipedia is convenient but not always complete. Use ENDF sources for serious work.

---

## Data Breakdown by Isotope

### U-235 (Uranium-235)

```json
{
  "molar_mass": 235.0439299,        // From NIST Atomic Weights
  "xs_fission": 584.4,              // ENDF/B-VIII.0, MT=18, 0.0253 eV
  "xs_capture": 98.81,              // ENDF/B-VIII.0, MT=102, 0.0253 eV
  "xs_elastic": 10.0,               // ENDF/B-VIII.0, MT=2, 0.0253 eV
  "xs_total": 693.21,               // Sum of all reactions
  "nu": 2.4367,                     // ENDF/B-VIII.0, average ν
  "density_metal": 19.1             // CRC Handbook, uranium metal at 293 K
}
```

**Sources:**
- **Cross sections:** ENDF/B-VIII.0 via JANIS (https://www.oecd-nea.org/janisweb/)
- **Molar mass:** NIST Atomic Weights (https://www.nist.gov/pml/atomic-weights-and-isotopic-compositions)
- **Density:** CRC Handbook of Chemistry and Physics
- **nu:** ENDF/B-VIII.0, prompt neutron multiplicity

**How to verify:**
1. Go to JANIS: https://www.oecd-nea.org/janisweb/
2. Search "U-235"
3. Cross Section → Neutron Cross Section
4. Library: ENDF/B-VIII.0
5. Energy: 0.0253 eV
6. Check values match

---

### Pu-239 (Plutonium-239)

```json
{
  "molar_mass": 239.0521634,        // NIST
  "xs_fission": 747.4,              // ENDF/B-VIII.0, thermal
  "xs_capture": 268.8,              // ENDF/B-VIII.0, thermal
  "xs_elastic": 7.97,               // ENDF/B-VIII.0, thermal
  "nu": 2.8758,                     // ENDF/B-VIII.0
  "density_metal": 19.86            // CRC Handbook
}
```

**Note:** Pu-239 values vary between ENDF versions. These are from ENDF/B-VIII.0 (2018).

---

### U-238 (Uranium-238)

```json
{
  "xs_fission": 0.00001,            // Essentially zero at thermal
  "xs_capture": 2.717,              // ENDF/B-VIII.0
  "xs_elastic": 8.871,              // ENDF/B-VIII.0
  "nu": 0.0                         // Not fissile at thermal energies
}
```

**Important:** U-238 is NOT fissile at thermal energies. It only fissions with fast neutrons (>1 MeV).

---

### H-1 (Hydrogen-1)

```json
{
  "xs_capture": 0.3326,             // ENDF/B-VIII.0 (n,γ reaction)
  "xs_elastic": 20.491,             // ENDF/B-VIII.0 (scattering)
  "xs_fission": 0.0                 // Hydrogen doesn't fission
}
```

**Use case:** Water moderator (H₂O) in reactors

---

### B-10 (Boron-10)

```json
{
  "xs_capture": 3835.0,             // ENDF/B-VIII.0 (VERY high!)
  "xs_elastic": 2.0,                // ENDF/B-VIII.0
  "xs_fission": 0.0
}
```

**Note:** B-10 has one of the highest thermal neutron capture cross sections! Used in control rods.

**Reaction:** B-10 + n → Li-7 + α (helium-4)

---

## Accessing ENDF Data Programmatically

### Option 1: Download Pre-Processed OpenMC Data

OpenMC provides HDF5 files with ENDF data already processed:

```bash
# Download OpenMC cross section library
wget https://anl.box.com/shared/static/9igk353zpy8fn9ttvtrqgzvw1vtejoz6.xz -O nndc_hdf5.tar.xz
tar -xvf nndc_hdf5.tar.xz
```

This gives you HDF5 files for 400+ isotopes!

### Option 2: Use Python Libraries

**PyNE (Python for Nuclear Engineering):**
```python
from pyne import data

# Get U-235 thermal fission cross section
xs_fission = data.sigma_f('U235', 0.0253)  # eV
print(f"σ_f = {xs_fission} barns")
```

**OpenMC Python API:**
```python
import openmc.data

# Load U-235 data
u235 = openmc.data.IncidentNeutron.from_njoy('U235.endf')

# Get thermal cross sections
E = 0.0253  # eV
xs_fission = u235[18].xs['294K'](E)  # MT=18 is fission
```

### Option 3: Download Raw ENDF Files

Direct download of ENDF/B-VIII.0:
- **Website:** https://www.nndc.bnl.gov/endf-b8.0/download.html
- **Format:** Text files in ENDF-6 format (complex!)
- **Tool needed:** NJOY to process them

---

## What Each Value Means

### Cross Sections (barns)

**1 barn = 10⁻²⁴ cm²**

Think of it as the "target size" the nucleus presents to neutrons.

| Reaction | Symbol | MT Number | Meaning |
|----------|--------|-----------|---------|
| **Fission** | σ_f | MT=18 | Nucleus splits, releases neutrons |
| **Capture** | σ_γ | MT=102 | Neutron absorbed, γ-ray emitted |
| **Elastic** | σ_el | MT=2 | Neutron bounces off, no energy loss |
| **Total** | σ_tot | MT=1 | Sum of all possible reactions |

### nu (ν)

Average number of neutrons released per fission event.

**Example:** U-235 has ν = 2.4367
- ~43% of fissions produce 2 neutrons
- ~57% of fissions produce 3 neutrons
- Average = 2.4367

### Molar Mass

From **NIST Atomic Weights and Isotopic Compositions:**
- Website: https://www.nist.gov/pml/atomic-weights-and-isotopic-compositions
- Based on carbon-12 scale
- Unit: g/mol (grams per mole)

**Example:** U-235 = 235.0439299 g/mol

### Density (Metal)

Physical density of the pure metal at room temperature (293 K).

**Source:** CRC Handbook of Chemistry and Physics

**Example:** Uranium metal = 19.1 g/cm³ (very dense!)

---

## Energy Note: Why 0.0253 eV?

**0.0253 eV = thermal neutron energy at room temperature (293.6 K)**

**Physics:**
```
E = (3/2) × k_B × T

k_B = Boltzmann constant = 8.617 × 10⁻⁵ eV/K
T = 293.6 K (20.45°C)

E = (3/2) × 8.617×10⁻⁵ × 293.6 = 0.0253 eV
```

This is the **standard reference energy** for thermal neutron data.

**Neutron velocity at 0.0253 eV:**
```
v = sqrt(2E/m) ≈ 2200 m/s

That's why you'll see "2200 m/s cross sections" in literature!
```

---

## Validation: Cross-Check Your Data

### Method 1: Compare Libraries

Different libraries should give similar (but not identical) values:

| Library | Version | U-235 σ_f (thermal) | U-235 ν |
|---------|---------|---------------------|---------|
| ENDF/B-VIII.0 | 2018 | 584.4 barns | 2.4367 |
| ENDF/B-VII.1 | 2011 | 585.0 barns | 2.4355 |
| JEFF-3.3 | 2017 | 584.8 barns | 2.4374 |
| JENDL-5.0 | 2021 | 585.1 barns | 2.4368 |

**Small differences** (< 1%) are normal due to different experimental analyses.

### Method 2: Check Critical Mass

If your simulation gives **k_eff ≈ 1.0** at known critical configurations, your data is correct!

**U-235 bare sphere critical mass:**
- **Measured:** 52 kg (radius ≈ 8.7 cm)
- **Your simulation should give:** k_eff ≈ 1.0 at this size

---

## References

### Official Nuclear Data Libraries

1. **ENDF/B-VIII.0** (USA)
   - https://www.nndc.bnl.gov/endf-b8.0/

2. **JEFF-3.3** (Europe)
   - https://www.oecd-nea.org/dbdata/jeff/

3. **JENDL-5.0** (Japan)
   - https://wwwndc.jaea.go.jp/jendl/jendl.html

4. **CENDL** (China)
   - http://www.nuclear.csdb.cn/

### Interactive Tools

1. **JANIS** (Best for looking up values)
   - https://www.oecd-nea.org/janisweb/

2. **NNDC Sigma** (Plot cross sections)
   - https://www.nndc.bnl.gov/sigma/

3. **IAEA LiveChart** (Nuclear properties)
   - https://nds.iaea.org/relnsd/vcharthtml/VChartHTML.html

### Textbooks

1. **Lamarsh & Baratta** - "Introduction to Nuclear Engineering"
   - Has tables of thermal cross sections in Appendix

2. **Duderstadt & Hamilton** - "Nuclear Reactor Analysis"
   - Comprehensive nuclear data tables

### Standards

1. **NIST Atomic Weights**
   - https://www.nist.gov/pml/atomic-weights-and-isotopic-compositions

2. **CODATA Fundamental Constants**
   - https://physics.nist.gov/cuu/Constants/

---

## Summary

**Where did the data come from?**

✅ **Cross sections (σ)** → ENDF/B-VIII.0 via JANIS web interface
✅ **Neutron multiplicity (ν)** → ENDF/B-VIII.0
✅ **Molar masses** → NIST Atomic Weights
✅ **Densities** → CRC Handbook of Chemistry and Physics
✅ **Energy** → 0.0253 eV (thermal neutron standard)

**Is this data trustworthy?**

YES! ENDF is:
- Used by nuclear power industry
- Used by national laboratories (Los Alamos, Oak Ridge, etc.)
- Used by nuclear weapons labs
- Validated against thousands of experiments
- Peer-reviewed by international experts

**Can you verify it?**

YES! Go to JANIS (https://www.oecd-nea.org/janisweb/) and look up any isotope to check the values.

---

## Quick Lookup Table (Common Isotopes)

All values at **0.0253 eV (thermal)**:

| Isotope | σ_fission | σ_capture | σ_elastic | ν | Use |
|---------|-----------|-----------|-----------|---|-----|
| **U-235** | 584.4 | 98.81 | 10.0 | 2.44 | Reactor fuel |
| **U-238** | ~0 | 2.72 | 8.87 | - | Fertile material |
| **Pu-239** | 747.4 | 268.8 | 7.97 | 2.88 | Weapons/fuel |
| **Pu-240** | 0.064 | 289.6 | 1.6 | 2.9 | Contaminant |
| **H-1** | 0 | 0.33 | 20.49 | - | Moderator (water) |
| **C-12** | 0 | 0.004 | 4.75 | - | Moderator (graphite) |
| **B-10** | 0 | **3835** | 2.0 | - | Control rods |
| **Cd-113** | 0 | **20600** | 5.0 | - | Control rods |

Source: ENDF/B-VIII.0 (2018)

---

🎉 **Now you know exactly where every number in `nuclear_data.json` comes from!**
