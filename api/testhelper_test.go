package api

import loadingnucleardata "makarov-chains/loading_nuclear_data"

func newTestLibrary() (*loadingnucleardata.NuclearDataLibrary, error) {
	return loadingnucleardata.LoadNuclearData("../nuclear_data.json")
}
