package chaining

import (
	"fmt"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
)

func processHPChangeDeclaration(state engine.GameState, node *chainTypes.HPChangeDeclaration) {

	// iterate through the units, looking for triggers
	// TODO: not sure what order to process the triggers

	for i := 0; i < len(state.HP); i++ {
		if state.HP[i] > 0 && state.Zone[i] == 1 { // alive and on-field
			startIndex := state.AbilityStartIdx[i]
			for j := 0; j < state.AbilityCount[i]; j++ {
				functionIndex := startIndex + j
				fmt.Println("functionIndex", functionIndex)
				// check if function can react
			}
		}
	}

}
