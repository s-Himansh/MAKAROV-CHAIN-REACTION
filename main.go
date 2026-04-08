package main

import (
	"fmt"
	"log"
	loadingnucleardata "makarov-chains/loading_nuclear_data"
	"makarov-chains/simulation"
)

func main() {
	// Load nuclear data
	library, err := loadingnucleardata.LoadNuclearData("nuclear_data.json")
	if err != nil {
		log.Fatal("Failed to load nuclear data:", err)
	}

	fmt.Println("Loaded:", library.RetrieveMetaData())
	fmt.Println()

	// Create simulation
	// 20 cm thick U-235 slab
	sim, err := simulation.New(library, "U-235", 20.0)
	if err != nil {
		log.Fatal("Failed to create simulation:", err)

		return
	}

	// run simulation
	sim.Run()

	// print results
	sim.PrintResults()
}
