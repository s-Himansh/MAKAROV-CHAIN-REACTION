package models

// accumulates statistics across the chain
type Tally struct {
	NeutronsPerGen   []int
	TotalFissions    int
	TotalAbsorptions int
	TotalEscapes     int
	KEff             float64

	// k_eff = (neutrons in gen N+1) / (neutrons in gen N)
	// k < 1 → subcritical (reaction dies out)
	// k = 1 → critical (sustained chain reaction)
	// k > 1 → supercritical (growing reaction)
}
