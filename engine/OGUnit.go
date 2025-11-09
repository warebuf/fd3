package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type OGUnit struct {
	Name      string   `json:"Name"`
	HP        int      `json:"HP"`
	ATK       int      `json:"ATK"`
	SPD       int      `json:"SPD"`
	Functions []string `json:"Functions"`
}

func LoadUnitData() []OGUnit {
	path := filepath.FromSlash("data/units.json")

	file, _ := os.Open(path)
	defer file.Close()

	var originals []OGUnit
	json.NewDecoder(file).Decode(&originals)

	fmt.Println("originals", originals)
	return originals
}

// Factory
func NewUnitFromID(originals []OGUnit, OGID int, playerID uint) {

	// if OGID is 0, do not create, as the 0 element is just a placeholder
	// if(OGID == 0) {return}

	// OG := originals[OGID]

	return
}
