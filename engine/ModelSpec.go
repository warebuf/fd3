package engine

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"fmt"
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

func LoadModels() []Model {
	path := filepath.FromSlash("data/units.csv")

	file, _ := os.Open(path)
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read() // skip the header row

	var models []Model
	abilities := make(map[int][]AbilityDef) // unitNo to Abilities arr

	// iterate through the rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		// parse cells
		unitNo, _ := strconv.Atoi(record[0])
		name := record[1]
		HP, _ := strconv.Atoi(record[2])
		abilityNo, _ := strconv.Atoi(record[3])
		abilityType := record[4]
		cd := 0
		if record[5] != "-" {
			cd, _ = strconv.Atoi(record[5])
		}
		description := record[6]

		// add abilities to abilities arr
		ability := AbilityDef{
			AbilityNo:   abilityNo,
			Type:        abilityType,
			CD:          cd,
			Description: description,
		}
		abilities[unitNo] = append(abilities[unitNo], ability)

		// add the unit to the arr
		if unitNo > len(models) { // if does not exist, assumes we don't skip UnitNos / they are sequential
			models = append(models, Model{
				UnitNo:    uint(unitNo),
				Name:      name,
				HP:        HP,
				Abilities: []AbilityDef{},
			})
		}

	}

	// edit to get the final objects
	for i := 0; i < len(models); i++ {
		models[i].Abilities = abilities[i+1] // unitNo starts at 1 not 0
	}

	fmt.Println("models", models)

	return models
}
