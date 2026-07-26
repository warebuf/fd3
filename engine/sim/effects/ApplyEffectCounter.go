package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectCounter(gs *state.GameState, logList *[]any, UpdateID *int, sourceCardIndex int, executeAtWallTime int64) {
	if len(gs.ReactionStack) == 0 {
		return
	}

	target := gs.ReactionStack[len(gs.ReactionStack)-1]
	gs.ReactionStack = gs.ReactionStack[:len(gs.ReactionStack)-1]

	if target.Type != state.ReactionStructureAbility {
		card := &gs.Cards[target.CardIndex]
		switch target.Type {
		case state.ReactionSpell, state.ReactionSummon, state.ReactionContinuousResolution:
			card.Zone = state.ZoneGraveyard
			gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardIndex)
		case state.ReactionContinuousAbility:
			for i, id := range gs.Continuous[card.Team] {
				if id == target.CardIndex {
					gs.Continuous[card.Team] = append(gs.Continuous[card.Team][:i], gs.Continuous[card.Team][i+1:]...)
					break
				}
			}
			card.Zone = state.ZoneGraveyard
			gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardIndex)
		case state.ReactionUnitAbility:
			if target.ActiveIndex >= 0 {
				card.PendingCharges--
			}
		}
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
