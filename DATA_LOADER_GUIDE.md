# Nuclear Data Loader Implementation Guide

> How to build `models/nuclear_data.go` step by step

## Overview

The nuclear data loader reads `nuclear_data.json` and provides an easy interface to:
1. Load isotope data (cross sections, molar mass, etc.)
2. Create `Nuclide` structs from that data
3. Create `Material` structs with proper densities

---

## Understanding nuclear_data.json Structure

First, let's look at the JSON file structure:

```json
{
  "metadata": {
    "description": "Thermal neutron cross sections at 0.0253 eV (293.6 K)",
    "source": "ENDF/B-VIII.0",
    "energy_eV": 0.0253
  },
  "isotopes": {
    "U-235": {
      "name": "Uranium-235",
      "molar_mass": 235.0439299,
      "xs_fission": 584.4,
      "xs_capture": 98.81,
      "xs_elastic": 10.0,
      "xs_total": 693.21,
      "nu": 2.4367,
      "description": "Primary reactor fuel",
      "natural_abundance": 0.72,
      "density_metal": 19.1
    },
    "Pu-239": { ... },
    ...
  }
}
```

### Key Points

1. **Top level**: `metadata` and `isotopes`
2. **isotopes**: A map/dictionary where keys are isotope names ("U-235", "Pu-239")
3. **Each isotope**: Has properties like cross sections, molar mass, density

---

## Step 1: Define Go Structs to Match JSON

### Why?
Go's `json.Unmarshal()` automatically fills struct fields from JSON if the names match.

### IsotopeData Struct

This represents ONE isotope's data from the JSON:

```go
type IsotopeData struct {
    Name             string  `json:"name"`
    MolarMass        float64 `json:"molar_mass"`
    XSFission        float64 `json:"xs_fission"`
    XSCapture        float64 `json:"xs_capture"`
    XSElastic        float64 `json:"xs_elastic"`
    XSTotal          float64 `json:"xs_total"`
    Nu               float64 `json:"nu"`
    Description      string  `json:"description"`
    NaturalAbundance float64 `json:"natural_abundance"`
    DensityMetal     float64 `json:"density_metal"`
}
```

**Explanation:**
- Each field maps to a JSON property
- `` `json:"name"` `` tells Go which JSON key to read
- `float64` for numbers, `string` for text

### NuclearDataFile Struct

This represents the ENTIRE JSON file:

```go
type NuclearDataFile struct {
    Metadata struct {
        Description string  `json:"description"`
        Source      string  `json:"source"`
        EnergyEV    float64 `json:"energy_eV"`
        Units       struct {
            CrossSection string `json:"cross_section"`
            MolarMass    string `json:"molar_mass"`
            Nu           string `json:"nu"`
        } `json:"units"`
    } `json:"metadata"`
    Isotopes map[string]IsotopeData `json:"isotopes"`
}
```

**Explanation:**
- `Metadata` is a nested struct (struct inside struct)
- `Isotopes` is a **map** (like a dictionary)
  - Key: `string` (isotope name like "U-235")
  - Value: `IsotopeData` (the data for that isotope)

---

## Step 2: Create the Library Manager

### NuclearDataLibrary Struct

This will hold the loaded data and provide methods to access it:

```go
type NuclearDataLibrary struct {
    data     NuclearDataFile            // The entire JSON file
    isotopes map[string]IsotopeData     // Quick access to isotopes
}
```

**Why two fields?**
- `data`: Full JSON file (includes metadata)
- `isotopes`: Direct reference to the isotopes map for convenience

---

## Step 3: Load the JSON File

### LoadNuclearData Function

This is the main function that reads and parses the JSON:

