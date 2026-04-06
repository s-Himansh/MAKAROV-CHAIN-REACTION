package models

// wraps a Nuclide with macroscopic, bulk properties
type Material struct {
	Nuclide       Nuclide
	Density       float64 // g/cm³
	AtomicDensity float64 // atoms/cm³
}

func (m *Material) CalcAtomicDensity() {
	// n = (density × Avogadro) / MolarMass
	// Use molar mass from the nuclide
	m.AtomicDensity = (m.Density * AvogadroNumber) / m.Nuclide.MolarMass
}

func (m *Material) MacroScopicCrossSection(microScopicXS float64) float64 {
	// Macroscopic cross section Σ = n × σ
	// 1 barn = 1e-24 cm²

	barn := 1e-24

	return m.AtomicDensity * microScopicXS * barn
}

func (m *Material) TotalCrossSections() float64 {
	// per nuclide, this method retrieves total cross sections
	return m.MacroScopicCrossSection(m.Nuclide.FissionCrossSection + m.Nuclide.AbsorptionCrossSection + m.Nuclide.ScatterCrossSection)
}
