package api

import loadingnucleardata "github.com/s-Himansh/MAKAROV-CHAIN-REACTION/loading_nuclear_data"

func newTestLibrary() (*loadingnucleardata.NuclearDataLibrary, error) {
	return loadingnucleardata.LoadNuclearData("../nuclear_data.json")
}
