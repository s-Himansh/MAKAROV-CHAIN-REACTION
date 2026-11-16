package models

// controls the Monte Carlo engine
type SimParams struct {
	NumParticles   int   // neutrons per generation
	NumGenerations int   // drives the Markov chain iteration
	Seed           int64 // makes runs reproducible.
	Inactive       int   // generations discarded before averaging k; 0 = automatic: about (optical thickness)^2, at least 10, at most half the run
}
