package simulation

import (
	"fmt"
	"makarov-chains/models"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	library "makarov-chains/loading_nuclear_data"
)

const maxCollisions = 10000

type Simulation struct {
	Material *models.Material
	Geometry *models.Geometry
	Params   *models.SimParams
	Tally    *models.Tally
	Quiet    bool

	ΣFission    float64
	ΣAbsorption float64
	ΣScatter    float64
	ΣTotal      float64
}

func (s *Simulation) SetQuiet(quiet bool) {
	s.Quiet = quiet
}

func New(library *library.NuclearDataLibrary, isotope string, thickness float64) (*Simulation, error) {
	material, err := library.CreateMetalMaterial(isotope)
	if err != nil {
		return nil, err
	}

	sim := &Simulation{
		Material: material,
		Geometry: &models.Geometry{Thickness: thickness, Type: models.Slab},
		Params:   &models.SimParams{NumParticles: 1000, NumGenerations: 50, Seed: time.Now().UnixNano()},
		Tally:    &models.Tally{NeutronsPerGen: make([]int, 0)},
	}

	sim.ΣFission = material.MacroScopicCrossSection(material.Nuclide.FissionCrossSection)
	sim.ΣAbsorption = material.MacroScopicCrossSection(material.Nuclide.AbsorptionCrossSection)
	sim.ΣScatter = material.MacroScopicCrossSection(material.Nuclide.ScatterCrossSection)
	sim.ΣTotal = material.TotalCrossSections()

	return sim, nil
}

func (s *Simulation) Run() {
	rng := rand.New(rand.NewPCG(uint64(s.Params.Seed), 0))

	neutronBank := s.Params.NumParticles

	geoDesc := fmt.Sprintf("1D slab, %.1f cm thick", s.Geometry.Thickness)
	if s.Geometry.Type == models.Sphere {
		geoDesc = fmt.Sprintf("sphere, %.1f cm radius", s.Geometry.Thickness)
	}

	if !s.Quiet {
		fmt.Printf("Starting simulation: %s, %s\n", s.Material.Nuclide.Name, geoDesc)
		fmt.Printf("Σ_total = %.2f cm⁻¹, MFP = %.4f cm\n\n", s.ΣTotal, 1.0/s.ΣTotal)
	}

	for gen := range s.Params.NumGenerations {
		if neutronBank == 0 {
			if !s.Quiet {
				fmt.Println("Chain reaction died (subcritical)")
			}
			break
		}

		nextBank := 0
		fissions := 0
		absorptions := 0
		escapes := 0

		for n := 0; n < neutronBank; n++ {
			outcome := s.trackNeutron(rng)

			switch outcome.Type {
			case models.Fission:
				fissions++
				nextBank += outcome.NewNeutrons
				s.Tally.TotalFissions++
			case models.Absorption:
				absorptions++
				s.Tally.TotalAbsorptions++
			case models.Escape:
				escapes++
				s.Tally.TotalEscapes++
			}
		}

		s.Tally.NeutronsPerGen = append(s.Tally.NeutronsPerGen, nextBank)

		k_gen := float64(nextBank) / float64(neutronBank)

		if !s.Quiet {
			fmt.Printf("Gen %2d: %5d → %5d neutrons (F:%4d A:%4d E:%4d) k=%.3f\n", gen+1, neutronBank, nextBank, fissions, absorptions, escapes, k_gen)
		}

		if nextBank > s.Params.NumParticles {
			neutronBank = s.Params.NumParticles
		} else if nextBank == 0 {
			neutronBank = 0
		} else {
			neutronBank = nextBank
		}
	}

	s.calculateKeff()
}

func (s *Simulation) PrintResults() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("SIMULATION RESULTS")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Material: %s\n", s.Material.Nuclide.Name)

	if s.Geometry.Type == models.Sphere {
		fmt.Printf("Geometry: sphere, %.2f cm radius\n", s.Geometry.Thickness)
	} else {
		fmt.Printf("Geometry: 1D slab, %.2f cm thick\n", s.Geometry.Thickness)
	}

	fmt.Printf("\nTotal events:\n")
	fmt.Printf("  Fissions:    %d\n", s.Tally.TotalFissions)
	fmt.Printf("  Absorptions: %d\n", s.Tally.TotalAbsorptions)
	fmt.Printf("  Escapes:     %d\n", s.Tally.TotalEscapes)
	fmt.Printf("\nk_effective: %.4f\n", s.Tally.KEff)

	if s.Tally.KEff > 1.0 {
		fmt.Println("→ SUPERCRITICAL (growing reaction)")
	} else if s.Tally.KEff < 1.0 {
		fmt.Println("→ SUBCRITICAL (dying reaction)")
	} else {
		fmt.Println("→ CRITICAL (sustained reaction)")
	}
}

