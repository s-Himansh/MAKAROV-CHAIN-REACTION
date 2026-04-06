package loadingnucleardata

import (
	"encoding/json"
	"fmt"
	"makarov-chains/models"
	"os"
)

type NuclearDataLibrary struct {
	data     *models.NuclearDataFile
	isotopes map[string]*models.IsotopeData
}

// LoadNuclearData is base function that actually loads the nuclear data provided from nuclear_data.json. Later will be stored in database
func LoadNuclearData(path string) (*NuclearDataLibrary, error) {
	nuclearData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read nuclear data file: %w", err)
	}

	data := &models.NuclearDataFile{}

	err = json.Unmarshal(nuclearData, data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nuclear data JSON: %w", err)
	}

	return &NuclearDataLibrary{data, data.Isotopes}, nil
}

func (ndl *NuclearDataLibrary) CreateNuclide(isotopeName string) (*models.Nuclide, error) {
	isotope, err := ndl.RetrieveIsotope(isotopeName)
	if err != nil {
		return &models.Nuclide{}, err
	}

	return &models.Nuclide{Name: isotopeName, FissionCrossSection: isotope.XSFission, AbsorptionCrossSection: isotope.XSCapture,
		ScatterCrossSection: isotope.XSElastic, Nu: isotope.Nu, MolarMass: isotope.MolarMass}, nil

}

func (ndl *NuclearDataLibrary) CreateMaterial(isotopeName string, density float64) (*models.Material, error) {
	nuclide, err := ndl.CreateNuclide(isotopeName)
	if err != nil {
		return &models.Material{}, err
	}

	material := &models.Material{Nuclide: *nuclide, Density: density}

	material.CalcAtomicDensity()

	return material, err
}

func (ndl *NuclearDataLibrary) CreateMetalMaterial(isotopeName string) (*models.Material, error) {
	isotope, err := ndl.RetrieveIsotope(isotopeName)
	if err != nil {
		return &models.Material{}, err
	}

	if isotope.DensityMetal == 0 {
		return &models.Material{}, fmt.Errorf("no metal density data for %s", isotopeName)
	}

	return ndl.CreateMaterial(isotopeName, isotope.DensityMetal)
}

// these are helper functions

func (ndl *NuclearDataLibrary) RetrieveIsotope(name string) (*models.IsotopeData, error) {
	isotope, ok := ndl.isotopes[name]
	if !ok {
		return &models.IsotopeData{}, fmt.Errorf("isotope %s not found in nuclear data library", name)
	}

	return isotope, nil
}

func (ndl *NuclearDataLibrary) ListIsotopeNames() []string {
	names := make([]string, 0, len(ndl.isotopes))

	for name := range ndl.isotopes {
		names = append(names, name)
	}

	return names
}

func (ndl *NuclearDataLibrary) RetrieveMetaData() string {
	return fmt.Sprintf("%s (%s) at %.4f eV", ndl.data.Metadata.Description, ndl.data.Metadata.Source, ndl.data.Metadata.EnergyEV)
}
