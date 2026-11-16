package simulation

import (
	"math"
	"math/rand/v2"
	"testing"

	loadingnucleardata "github.com/s-Himansh/MAKAROV-CHAIN-REACTION/loading_nuclear_data"
	"github.com/s-Himansh/MAKAROV-CHAIN-REACTION/models"
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
		outcome := sim.trackNeutron1D(rng, sim.sampleSource(rng))
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
		outcome := sim.trackNeutron3D(rng, sim.sampleSource(rng))
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

func TestOnGeneration(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 20.0)
	sim.Params.NumParticles = 50
	sim.Params.NumGenerations = 5
	calls := 0
	sim.OnGeneration = func(g GenStats) {
		calls++
		if g.Generation != calls || g.NeutronsIn <= 0 {
			t.Errorf("bad stats %+v", g)
		}
	}

	sim.Run()

	if calls != len(sim.Tally.NeutronsPerGen) || calls == 0 {
		t.Errorf("callback calls = %d, generations = %d", calls, len(sim.Tally.NeutronsPerGen))
	}
}

func TestCalculateKeffSubcritical(t *testing.T) {
	sim := &Simulation{
		Params: &models.SimParams{NumParticles: 100},
		Tally: &models.Tally{
			NeutronsPerGen: []int{80, 80, 80, 80, 80},
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

func TestSampleSourceInsideGeometry(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 10.0)
	rng := rand.New(rand.NewPCG(1, 0))

	for range 1000 {
		if p := sim.sampleSource(rng); p.X < 0 || p.X > 10 {
			t.Fatalf("slab site %+v outside [0,10]", p)
		}
	}
	sim.Geometry.Type = models.Sphere
	for range 1000 {
		p := sim.sampleSource(rng)
		if r := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z); r > 10 {
			t.Fatalf("sphere site %+v has r=%f > 10", p, r)
		}
	}
}

func TestRunReproducible(t *testing.T) {
	lib := loadTestLibrary(t)
	run := func() []int {
		sim, _ := New(lib, "U-235", 5.0)
		sim.Params.NumParticles, sim.Params.NumGenerations, sim.Params.Seed = 200, 8, 7
		sim.Run()
		return sim.Tally.NeutronsPerGen
	}
	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("gen %d differs: %d vs %d", i, a[i], b[i])
		}
	}
}

// A large sphere has negligible leakage, so k_eff must approach the analytic
// infinite-medium value nu*Sigma_f / (Sigma_f + Sigma_capture).
func TestKeffApproachesKInfinity(t *testing.T) {
	lib := loadTestLibrary(t)
	sim, _ := New(lib, "U-235", 5.0)
	sim.Geometry.Type = models.Sphere
	sim.Params.NumParticles, sim.Params.NumGenerations, sim.Params.Seed = 2000, 40, 11

	sim.Run()

	kInf := sim.Material.Nuclide.Nu * sim.ΣFission / (sim.ΣFission + sim.ΣAbsorption)
	if got := sim.Tally.KEff; math.Abs(got-kInf)/kInf > 0.02 {
		t.Errorf("k_eff = %.4f, want within 2%% of k_inf = %.4f", got, kInf)
	}
}

func TestInactiveSkipsGenerations(t *testing.T) {
	sim := &Simulation{
		Params: &models.SimParams{NumParticles: 100, Inactive: 3},
		Tally:  &models.Tally{NeutronsPerGen: []int{100, 100, 100, 200, 200, 200}},
	}

	sim.calculateKeff()

	// Generations 4-6 each have k = 200/100 = 2 once the first three are skipped.
	if sim.Tally.KEff != 2.0 {
		t.Errorf("k_eff = %f, want 2.0", sim.Tally.KEff)
	}
}

func TestAutoInactiveScalesWithOpticalThickness(t *testing.T) {
	lib := loadTestLibrary(t)
	thin, _ := New(lib, "U-235", 0.05)
	thick, _ := New(lib, "U-235", 0.3)

	if got := thin.inactiveGenerations(400); got != 10 {
		t.Errorf("thin: got %d, want 10", got)
	}
	if got := thick.inactiveGenerations(400); got < 90 || got > 120 {
		t.Errorf("thick: got %d, want about 104", got)
	}
	if got := thick.inactiveGenerations(50); got != 25 {
		t.Errorf("short run: got %d, want half (25)", got)
	}
}