```go
func LoadNuclearData(filepath string) (*NuclearDataLibrary, error) {
    // 1. Read the file into memory
    file, err := os.ReadFile(filepath)
    if err != nil {
        return nil, fmt.Errorf("failed to read nuclear data file: %w", err)
    }

    // 2. Parse JSON into our struct
    var data NuclearDataFile
    err = json.Unmarshal(file, &data)
    if err != nil {
        return nil, fmt.Errorf("failed to parse nuclear data JSON: %w", err)
    }

    // 3. Create and return the library
    return &NuclearDataLibrary{
        data:     data,
        isotopes: data.Isotopes, // Reference to isotopes map
    }, nil
}
```

**Step by Step:**

**1. Read the file**
```go
file, err := os.ReadFile(filepath)
```
- `os.ReadFile()` reads entire file into `file` (a byte slice `[]byte`)
- Returns error if file doesn't exist or can't be read
- `filepath` is like `"nuclear_data.json"`

**2. Parse JSON**
```go
var data NuclearDataFile
err = json.Unmarshal(file, &data)
```
- `json.Unmarshal()` converts JSON bytes into Go struct
- `&data` means "fill this struct with the JSON data"
- Go automatically maps JSON fields to struct fields using the `` `json:"..."` `` tags

**3. Create library**
```go
return &NuclearDataLibrary{
    data:     data,
    isotopes: data.Isotopes,
}, nil
```
- Create a `NuclearDataLibrary` struct
- Store the full data
- Also store direct reference to isotopes map
- Return a pointer (`*NuclearDataLibrary`) so we can call methods on it

---

## Step 4: Helper Functions

### GetIsotope - Retrieve One Isotope

```go
func (ndl *NuclearDataLibrary) GetIsotope(name string) (IsotopeData, error) {
    isotope, exists := ndl.isotopes[name]
    if !exists {
        return IsotopeData{}, fmt.Errorf("isotope %s not found in nuclear data library", name)
    }
    return isotope, nil
}
```

**Explanation:**
- `(ndl *NuclearDataLibrary)` - This is a method on NuclearDataLibrary
- `ndl.isotopes[name]` - Look up isotope in the map
- `isotope, exists := ...` - Go map lookup returns TWO values:
  - `isotope`: the data (if found)
  - `exists`: `true` if found, `false` if not
- If not found, return empty struct and error
- If found, return the isotope data

**Example usage:**
```go
u235Data, err := library.GetIsotope("U-235")
if err != nil {
    log.Fatal(err)
}
fmt.Println(u235Data.MolarMass) // 235.0439299
```

### ListIsotopes - Get All Available Names

```go
func (ndl *NuclearDataLibrary) ListIsotopes() []string {
    names := make([]string, 0, len(ndl.isotopes))
    for name := range ndl.isotopes {
        names = append(names, name)
    }
    return names
}
```

**Explanation:**
- Create empty slice with capacity = number of isotopes
- Loop through map keys (`for name := range ...`)
- Append each name to the slice
- Return the list

**Example usage:**
```go
isotopes := library.ListIsotopes()
fmt.Println(isotopes) // ["U-235", "U-238", "Pu-239", ...]
```

### GetMetadata - Get Library Info

```go
func (ndl *NuclearDataLibrary) GetMetadata() string {
    return fmt.Sprintf("%s (%s) at %.4f eV",
        ndl.data.Metadata.Description,
        ndl.data.Metadata.Source,
        ndl.data.Metadata.EnergyEV)
}
```

**Explanation:**
- Access metadata from the `data` field
- Format as nice string
- Return description like: "Thermal neutron cross sections (ENDF/B-VIII.0) at 0.0253 eV"

---

## Step 5: Create Nuclide from IsotopeData

### CreateNuclide - Convert JSON Data to Nuclide Struct

