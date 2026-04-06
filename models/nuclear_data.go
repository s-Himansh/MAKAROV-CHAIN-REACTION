package models

type IsotopeData struct {
	Name             string  `json:"name"`
	MolarMass        float64 `json:"molar_mass"`
	XSFission        float64 `json:"xs_fission"`
	XSCapture        float64 `json:"xs_capture"`
	XSElastic        float64 `json:"xs_elastic"`
	XSTotal          float64 `json:"xs_total"`
	Nu               float64 `json:"nu"`
	Description      string  `json:"description"`
	NaturalAbundance float64 `json:"natural_abundance"`
	DensityMetal     float64 `json:"density_metal"`
}

type NuclearDataFile struct {
	Metadata *NuclearMetaData        `json:"metadata"`
	Isotopes map[string]*IsotopeData `json:"isotopes"`
}

type NuclearMetaData struct {
	Description string  `json:"description"`
	Source      string  `json:"source"`
	EnergyEV    float64 `json:"energy_eV"`
	Units       *Units  `json:"units"`
}

type Units struct {
	CrossSection string `json:"cross_section"`
	MolarMass    string `json:"molar_mass"`
	Nu           string `json:"nu"`
}
