package models

type InteractionType int

// this works as a auto-increment counter with Fission = 0, Absorption = 1 and so on
const (
	Fission InteractionType = iota
	Absorption
	Scatter
	Escape
)

// String
func (it InteractionType) String() string {
	switch it {
	case Fission:
		return "FISSION"
	case Absorption:
		return "ABSORPTION"
	case Scatter:
		return "SCATTER"
	case Escape:
		return "ESCAPE"
	default:
		return "UNKNOWN"
	}
}

// Site is a point in space where a neutron is born or fissions.
type Site struct{ X, Y, Z float64 }

// NeutronOutcome holds the result of tracking one neutron
type NeutronOutcome struct {
	Type        InteractionType
	NewNeutrons int  // Only non-zero for fission
	Site        Site // Where the fission happened
}