```go
func (ndl *NuclearDataLibrary) CreateNuclide(isotopeName string) (Nuclide, error) {
    // 1. Get isotope data from library
    data, err := ndl.GetIsotope(isotopeName)
    if err != nil {
        return Nuclide{}, err
    }

    // 2. Map IsotopeData fields to Nuclide fields
    return Nuclide{
        Name:                   isotopeName,
        FissionCrossSection:    data.XSFission,
        AbsorptionCrossSection: data.XSCapture,  // capture = absorption
        ScatterCrossSection:    data.XSElastic,  // elastic = scatter
        Nu:                     data.Nu,
        MolarMass:              data.MolarMass,
    }, nil
}
```

**Explanation:**

**Why map fields?**
- `IsotopeData` uses JSON naming (`XSFission`, `XSCapture`)
- `Nuclide` uses our physics naming (`FissionCrossSection`, `AbsorptionCrossSection`)
- This function translates between them

**Field mappings:**
- `XSFission` → `FissionCrossSection`
- `XSCapture` → `AbsorptionCrossSection` (radiative capture)
- `XSElastic` → `ScatterCrossSection` (elastic scattering)
- `Nu` → `Nu` (same)
- `MolarMass` → `MolarMass` (same)

---

## Step 6: Create Material from IsotopeData

### CreateMaterial - Create Material with Custom Density

```go
func (ndl *NuclearDataLibrary) CreateMaterial(isotopeName string, density float64) (Material, error) {
    // 1. Create the nuclide
    nuclide, err := ndl.CreateNuclide(isotopeName)
    if err != nil {
        return Material{}, err
    }

    // 2. Create material with given density
    material := Material{
        Nuclide: nuclide,
        Density: density,
    }

    // 3. Calculate atomic density
    material.CalcAtomicDensity()

    return material, nil
}
```

**Explanation:**
- Get nuclide data
- Create `Material` struct
- Call `CalcAtomicDensity()` to compute atoms/cm³
- Return ready-to-use material

**Example usage:**
```go
// U-235 at 19.1 g/cm³
u235, err := library.CreateMaterial("U-235", 19.1)
```

### CreateMetalMaterial - Use Default Metal Density

```go
func (ndl *NuclearDataLibrary) CreateMetalMaterial(isotopeName string) (Material, error) {
    // 1. Get isotope data
    data, err := ndl.GetIsotope(isotopeName)
    if err != nil {
        return Material{}, err
    }

    // 2. Check if metal density is available
    if data.DensityMetal == 0 {
        return Material{}, fmt.Errorf("no metal density data for %s", isotopeName)
    }

    // 3. Use CreateMaterial with default density
    return ndl.CreateMaterial(isotopeName, data.DensityMetal)
}
```

**Explanation:**
- Get isotope data
- Check if `density_metal` is set in JSON
- If yes, create material using that density
- If no, return error

**Example usage:**
```go
// U-235 at its default metal density (19.1 g/cm³)
u235, err := library.CreateMetalMaterial("U-235")
```

---

## Complete Implementation

Here's the full `nuclear_data.go` file with all pieces together:

