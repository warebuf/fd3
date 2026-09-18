package schedule

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleStructureManaRegen(gs *state.GameState, pq *state.SimulationPQ, teamIndex int, structureIndex int, startAtSimTime int64) {

	structure := gs.Structures[teamIndex][structureIndex]
	if structure.HP <= 0 {
		return
	}

	structDef := defs.StructureDefs[structure.DefNo]
	if structDef.ManaRegen == 0 || structDef.ManaRegenInterval == 0 {
		return
	}
	scheduledTime := startAtSimTime + structDef.ManaRegenInterval/constants.SIM_SPEED - constants.LEAD
	heap.Push(pq, &state.SimulationItem{
		StartAtSimTime:   startAtSimTime,
		PopAtSimTime:     scheduledTime,
		ExecuteAtSimTime: scheduledTime + constants.LEAD,
		AbilityIndex:     -1,
		CardIndex:        -1,
		TeamIndex:        teamIndex,
		StructureIndex:   structureIndex,
		Type:             state.SimItemStructureRegen,
		Effects:          []state.Effect{{Type: state.EffectMana, A: teamIndex, B: structDef.ManaRegen}},
	})

}
