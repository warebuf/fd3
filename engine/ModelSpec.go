package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Model struct {
	Number    uint      `json:"Number"`
	Name      string    `json:"Name"`
	HP        int       `json:"HP"`
	Abilities []Ability `json:"Abilities"`
}

type Ability struct {
	AbilityID   int    `json:"AbilityID"`
	Name        string `json:"Name"`
	CD          int64  `json:"CD"`
	Description string `json:"Description"`
}

func LoadModelSpecs() []Model {
	path := filepath.FromSlash("data/units.json")

	file, _ := os.Open(path)
	defer file.Close()

	var originals []Model
	json.NewDecoder(file).Decode(&originals)

	return originals
}