```go
package models

import (
    "encoding/json"
    "fmt"
    "os"
)

// IsotopeData holds nuclear data for a single isotope from the JSON file
type IsotopeData struct {
    Name             string  `json:"name"`
    MolarMass        float64 `json:"molar_mass"`
    XSFission        float64 `json:"xs_fission"`
    XSCapture        float64 `json:"xs_capture"`
    XSElastic        float64 `json:"xs_elastic"`
    XSTotal          float64 `json:"xs_total"`
    Nu               float64 `json:"nu"`
    Description      string  `json:"description"`
    NaturalAbundance float64 `json:"natural_abundance"`
    DensityMetal     float64 `json:"density_metal"`
}

// NuclearDataFile represents the structure of nuclear_data.json
type NuclearDataFile struct {
    Metadata struct {
        Description string  `json:"description"`
        Source      string  `json:"source"`
        EnergyEV    float64 `json:"energy_eV"`
        Units       struct {
            CrossSection string `json:"cross_section"`
            MolarMass    string `json:"molar_mass"`
            Nu           string `json:"nu"`
        } `json:"units"`
    } `json:"metadata"`
    Isotopes map[string]IsotopeData `json:"isotopes"`
}

// NuclearDataLibrary manages loading and accessing nuclear data
type NuclearDataLibrary struct {
    data     NuclearDataFile
    isotopes map[string]IsotopeData
}

// LoadNuclearData loads the nuclear data from a JSON file
func LoadNuclearData(filepath string) (*NuclearDataLibrary, error) {
    // Read file
    file, err := os.ReadFile(filepath)
    if err != nil {
        return nil, fmt.Errorf("failed to read nuclear data file: %w", err)
    }

    // Parse JSON
    var data NuclearDataFile
    err = json.Unmarshal(file, &data)
    if err != nil {
        return nil, fmt.Errorf("failed to parse nuclear data JSON: %w", err)
    }

    // Create library
    return &NuclearDataLibrary{
        data:     data,
        isotopes: data.Isotopes,
    }, nil
}

// GetIsotope retrieves data for a specific isotope
func (ndl *NuclearDataLibrary) GetIsotope(name string) (IsotopeData, error) {
    isotope, exists := ndl.isotopes[name]
    if !exists {
        return IsotopeData{}, fmt.Errorf("isotope %s not found in nuclear data library", name)
    }
    return isotope, nil
}

// ListIsotopes returns a list of all available isotope names
func (ndl *NuclearDataLibrary) ListIsotopes() []string {
    names := make([]string, 0, len(ndl.isotopes))
    for name := range ndl.isotopes {
        names = append(names, name)
    }
    return names
}

// GetMetadata returns information about the nuclear data library
func (ndl *NuclearDataLibrary) GetMetadata() string {
    return fmt.Sprintf("%s (%s) at %.4f eV",
        ndl.data.Metadata.Description,
        ndl.data.Metadata.Source,
        ndl.data.Metadata.EnergyEV)
}

// CreateNuclide creates a Nuclide struct from isotope data
func (ndl *NuclearDataLibrary) CreateNuclide(isotopeName string) (Nuclide, error) {
    data, err := ndl.GetIsotope(isotopeName)
    if err != nil {
        return Nuclide{}, err
    }

    return Nuclide{
        Name:                   isotopeName,
        FissionCrossSection:    data.XSFission,
        AbsorptionCrossSection: data.XSCapture,
        ScatterCrossSection:    data.XSElastic,
        Nu:                     data.Nu,
        MolarMass:              data.MolarMass,
    }, nil
}

// CreateMaterial creates a Material from isotope data with custom density
func (ndl *NuclearDataLibrary) CreateMaterial(isotopeName string, density float64) (Material, error) {
    nuclide, err := ndl.CreateNuclide(isotopeName)
    if err != nil {
        return Material{}, err
    }

    material := Material{
        Nuclide: nuclide,
        Density: density,
    }

    material.CalcAtomicDensity()

    return material, nil
}

// CreateMetalMaterial creates a Material using the default metal density from the database
func (ndl *NuclearDataLibrary) CreateMetalMaterial(isotopeName string) (Material, error) {
    data, err := ndl.GetIsotope(isotopeName)
    if err != nil {
        return Material{}, err
    }

    if data.DensityMetal == 0 {
        return Material{}, fmt.Errorf("no metal density data for %s", isotopeName)
    }

    return ndl.CreateMaterial(isotopeName, data.DensityMetal)
}
```

---

## How to Use It

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "yourproject/models"
)

