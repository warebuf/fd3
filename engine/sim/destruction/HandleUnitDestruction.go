package destruction

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// removes a unit from the field, clears ability timers, clears preselects pointing at it
func HandleUnitDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	card := &gs.Cards[cardIndex]
	if card.Zone != state.ZoneField {
		return
	}

	if RedirectDestructionToEquip(gs, pq, logList, UpdateID, cardIndex, executeAtWallTime) {
		return
	}

	cardTeam := card.Team

	// remove from field
	for i, ci := range gs.Field[cardTeam] {
		if ci == cardIndex {
			gs.Field[cardTeam] = append(gs.Field[cardTeam][:i], gs.Field[cardTeam][i+1:]...)
			break
		}
	}
	card.Zone = state.ZoneGraveyard
	if !card.IsToken {
		gs.Grave[cardTeam] = append(gs.Grave[cardTeam], cardIndex)
	}

	// cancel scheduled abilities
	pq.RemoveSimItemByCardIndex(cardIndex)
	card.OnTick = [3]state.TickAbility{}

	ClearPreselectsTargeting(gs, cardIndex)
	RemoveChainItemsForCard(gs, logList, UpdateID, cardIndex, executeAtWallTime)
	DestroyEquipsForHost(gs, pq, logList, UpdateID, cardIndex, executeAtWallTime)

	*logList = append(*logList, logs.DestroyCardLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.DestroyCard,
		TargetAll:       false,
		TargetTeamIndex: cardTeam,
		TargetCardIndex: cardIndex,
	})
	*UpdateID++
}
