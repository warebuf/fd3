package simulate

import (
	"fmt"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
	"github.com/warebuf/fd3/engine/functions"
)

func Simulate(state *engine.GameState, activatedAbilities []int) {

	var startingNodes []chainTypes.ChainNodes
	fmt.Println(len(startingNodes))

	// go through active abilities, add them to the StartingEvents
	for i := 0; i < len(activatedAbilities); i++ {

		// given the abilityIndex, get the unitIndex
		TEMP := 0 // TODO: replace temp

		node := createStartingAbilityNode(state, TEMP, activatedAbilities[i])
		if node != nil {
			startingNodes = append(startingNodes, node)
		}

	}

	// iterate through the state, find units that are ready to act
	for unitIndex := 0; unitIndex < len(state.HP); unitIndex++ {

		// check if they want to be summoned
		if state.HP[unitIndex] > 0 && state.Zone[unitIndex] == 3 { // presummon zone
			if state.SummonCharge[unitIndex] >= 2000 { // HARDCODED to 2 seconds
				timeToAutoEvent = timeLeft
			}
		}

		// check their abilities, if they can be activated
		funcStartIndex := state.FunctionStartIdx[unitIndex]
		for j := 0; j < len(state.FunctionCount); j++ {
			functionIndex := funcStartIndex + j
			abilityNo := state.Functions[functionIndex]
			if ValidFunction(state, unitIndex, abilityNo) {
				node := createStartingAbilityNode(state, unitIndex, abilityNo)
				if node != nil {
					startingNodes = append(startingNodes, node)
				}
			}

		}

	}

	// sort the starting nodes by priority

	// process the starting nodes
	var nodes []chainTypes.ChainNodes
	fmt.Println(len(nodes))

	// somehow return the results so we can communicate it with the frontend

}

func createStartingAbilityNode(state *engine.GameState, casterIndex int, abilityNo int) chainTypes.ChainNodes {

	switch abilityNo {
	case 0:
		return functions.A00000(state, casterIndex)
	}

	return nil

}
