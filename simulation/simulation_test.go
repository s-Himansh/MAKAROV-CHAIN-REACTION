package simulation

import (
	"math/rand/v2"
	"testing"

	loadingnucleardata "makarov-chains/loading_nuclear_data"
	"makarov-chains/models"
)

func loadTestLibrary(t *testing.T) *loadingnucleardata.NuclearDataLibrary {
	t.Helper()
	lib, err := loadingnucleardata.LoadNuclearData("../nuclear_data.json")
	if err != nil {
		t.Fatalf("Failed to load nuclear data: %v", err)
	}
	return lib
}

func TestNewSimulation(t *testing.T) {
	lib := loadTestLibrary(t)

	sim, err := New(lib, "U-235", 20.0)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if sim.Material == nil {
		t.Error("Material is nil")
	}
	if sim.Geometry.Thickness != 20.0 {
		t.Errorf("Thickness = %f, want 20.0", sim.Geometry.Thickness)
	}
	if sim.ΣTotal <= 0 {
		t.Errorf("ΣTotal = %f, want > 0", sim.ΣTotal)
	}
}

func TestNewSimulationInvalidIsotope(t *testing.T) {
	lib := loadTestLibrary(t)

	_, err := New(lib, "X-999", 20.0)
	if err == nil {
		t.Error("Expected error for invalid isotope, got nil")
	}
}

func TestSampleCollisionDistance(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)

	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 1000 {
		dist := sim.sampleCollisionDistance(rng)
		if dist <= 0 {
			t.Fatalf("iteration %d: distance = %f, want > 0", i, dist)
		}
	}
}

func TestSampleInteractionType(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)

	rng := rand.New(rand.NewPCG(42, 0))

	counts := map[models.InteractionType]int{}
	for range 10000 {
		it := sim.sampleInteractionType(rng)
		counts[it]++
	}

	if counts[models.Fission] == 0 {
		t.Error("No fission events sampled")
	}
	if counts[models.Absorption] == 0 {
		t.Error("No absorption events sampled")
	}
	if counts[models.Scatter] == 0 {
		t.Error("No scatter events sampled")
	}
}

func TestSampleNeutronEmission(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)

	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 1000 {
		n := sim.sampleNeutronEmission(rng)
		if n < 2 || n > 3 {
			t.Fatalf("iteration %d: neutrons = %d, want 2 or 3", i, n)
		}
	}
}

func TestSampleIsotropicDirection(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)

	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 1000 {
		dx, dy, dz := sim.sampleIsotropicDirection(rng)
		mag := dx*dx + dy*dy + dz*dz
		if mag < 0.99 || mag > 1.01 {
			t.Fatalf("iteration %d: magnitude = %f, want ~1.0", i, mag)
		}
	}
}

func TestTrackNeutron1D(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)
	sim.Geometry.Type = models.Slab

	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 100 {
		outcome := sim.trackNeutron1D(rng)
		if outcome == nil {
			t.Fatalf("iteration %d: nil outcome", i)
		}
	}
}

func TestTrackNeutron3D(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)
	sim.Geometry.Type = models.Sphere

	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 100 {
		outcome := sim.trackNeutron3D(rng)
		if outcome == nil {
			t.Fatalf("iteration %d: nil outcome", i)
		}
	}
}

func TestRunCompletes(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)
	sim.Params.NumParticles = 100
	sim.Params.NumGenerations = 10

	sim.Run()

	if sim.Tally.KEff <= 0 {
		t.Errorf("k_eff = %f, want > 0", sim.Tally.KEff)
	}
	if len(sim.Tally.NeutronsPerGen) == 0 {
		t.Error("No generation data recorded")
	}
}

func TestRunSphereCompletes(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)
	sim.Geometry.Type = models.Sphere
	sim.Params.NumParticles = 100
	sim.Params.NumGenerations = 10

	sim.Run()

	if sim.Tally.KEff <= 0 {
		t.Errorf("k_eff = %f, want > 0", sim.Tally.KEff)
	}
}

func TestRunQuiet(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)
	sim.Params.NumParticles = 50
	sim.Params.NumGenerations = 5
	sim.SetQuiet(true)

	sim.Run()

	if sim.Tally.KEff <= 0 {
		t.Errorf("k_eff = %f, want > 0", sim.Tally.KEff)
	}
}

func TestCalculateKeffSubcritical(t *testing.T) {
	sim := &Simulation{
		Params: &models.SimParams{NumParticles: 100},
		Tally: &models.Tally{
			NeutronsPerGen: []int{100, 50, 25, 12, 6},
		},
	}

	sim.calculateKeff()

	if sim.Tally.KEff >= 1.0 {
		t.Errorf("k_eff = %f, want < 1.0 for subcritical", sim.Tally.KEff)
	}
}

func TestCalculateKeffSupercritical(t *testing.T) {
	sim := &Simulation{
		Params: &models.SimParams{NumParticles: 100},
		Tally: &models.Tally{
			NeutronsPerGen: []int{100, 200, 400, 800, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000},
		},
	}

	sim.calculateKeff()

	if sim.Tally.KEff <= 1.0 {
		t.Errorf("k_eff = %f, want > 1.0 for supercritical", sim.Tally.KEff)
	}
}
