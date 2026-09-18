package autoability

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// when mana is added, collect the team's structures that can auto-draw, cheapest slot order
// costs accumulate across slots so the caller never pushes a draw the team can't afford
func AutoAbilityStructureDrawCheck(gs *state.GameState, teamIndex int) []int {
	if teamIndex < 0 || teamIndex >= len(gs.Structures) {
		return nil
	}

	var structureIndexes []int
	remaining := gs.Mana[teamIndex]
	for i := range gs.Structures[teamIndex] {
		structure := gs.Structures[teamIndex][i]
		if !structure.AutoAbility || structure.HP <= 0 {
			continue
		}

		drawCost := defs.StructureDefs[structure.DefNo].DrawCost
		if remaining < drawCost {
			continue
		}
		remaining -= drawCost
		structureIndexes = append(structureIndexes, i)
	}
	return structureIndexes
}
