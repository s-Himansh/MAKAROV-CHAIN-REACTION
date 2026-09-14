package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	loadingnucleardata "makarov-chains/loading_nuclear_data"
	"makarov-chains/models"
	"makarov-chains/simulation"
)

type Results struct {
	Material      string  `json:"material"`
	GeometryType  string  `json:"geometry"`
	Thickness     float64 `json:"thickness"`
	NumParticles  int     `json:"num_particles"`
	NumGenerations int    `json:"num_generations"`
	Seed          int64   `json:"seed"`
	TotalFissions int     `json:"total_fissions"`
	TotalAbsorptions int  `json:"total_absorptions"`
	TotalEscapes  int     `json:"total_escapes"`
	KEff          float64 `json:"k_effective"`
	Status        string  `json:"status"`
	NeutronsPerGen []int  `json:"neutrons_per_gen"`
}

func main() {
	isotope := flag.String("isotope", "U-235", "Isotope to simulate (U-235, U-238, Pu-239, Pu-241)")
	thickness := flag.Float64("thickness", 20.0, "Slab thickness or sphere radius in cm")
	particles := flag.Int("particles", 1000, "Neutrons per generation")
	generations := flag.Int("generations", 50, "Number of generations to run")
	geometry := flag.String("geometry", "slab", "Geometry type: slab or sphere")
	seed := flag.Int64("seed", 0, "Random seed (0 = auto)")
	output := flag.String("output", "", "Output file path (.json or .csv)")
	quiet := flag.Bool("quiet", false, "Suppress generation-by-generation output")

	flag.Parse()

	library, err := loadingnucleardata.LoadNuclearData("nuclear_data.json")
	if err != nil {
		log.Fatal("Failed to load nuclear data:", err)
	}

	fmt.Println("Loaded:", library.RetrieveMetaData())
	fmt.Println()

	geoType := models.Slab
	if strings.ToLower(*geometry) == "sphere" {
		geoType = models.Sphere
	}

	sim, err := simulation.New(library, *isotope, *thickness)
	if err != nil {
		log.Fatal("Failed to create simulation:", err)
	}

	sim.Params.NumParticles = *particles
	sim.Params.NumGenerations = *generations
	sim.Geometry.Type = geoType
	if *seed != 0 {
		sim.Params.Seed = *seed
	}

	if *quiet {
		sim.SetQuiet(true)
	}

	sim.Run()
	sim.PrintResults()

	if *output != "" {
		results := Results{
			Material:         sim.Material.Nuclide.Name,
			GeometryType:     strings.ToLower(*geometry),
			Thickness:        *thickness,
			NumParticles:     *particles,
			NumGenerations:   *generations,
			Seed:             sim.Params.Seed,
			TotalFissions:    sim.Tally.TotalFissions,
			TotalAbsorptions: sim.Tally.TotalAbsorptions,
			TotalEscapes:     sim.Tally.TotalEscapes,
			KEff:             sim.Tally.KEff,
			NeutronsPerGen:   sim.Tally.NeutronsPerGen,
		}

		switch {
		case sim.Tally.KEff > 1.0:
			results.Status = "supercritical"
		case sim.Tally.KEff < 1.0:
			results.Status = "subcritical"
		default:
			results.Status = "critical"
		}

		if strings.HasSuffix(*output, ".csv") {
			writeCSV(*output, results)
		} else {
			writeJSON(*output, results)
		}
		fmt.Printf("\nResults saved to %s\n", *output)
	}
}

func writeJSON(path string, r Results) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal JSON:", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatal("Failed to write file:", err)
	}
}

func writeCSV(path string, r Results) {
	var sb strings.Builder
	sb.WriteString("generation,neutrons\n")
	for i, n := range r.NeutronsPerGen {
		sb.WriteString(fmt.Sprintf("%d,%d\n", i+1, n))
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		log.Fatal("Failed to write file:", err)
	}
}
