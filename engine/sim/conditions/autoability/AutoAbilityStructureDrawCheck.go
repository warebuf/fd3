package autoability

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// when mana is added, check if a team's structure can auto-draw
func AutoAbilityStructureDrawCheck(gs *state.GameState, teamIndex int) bool {
	if teamIndex < 0 || teamIndex >= len(gs.Structures) {
		return false
	}

	structure := gs.Structures[teamIndex]
	if !structure.AutoAbility {
		return false
	}

	structDef := defs.StructureDefs[structure.DefNo]
	return gs.Mana[teamIndex] >= structDef.DrawCost
}
