package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectCounterAndDestroy(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, sourceCardIndex int, executeAtWallTime int64) {
	if len(gs.ReactionStack) == 0 {
		return
	}

	target := gs.ReactionStack[len(gs.ReactionStack)-1]
	gs.ReactionStack = gs.ReactionStack[:len(gs.ReactionStack)-1]

	switch target.Type {
	case state.ReactionSpell, state.ReactionSummon, state.ReactionContinuousResolution:
		card := &gs.Cards[target.CardIndex]
		card.Zone = state.ZoneGraveyard
		gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardIndex)
	case state.ReactionContinuousAbility:
		destruction.HandleContinuousDestruction(gs, pq, logList, UpdateID, target.CardIndex, executeAtWallTime)
	case state.ReactionUnitAbility:
		destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, target.CardIndex, executeAtWallTime)
	case state.ReactionStructureAbility:
		// no source card, CardIndex is -1
	}

	*logList = append(*logList, logs.CounterStackLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.CounterStack,
		SourceCardIndex: sourceCardIndex,
		TargetCardIndex: target.CardIndex,
	})
	*UpdateID++
}