func (s *Simulation) calculateKeff() {
	if len(s.Tally.NeutronsPerGen) < 2 {
		s.Tally.KEff = 0
		return
	}

	skipGens := 10
	if len(s.Tally.NeutronsPerGen) < 20 {
		skipGens = len(s.Tally.NeutronsPerGen) / 2
	}

	sum := 0.0
	count := 0

	for i := skipGens; i < len(s.Tally.NeutronsPerGen); i++ {
		var neutronsIn int
		if i == 0 {
			neutronsIn = s.Params.NumParticles
		} else {
			prevGen := s.Tally.NeutronsPerGen[i-1]
			if prevGen > s.Params.NumParticles {
				neutronsIn = s.Params.NumParticles
			} else {
				neutronsIn = prevGen
			}
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

func (s *Simulation) sampleCollisionDistance(rng *rand.Rand) float64 {
	ξ := rng.Float64()
	λ := 1.0 / s.ΣTotal
	return -λ * math.Log(ξ)
}

func (s *Simulation) sampleInteractionType(rng *rand.Rand) models.InteractionType {
	ξ := rng.Float64()

	fissionProbability := s.ΣFission / s.ΣTotal
	absorptionProbability := s.ΣAbsorption / s.ΣTotal

	switch {
	case ξ < fissionProbability:
		return models.Fission
	case ξ < fissionProbability+absorptionProbability:
		return models.Absorption
	default:
		return models.Scatter
	}
}

func (s *Simulation) sampleNeutronEmission(rng *rand.Rand) int {
	nu := s.Material.Nuclide.Nu
	baseNeutrons := int(nu)
	fraction := nu - float64(baseNeutrons)

	if rng.Float64() < fraction {
		return baseNeutrons + 1
	}
	return baseNeutrons
}

func (s *Simulation) sampleIsotropicDirection(rng *rand.Rand) (float64, float64, float64) {
	cosTheta := 2*rng.Float64() - 1
	sinTheta := math.Sqrt(1 - cosTheta*cosTheta)
	phi := 2 * math.Pi * rng.Float64()

	dx := sinTheta * math.Cos(phi)
	dy := sinTheta * math.Sin(phi)
	dz := cosTheta

	return dx, dy, dz
}

func (s *Simulation) trackNeutron(rng *rand.Rand) *models.NeutronOutcome {
	if s.Geometry.Type == models.Sphere {
		return s.trackNeutron3D(rng)
	}
	return s.trackNeutron1D(rng)
}

func (s *Simulation) trackNeutron1D(rng *rand.Rand) *models.NeutronOutcome {
	position := s.Geometry.Thickness / 2.0
	slabThickness := s.Geometry.Thickness

	for i := range maxCollisions {
		distance := s.sampleCollisionDistance(rng)

		if rng.Float64() < 0.5 {
			position += distance
		} else {
			position -= distance
		}

		if position < 0 || position > slabThickness {
			return &models.NeutronOutcome{Type: models.Escape}
		}

		interactionType := s.sampleInteractionType(rng)

		switch interactionType {
		case models.Fission:
			return &models.NeutronOutcome{Type: models.Fission, NewNeutrons: s.sampleNeutronEmission(rng)}
		case models.Absorption:
			return &models.NeutronOutcome{Type: models.Absorption}
		default:
			_ = i
			continue
		}
	}

	return &models.NeutronOutcome{Type: models.Absorption}
}

func (s *Simulation) trackNeutron3D(rng *rand.Rand) *models.NeutronOutcome {
	radius := s.Geometry.Thickness
	x, y, z := 0.0, 0.0, 0.0

	for i := range maxCollisions {
		distance := s.sampleCollisionDistance(rng)
		dx, dy, dz := s.sampleIsotropicDirection(rng)

		x += distance * dx
		y += distance * dy
		z += distance * dz

		r := math.Sqrt(x*x + y*y + z*z)
		if r > radius {
			return &models.NeutronOutcome{Type: models.Escape}
		}

		interactionType := s.sampleInteractionType(rng)

		switch interactionType {
		case models.Fission:
			return &models.NeutronOutcome{Type: models.Fission, NewNeutrons: s.sampleNeutronEmission(rng)}
		case models.Absorption:
			return &models.NeutronOutcome{Type: models.Absorption}
		default:
			_ = i
			continue
		}
	}

	return &models.NeutronOutcome{Type: models.Absorption}
}
