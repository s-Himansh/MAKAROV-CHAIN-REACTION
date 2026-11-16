package simulation

import (
	"fmt"
	"io"
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
	// OnGeneration, if set, is called after each generation. Run itself never prints.
	OnGeneration func(GenStats)

	ΣFission    float64
	ΣAbsorption float64
	ΣScatter    float64
	ΣTotal      float64
}

// GenStats summarises one generation for OnGeneration callbacks.
type GenStats struct {
	Generation  int
	NeutronsIn  int
	NeutronsOut int
	Fissions    int
	Absorptions int
	Escapes     int
	K           float64
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

	bank := make([]models.Site, s.Params.NumParticles)
	for i := range bank {
		bank[i] = s.sampleSource(rng)
	}

	for gen := range s.Params.NumGenerations {
		if len(bank) == 0 {
			break
		}

		var next []models.Site
		fissions := 0
		absorptions := 0
		escapes := 0

		for _, src := range bank {
			outcome := s.trackNeutron(rng, src)

			switch outcome.Type {
			case models.Fission:
				fissions++
				// Daughters are born where the parent fissioned.
				for range outcome.NewNeutrons {
					next = append(next, outcome.Site)
				}
				s.Tally.TotalFissions++
			case models.Absorption:
				absorptions++
				s.Tally.TotalAbsorptions++
			case models.Escape:
				escapes++
				s.Tally.TotalEscapes++
			}
		}

		s.Tally.NeutronsPerGen = append(s.Tally.NeutronsPerGen, len(next))

		k_gen := float64(len(next)) / float64(len(bank))

		if s.OnGeneration != nil {
			s.OnGeneration(GenStats{gen + 1, len(bank), len(next), fissions, absorptions, escapes, k_gen})
		}

		// Population control: keep a random NumParticles of the new sites.
		if len(next) > s.Params.NumParticles {
			rng.Shuffle(len(next), func(i, j int) { next[i], next[j] = next[j], next[i] })
			next = next[:s.Params.NumParticles]
		}
		bank = next
	}

	s.calculateKeff()
}

func (s *Simulation) WriteResults(w io.Writer) {
	fmt.Fprintln(w, "\n"+strings.Repeat("=", 50))
	fmt.Fprintln(w, "SIMULATION RESULTS")
	fmt.Fprintln(w, strings.Repeat("=", 50))
	fmt.Fprintf(w, "Material: %s\n", s.Material.Nuclide.Name)

	if s.Geometry.Type == models.Sphere {
		fmt.Fprintf(w, "Geometry: sphere, %.2f cm radius\n", s.Geometry.Thickness)
	} else {
		fmt.Fprintf(w, "Geometry: 1D slab, %.2f cm thick\n", s.Geometry.Thickness)
	}

	fmt.Fprintf(w, "\nTotal events:\n")
	fmt.Fprintf(w, "  Fissions:    %d\n", s.Tally.TotalFissions)
	fmt.Fprintf(w, "  Absorptions: %d\n", s.Tally.TotalAbsorptions)
	fmt.Fprintf(w, "  Escapes:     %d\n", s.Tally.TotalEscapes)
	fmt.Fprintf(w, "\nk_effective: %.4f\n", s.Tally.KEff)

	if s.Tally.KEff > 1.0 {
		fmt.Fprintln(w, "→ SUPERCRITICAL (growing reaction)")
	} else if s.Tally.KEff < 1.0 {
		fmt.Fprintln(w, "→ SUBCRITICAL (dying reaction)")
	} else {
		fmt.Fprintln(w, "→ CRITICAL (sustained reaction)")
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

// sampleSource returns a uniform random point inside the geometry.
func (s *Simulation) sampleSource(rng *rand.Rand) models.Site {
	if s.Geometry.Type != models.Sphere {
		return models.Site{X: s.Geometry.Thickness * rng.Float64()}
	}
	r := s.Geometry.Thickness * math.Cbrt(rng.Float64())
	dx, dy, dz := s.sampleIsotropicDirection(rng)
	return models.Site{X: r * dx, Y: r * dy, Z: r * dz}
}

func (s *Simulation) trackNeutron(rng *rand.Rand, src models.Site) *models.NeutronOutcome {
	if s.Geometry.Type == models.Sphere {
		return s.trackNeutron3D(rng, src)
	}
	return s.trackNeutron1D(rng, src)
}

func (s *Simulation) trackNeutron1D(rng *rand.Rand, src models.Site) *models.NeutronOutcome {
	position := src.X
	slabThickness := s.Geometry.Thickness

	for i := range maxCollisions {
		distance := s.sampleCollisionDistance(rng)

		// Isotropic scattering: direction cosine is uniform on [-1, 1].
		position += distance * (2*rng.Float64() - 1)

		if position < 0 || position > slabThickness {
			return &models.NeutronOutcome{Type: models.Escape}
		}

		interactionType := s.sampleInteractionType(rng)

		switch interactionType {
		case models.Fission:
			return &models.NeutronOutcome{Type: models.Fission, NewNeutrons: s.sampleNeutronEmission(rng), Site: models.Site{X: position}}
		case models.Absorption:
			return &models.NeutronOutcome{Type: models.Absorption}
		default:
			_ = i
			continue
		}
	}

	return &models.NeutronOutcome{Type: models.Absorption}
}

func (s *Simulation) trackNeutron3D(rng *rand.Rand, src models.Site) *models.NeutronOutcome {
	radius := s.Geometry.Thickness
	x, y, z := src.X, src.Y, src.Z

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
			return &models.NeutronOutcome{Type: models.Fission, NewNeutrons: s.sampleNeutronEmission(rng), Site: models.Site{X: x, Y: y, Z: z}}
		case models.Absorption:
			return &models.NeutronOutcome{Type: models.Absorption}
		default:
			_ = i
			continue
		}
	}

	return &models.NeutronOutcome{Type: models.Absorption}
}
