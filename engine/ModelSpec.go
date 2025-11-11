package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type modelSpec struct {
	Number    uint     `json:"Number"`
	Name      string   `json:"Name"`
	HP        int      `json:"HP"`
	Functions []string `json:"Functions"`
}

func LoadModelSpecs() []modelSpec {
	path := filepath.FromSlash("data/units.json")

	file, _ := os.Open(path)
	defer file.Close()

	var originals []modelSpec
	json.NewDecoder(file).Decode(&originals)

	fmt.Println("originals", originals)
	return originals
}

// Factory
func NewUnitFromID(originals []modelSpec, OGID int, playerID uint) {

	// if OGID is 0, do not create, as the 0 element is just a placeholder
	// if(OGID == 0) {return}

	// OG := originals[OGID]

}
