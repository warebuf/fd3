package engine

import (
	"os"
	"path/filepath"
)

type Model struct {
	UnitNo    uint
	Name      string
	HP        int
	Abilities []AbilityDef
}

type AbilityDef struct {
	AbilityNo   int
	Type        string
	CD          int
	Description string
}

func LoadModelSpecs() []Model {
	path := filepath.FromSlash("data/units.csv")

	file, _ := os.Open(path)
	defer file.Close()

	// TODO: iterate through the rows
	// create a map
	// edit to get the final objects

	return
}
