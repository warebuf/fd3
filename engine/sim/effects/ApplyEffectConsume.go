package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectConsume(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team

	targetIndex := eff.A
	valid := false
	for _, ci := range gs.Field[teamIndex] {
		if ci == targetIndex {
			valid = true
			break
		}
	}
	if !valid {
		for _, ci := range gs.Continuous[teamIndex] {
			if ci == targetIndex {
				valid = true
				break
			}
		}
	}
	if !valid {
		return // preselected target is gone or was never set, ability fizzles
	}

	charges := gs.Cards[targetIndex].Charges

	// C: which of this card's Active slots receives the Bonus, set explicitly by the CardDef
	gs.Cards[sourceCardIndex].Actives[eff.C].Bonus = charges
	gs.Cards[sourceCardIndex].HP += charges
	*logList = append(*logList, logs.HealLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.Heal,
		HealerIndex:     sourceCardIndex,
		TargetCardIndex: sourceCardIndex,
		Amount:          charges,
	})
	*UpdateID++

	if gs.Cards[targetIndex].Type == state.CardTypeUnit {
		HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, teamIndex, executeAtWallTime)
	} else if gs.Cards[targetIndex].Type == state.CardTypeContinuous {
		HandleContinuousDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
	}
}
