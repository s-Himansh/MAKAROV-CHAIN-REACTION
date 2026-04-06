package main

import (
	"fmt"
	"log"
	library "makarov-chains/loading_nuclear_data"
)

func main() {
	library, err := library.LoadNuclearData("./nuclear_data.json")
	if err != nil {
		log.Fatal("Error loading data:", err)

		return
	}

	fmt.Println(library.RetrieveMetaData())
	fmt.Println("Available:", library.ListIsotopeNames())

	u235Data, _ := library.RetrieveIsotope("U-235")
	fmt.Printf("U-235 molar mass: %.2f g/mol\n", u235Data.MolarMass)

	u235Nuclide, _ := library.CreateNuclide("U-235")
	fmt.Printf("σ_fission: %.1f barns\n", u235Nuclide.FissionCrossSection)

	material, _ := library.CreateMaterial("U-235", 19.1)
	fmt.Printf("Atomic density: %.3e atoms/cm³\n", material.AtomicDensity)

	u235Metal, _ := library.CreateMetalMaterial("U-235")
	fmt.Printf("Default density: %.1f g/cm³\n", u235Metal.Density)
}
