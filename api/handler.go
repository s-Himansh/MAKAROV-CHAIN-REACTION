package api

import (
	"encoding/json"
	"fmt"
	"github.com/s-Himansh/MAKAROV-CHAIN-REACTION/models"
	"github.com/s-Himansh/MAKAROV-CHAIN-REACTION/simulation"
	"net/http"
	"strings"

	loadingnucleardata "github.com/s-Himansh/MAKAROV-CHAIN-REACTION/loading_nuclear_data"
)

type Handler struct {
	library *loadingnucleardata.NuclearDataLibrary
}

type SimRequest struct {
	Isotope     string  `json:"isotope"`
	Thickness   float64 `json:"thickness"`
	Geometry    string  `json:"geometry"`
	Particles   int     `json:"particles"`
	Generations int     `json:"generations"`
	Seed        int64   `json:"seed"`
}

type SimResponse struct {
	Material         string  `json:"material"`
	GeometryType     string  `json:"geometry"`
	Thickness        float64 `json:"thickness"`
	NumParticles     int     `json:"num_particles"`
	NumGenerations   int     `json:"num_generations"`
	Seed             int64   `json:"seed"`
	TotalFissions    int     `json:"total_fissions"`
	TotalAbsorptions int     `json:"total_absorptions"`
	TotalEscapes     int     `json:"total_escapes"`
	KEff             float64 `json:"k_effective"`
	Status           string  `json:"status"`
	NeutronsPerGen   []int   `json:"neutrons_per_gen"`
	MFP              float64 `json:"mean_free_path"`
	SigmaTotal       float64 `json:"sigma_total"`
}

type IsotopeInfo struct {
	Name         string  `json:"name"`
	FissionXS    float64 `json:"fission_xs"`
	AbsorptionXS float64 `json:"absorption_xs"`
	ScatterXS    float64 `json:"scatter_xs"`
	Nu           float64 `json:"nu"`
	Density      float64 `json:"density"`
	Description  string  `json:"description"`
}

type IsotopesResponse struct {
	Isotopes []IsotopeInfo `json:"isotopes"`
}

func NewHandler() (*Handler, error) {
	library, err := loadingnucleardata.LoadNuclearData("nuclear_data.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load nuclear data: %w", err)
	}
	return &Handler{library: library}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/simulate", h.handleSimulate)
	mux.HandleFunc("GET /api/isotopes", h.handleIsotopes)
	mux.HandleFunc("GET /api/health", h.handleHealth)
	mux.HandleFunc("OPTIONS /api/simulate", h.handleCORS)
	mux.HandleFunc("OPTIONS /api/isotopes", h.handleCORS)
}

// Limits keep one request from tying up the server.
const (
	maxBodyBytes   = 1 << 20
	maxParticles   = 100_000
	maxGenerations = 1_000
	maxThickness   = 1_000.0
)

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// handleCORS answers preflight requests only; real handlers call setCORS so
// they can still set Content-Type and a status after it.
func (h *Handler) handleCORS(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	setCORS(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleIsotopes(w http.ResponseWriter, r *http.Request) {
	setCORS(w)

	isotopes := []string{"U-235", "U-238", "Pu-239", "Pu-240", "Pu-241", "H-1", "C-12", "O-16", "B-10", "Cd-113"}

	var result []IsotopeInfo
	for _, name := range isotopes {
		iso, err := h.library.RetrieveIsotope(name)
		if err != nil {
			continue
		}
		result = append(result, IsotopeInfo{
			Name:         name,
			FissionXS:    iso.XSFission,
			AbsorptionXS: iso.XSCapture,
			ScatterXS:    iso.XSElastic,
			Nu:           iso.Nu,
			Density:      iso.DensityMetal,
			Description:  iso.Description,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IsotopesResponse{Isotopes: result})
}

func (h *Handler) handleSimulate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var req SimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Isotope == "" {
		req.Isotope = "U-235"
	}
	if req.Thickness <= 0 {
		req.Thickness = 20.0
	}
	if req.Particles <= 0 {
		req.Particles = 1000
	}
	if req.Generations <= 0 {
		req.Generations = 50
	}
	if req.Geometry == "" {
		req.Geometry = "slab"
	}

	geometry := strings.ToLower(req.Geometry)
	switch {
	case geometry != "slab" && geometry != "sphere":
		writeError(w, http.StatusBadRequest, "geometry must be slab or sphere")
		return
	case req.Particles > maxParticles, req.Generations > maxGenerations, req.Thickness > maxThickness:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("limits: particles <= %d, generations <= %d, thickness <= %g", maxParticles, maxGenerations, maxThickness))
		return
	}

	sim, err := simulation.New(h.library, req.Isotope, req.Thickness)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sim.Params.NumParticles = req.Particles
	sim.Params.NumGenerations = req.Generations
	if req.Seed != 0 {
		sim.Params.Seed = req.Seed
	}

	if geometry == "sphere" {
		sim.Geometry.Type = models.Sphere
	}

	sim.Run()

	status := "critical"
	switch {
	case sim.Tally.KEff > 1.0:
		status = "supercritical"
	case sim.Tally.KEff < 1.0:
		status = "subcritical"
	}

	resp := SimResponse{
		Material:         sim.Material.Nuclide.Name,
		GeometryType:     req.Geometry,
		Thickness:        req.Thickness,
		NumParticles:     req.Particles,
		NumGenerations:   req.Generations,
		Seed:             sim.Params.Seed,
		TotalFissions:    sim.Tally.TotalFissions,
		TotalAbsorptions: sim.Tally.TotalAbsorptions,
		TotalEscapes:     sim.Tally.TotalEscapes,
		KEff:             sim.Tally.KEff,
		Status:           status,
		NeutronsPerGen:   sim.Tally.NeutronsPerGen,
		MFP:              1.0 / sim.ΣTotal,
		SigmaTotal:       sim.ΣTotal,
	}

	setCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
