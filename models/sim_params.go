package models

// controls the Monte Carlo engine
type SimParams struct {
	NumParticles   int   // neutrons per generation
	NumGenerations int   // drives the Markov chain iteration
	Seed           int64 // makes runs reproducible.
}
