package simulate

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
	"github.com/warebuf/fd3/engine/functions"
)

func Simulate(state *engine.GameState, activatedAbilities []int) {

	var startingNodes []chainTypes.ChainNodes

	// go through activatedAbilities, add them to the startingNodes
	for i := 0; i < len(activatedAbilities); i++ {
		unitIndex := state.Abilities[activatedAbilities[i]].UnitIndex
		node := createStartingAbilityNode(state, unitIndex, activatedAbilities[i])
		if node != nil {
			startingNodes = append(startingNodes, node)
		}
	}

	// iterate through the state, find units that are ready to act
	for unitIndex := 0; unitIndex < len(state.HP); unitIndex++ {

		// check if they want to be summoned
		if state.HP[unitIndex] > 0 && state.Zone[unitIndex] == 3 { // presummon zone
			if state.SummonCharge[unitIndex] >= 2000 { // HARDCODED to 2 seconds
				startingNodes = append(startingNodes, &chainTypes.SummonDeclaration{
					SummonedIndex: unitIndex,
					Cancelled:     false, // always false at start
				})
			}
		}

		// check their abilities, if they can be activated
		funcStartIndex := state.AbilityStartIdx[unitIndex]
		for j := 0; j < len(state.AbilityCount); j++ {
			functionIndex := funcStartIndex + j
			abilityNo := state.Abilities[functionIndex].AbilityNo
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
	for i := 0; i < len(startingNodes); i++ {

	}

	// somehow return the results so we can communicate it with the frontend

}

func createStartingAbilityNode(state *engine.GameState, casterIndex int, abilityNo int) chainTypes.ChainNodes {
	switch abilityNo {
	case 0:
		return functions.A00000(state, casterIndex)
	case 1:
		return functions.A00001(state, casterIndex)
	case 2:
		return functions.A00002(state, casterIndex)
	case 3:
		return functions.A00003(state, casterIndex)
	case 4:
		return nil
	case 5:
		return functions.A00005(state, casterIndex)
	case 6:
		return functions.A00006(state, casterIndex)
	}
	return nil
}
