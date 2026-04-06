package models

// defines the physical properties of a single nuclear species
type Nuclide struct {
	Name                   string  // nuclide's name
	FissionCrossSection    float64 // a neutron causing fission (in barns)
	AbsorptionCrossSection float64 // a neutron is absorbed (in barns)
	ScatterCrossSection    float64 // a neutron got deviated (in barns)
	Nu                     float64 // no. of neutrons generated per fission event
	MolarMass              float64 // molar mass (g/mol)
}
