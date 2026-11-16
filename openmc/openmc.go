// Package openmc is a small Go client for OpenMC (https://github.com/openmc-dev/openmc).
// It writes the XML inputs for a bare-sphere eigenvalue problem, runs the
// openmc executable and reads back k-effective, so Go programs can drive
// OpenMC without the Python API.
//
// Scope: one homogeneous sphere in vacuum, continuous-energy data. The inputs
// follow OpenMC's documented XML formats; OpenMC itself must be installed and
// OPENMC_CROSS_SECTIONS must point at a cross-section library.
package openmc

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

// Nuclide is one component of the material, by OpenMC name (for example "U235")
// and atom fraction.
type Nuclide struct {
	Name string
	AO   float64
}

// Model describes a homogeneous sphere with a vacuum boundary.
type Model struct {
	Nuclides                     []Nuclide
	Density                      float64 // g/cm3
	Radius                       float64 // cm
	Particles, Batches, Inactive int
	Seed                         int64 // 0 keeps OpenMC's default
}

// Result is the combined k-effective estimate from an OpenMC run.
type Result struct {
	KEff, StdDev float64
	Output       string // OpenMC's standard output
}

func (m Model) validate() error {
	switch {
	case len(m.Nuclides) == 0:
		return errors.New("openmc: no nuclides")
	case m.Density <= 0 || m.Radius <= 0:
		return errors.New("openmc: density and radius must be positive")
	case m.Particles <= 0 || m.Batches <= 0 || m.Inactive < 0 || m.Inactive >= m.Batches:
		return errors.New("openmc: need particles > 0 and 0 <= inactive < batches")
	}
	return nil
}

type (
	xmlDensity struct {
		Value float64 `xml:"value,attr"`
		Units string  `xml:"units,attr"`
	}
	xmlNuclide struct {
		Name string  `xml:"name,attr"`
		AO   float64 `xml:"ao,attr"`
	}
	xmlMaterial struct {
		ID       int          `xml:"id,attr"`
		Density  xmlDensity   `xml:"density"`
		Nuclides []xmlNuclide `xml:"nuclide"`
	}
	xmlMaterials struct {
		XMLName  xml.Name      `xml:"materials"`
		Material []xmlMaterial `xml:"material"`
	}
	xmlSurface struct {
		ID       int    `xml:"id,attr"`
		Type     string `xml:"type,attr"`
		Coeffs   string `xml:"coeffs,attr"`
		Boundary string `xml:"boundary,attr"`
	}
	xmlCell struct {
		ID       int    `xml:"id,attr"`
		Material int    `xml:"material,attr"`
		Region   string `xml:"region,attr"`
	}
	xmlGeometry struct {
		XMLName xml.Name     `xml:"geometry"`
		Cell    []xmlCell    `xml:"cell"`
		Surface []xmlSurface `xml:"surface"`
	}
	xmlSpace struct {
		Type       string `xml:"type,attr"`
		Parameters string `xml:"parameters,attr"`
	}
	xmlSource struct {
		Space xmlSpace `xml:"space"`
	}
	xmlSettings struct {
		XMLName   xml.Name  `xml:"settings"`
		RunMode   string    `xml:"run_mode"`
		Particles int       `xml:"particles"`
		Batches   int       `xml:"batches"`
		Inactive  int       `xml:"inactive"`
		Seed      int64     `xml:"seed,omitempty"`
		Source    xmlSource `xml:"source"`
	}
)

// WriteXML writes materials.xml, geometry.xml and settings.xml into dir.
func (m Model) WriteXML(dir string) error {
	if err := m.validate(); err != nil {
		return err
	}
	mat := xmlMaterial{ID: 1, Density: xmlDensity{m.Density, "g/cm3"}}
	for _, n := range m.Nuclides {
		mat.Nuclides = append(mat.Nuclides, xmlNuclide(n))
	}
	files := map[string]any{
		"materials.xml": xmlMaterials{Material: []xmlMaterial{mat}},
		"geometry.xml": xmlGeometry{
			Cell:    []xmlCell{{ID: 1, Material: 1, Region: "-1"}},
			Surface: []xmlSurface{{ID: 1, Type: "sphere", Coeffs: fmt.Sprintf("0 0 0 %g", m.Radius), Boundary: "vacuum"}},
		},
		"settings.xml": xmlSettings{
			RunMode: "eigenvalue", Particles: m.Particles, Batches: m.Batches, Inactive: m.Inactive, Seed: m.Seed,
			Source: xmlSource{xmlSpace{Type: "point", Parameters: "0 0 0"}},
		},
	}
	for name, v := range files {
		b, err := xml.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), append([]byte(xml.Header), b...), 0o644); err != nil {
			return err
		}
	}
	return nil
}

var kEffLine = regexp.MustCompile(`Combined k-effective\s*=\s*(\S+)\s*\+/-\s*(\S+)`)

// ParseKEff extracts the combined k-effective and its standard deviation from
// OpenMC's standard output.
func ParseKEff(output string) (k, sd float64, err error) {
	m := kEffLine.FindStringSubmatch(output)
	if m == nil {
		return 0, 0, errors.New("openmc: no 'Combined k-effective' line in output")
	}
	if k, err = strconv.ParseFloat(m[1], 64); err != nil {
		return 0, 0, err
	}
	sd, err = strconv.ParseFloat(m[2], 64)
	return k, sd, err
}

// Run writes the inputs into dir, runs binary (usually "openmc") there and
// parses k-effective from its output.
func (m Model) Run(ctx context.Context, binary, dir string) (Result, error) {
	if err := m.WriteXML(dir); err != nil {
		return Result{}, err
	}
	cmd := exec.CommandContext(ctx, binary)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Output: string(out)}, fmt.Errorf("openmc: %w\n%s", err, out)
	}
	k, sd, err := ParseKEff(string(out))
	return Result{KEff: k, StdDev: sd, Output: string(out)}, err
}
