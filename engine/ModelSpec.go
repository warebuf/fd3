package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Model struct {
	Number    uint       `json:"Number"`
	Name      string     `json:"Name"`
	HP        int        `json:"HP"`
	Functions []Function `json:"Functions"`
}

type Function struct {
	FunctionID  int    `json:"FunctionID"`
	Type        string `json:"Type"`
	Target      string `json:"Target"`
	Name        string `json:"Name"`
	CD          int    `json:"CD"`
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
