// Package simulation is a one-group Monte Carlo neutron transport engine for
// criticality estimates in a slab or sphere of a single isotope.
//
// Build a Simulation with New, adjust Params and Geometry, then call Run.
// Run is silent; set OnGeneration to observe progress. Results are in Tally.
// A fixed Params.Seed makes a run reproducible.
package simulation
