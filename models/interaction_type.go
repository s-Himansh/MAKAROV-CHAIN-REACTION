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

// NeutronOutcome holds the result of tracking one neutron
type NeutronOutcome struct {
	Type        InteractionType
	NewNeutrons int // Only non-zero for fission
}
