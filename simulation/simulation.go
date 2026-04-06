package simulation

import (
	"makarov-chains/models"
	"math"
	"math/rand/v2"
	"time"

	library "makarov-chains/loading_nuclear_data"
)

type Simulation struct {
	// these are set of configurations set while the chain reaction is running
	Material *models.Material
	Geometry *models.Geometry
	Params   *models.SimParams

	// results of the reaction
	Tally *models.Tally

	// Pre-calculated macroscopic cross sections (probability of behaviour neutron will show while interacting with material)
	ΣFission    float64 // cm⁻¹
	ΣAbsorption float64 // cm⁻¹
	ΣScatter    float64 // cm⁻¹
	ΣTotal      float64 // cm⁻¹
}

// NewSimulation creates and initializes a simulation
func New(library *library.NuclearDataLibrary, isotope string, thickness float64) (*Simulation, error) {
	material, err := library.CreateMetalMaterial(isotope)
	if err != nil {
		return nil, err
	}

	sim := &Simulation{Material: material, Geometry: &models.Geometry{Volume: thickness},
		Params: &models.SimParams{NumParticles: 1000, NumGenerations: 50, Seed: time.Now().UnixNano()}, Tally: &models.Tally{NeutronsPerGen: make([]int, 0)}}

	sim.ΣFission = material.MacroScopicCrossSection(material.Nuclide.FissionCrossSection)
	sim.ΣAbsorption = material.MacroScopicCrossSection(material.Nuclide.AbsorptionCrossSection)
	sim.ΣScatter = material.MacroScopicCrossSection(material.Nuclide.ScatterCrossSection)

	sim.ΣTotal = material.TotalCrossSections()

	return sim, nil
}

// sampleCollisonDistance measures the distance to the next collison per neutron
func (s *Simulation) sampleCollisonDistance() float64 {
	// using exponential distribution: P(d) = Σ × e^(-Σd)

	// Random [0, 1)
	ξ := rand.Float64()

	// Mean free path [cm]
	λ := 1.0 / s.ΣTotal

	return -λ * math.Log(ξ) // Exponential sampling
}

func (s *Simulation) sampleInteractionType() models.InteractionType {
	// Random [0, 1)
	ξ := rand.Float64()

	// Calculate probabilities
	fissionProbability := s.ΣFission / s.ΣTotal
	absorptionProbability := s.ΣAbsorption / s.ΣTotal

	// Sample from discrete distribution
	switch {
	case ξ < fissionProbability:
		return models.Fission
	case ξ < fissionProbability+absorptionProbability:
		return models.Absorption
	default:
		return models.Scatter
	}
}

// sampleNeutronEmission samples number of neutrons from fission
// Uses simple interpolation between floor and ceiling of nu
func (s *Simulation) sampleNeutronEmission() int {
	nu := s.Material.Nuclide.Nu

	// count cannot be in decimals so rounding it off
	baseNeutrons := int(nu)

	// fraction of number that was rounded off
	fraction := nu - float64(baseNeutrons)

	// the round off value works on the probability of fraction being < or > than the randomly generated using rand package
	if rand.Float64() < fraction {
		return baseNeutrons + 1
	}
	return baseNeutrons
}

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

// trackNeutron follows one neutron through material until it's absorbed, fissions, or escapes
func (s *Simulation) trackNeutron() *models.NeutronOutcome {
	// start exactly at the center
	position := s.Geometry.Volume / 2.0
	slabThickness := s.Geometry.Volume

	for {
		// sample distance to next collison
		distance := s.sampleCollisonDistance()

		// on a 1 dimensional slab, neutron can either go left or right, so the probability to that direc. is 50-50
		// move the neutron the calculated direction in run-time
		if rand.Float64() < 0.5 {
			position += distance
		} else {
			position -= distance
		}

		// in case neutron escaped the material, that case we need to mark it as escaped and return it's processing
		if position < 0 || position > slabThickness {
			return &models.NeutronOutcome{Type: models.Escape}
		}

		// neutron collided with the isotope, we need to define it's interaction type with the material
		interactionType := s.sampleInteractionType()

		// handle interactions
		switch interactionType {
		case models.Fission:
			return &models.NeutronOutcome{Type: models.Fission, NewNeutrons: s.sampleNeutronEmission()}
		case models.Absorption:
			return &models.NeutronOutcome{Type: models.Absorption}
		default:
			// Neutron scattered, continues flying
			// In 1 dimensional slab: just continues in random direction
			// In 3 dimensional slab: would call sampleIsotropicDirection()

			continue
		}
	}
}
