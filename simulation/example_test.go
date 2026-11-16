package simulation_test

import (
	"fmt"

	loadingnucleardata "github.com/s-Himansh/MAKAROV-CHAIN-REACTION/loading_nuclear_data"
	"github.com/s-Himansh/MAKAROV-CHAIN-REACTION/models"
	"github.com/s-Himansh/MAKAROV-CHAIN-REACTION/simulation"
)

func Example() {
	lib, err := loadingnucleardata.LoadNuclearData("../nuclear_data.json")
	if err != nil {
		panic(err)
	}
	sim, err := simulation.New(lib, "U-235", 0.1)
	if err != nil {
		panic(err)
	}
	sim.Geometry.Type = models.Sphere
	sim.Params.NumParticles = 2000
	sim.Params.NumGenerations = 30
	sim.Params.Seed = 42

	sim.Run()

	fmt.Printf("k_eff = %.1f\n", sim.Tally.KEff)
	// Output: k_eff = 1.8
}