func main() {
    // 1. Load nuclear data
    library, err := models.LoadNuclearData("nuclear_data.json")
    if err != nil {
        log.Fatal("Error loading data:", err)
    }

    // 2. Get library info
    fmt.Println(library.GetMetadata())
    fmt.Println("Available:", library.ListIsotopes())

    // 3. Get raw isotope data
    u235Data, _ := library.GetIsotope("U-235")
    fmt.Printf("U-235 molar mass: %.2f g/mol\n", u235Data.MolarMass)

    // 4. Create a nuclide
    u235Nuclide, _ := library.CreateNuclide("U-235")
    fmt.Printf("σ_fission: %.1f barns\n", u235Nuclide.FissionCrossSection)

    // 5. Create a material (custom density)
    material, _ := library.CreateMaterial("U-235", 19.1)
    fmt.Printf("Atomic density: %.3e atoms/cm³\n", material.AtomicDensity)

    // 6. Create a material (default metal density)
    u235Metal, _ := library.CreateMetalMaterial("U-235")
    fmt.Printf("Default density: %.1f g/cm³\n", u235Metal.Density)
}
```

---

## Testing Your Implementation

Create a simple test:

```go
package main

import (
    "fmt"
    "log"
    "makarov-chains/models"
)

func main() {
    library, err := models.LoadNuclearData("nuclear_data.json")
    if err != nil {
        log.Fatal(err)
    }

    // Test 1: Load worked
    fmt.Println("✓ Loaded:", library.GetMetadata())

    // Test 2: Can list isotopes
    isotopes := library.ListIsotopes()
    fmt.Printf("✓ Found %d isotopes\n", len(isotopes))

    // Test 3: Can get isotope data
    u235Data, _ := library.GetIsotope("U-235")
    fmt.Printf("✓ U-235 molar mass: %.2f g/mol\n", u235Data.MolarMass)

    // Test 4: Can create nuclide
    u235, _ := library.CreateNuclide("U-235")
    fmt.Printf("✓ U-235 fission XS: %.1f barns\n", u235.FissionCrossSection)

    // Test 5: Can create material
    mat, _ := library.CreateMetalMaterial("U-235")
    fmt.Printf("✓ U-235 atomic density: %.3e atoms/cm³\n", mat.AtomicDensity)

    // Test 6: Cross section calculations work
    Σ_f := mat.MacroscopicCrossSection(mat.Nuclide.FissionCrossSection)
    fmt.Printf("✓ Σ_fission: %.2f cm⁻¹\n", Σ_f)

    fmt.Println("\n🎉 All tests passed!")
}
```

Expected output:
```
✓ Loaded: Thermal neutron cross sections at 0.0253 eV (293.6 K) (ENDF/B-VIII.0) at 0.0253 eV
✓ Found 10 isotopes
✓ U-235 molar mass: 235.04 g/mol
✓ U-235 fission XS: 584.4 barns
✓ U-235 atomic density: 4.893e+22 atoms/cm³
✓ Σ_fission: 28.59 cm⁻¹

🎉 All tests passed!
```

---

## Common Errors and Solutions

### Error: "no such file or directory"
```
failed to read nuclear data file: open nuclear_data.json: no such file or directory
```
**Solution:** Check the file path. Use absolute path or make sure you're running from the correct directory.

### Error: "cannot unmarshal"
```
failed to parse nuclear data JSON: json: cannot unmarshal...
```
**Solution:** JSON structure doesn't match Go struct. Check:
- Spelling in `` `json:"..."` `` tags
- JSON file is valid (use a JSON validator)

### Error: "isotope not found"
```
isotope U235 not found in nuclear data library
```
**Solution:** Check spelling. It's "U-235" with hyphen, not "U235".

---

## Summary

**What each part does:**

1. **IsotopeData** - Struct matching one isotope's JSON data
2. **NuclearDataFile** - Struct matching entire JSON file
3. **NuclearDataLibrary** - Manager that holds loaded data
4. **LoadNuclearData()** - Reads JSON file and parses it
5. **GetIsotope()** - Gets one isotope's data
6. **CreateNuclide()** - Converts IsotopeData → Nuclide
7. **CreateMaterial()** - Creates Material with density
8. **CreateMetalMaterial()** - Creates Material using default density

Now you can implement it yourself using this guide as reference! 🚀
