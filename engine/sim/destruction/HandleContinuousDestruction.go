package destruction

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// removes a continuous card, cancels any scheduled items, clears preselects pointing at it
func HandleContinuousDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	card := &gs.Cards[cardIndex]
	if card.Zone != state.ZoneContinuous {
		return
	}
	cardTeam := card.Team

	for i, ci := range gs.Continuous[cardTeam] {
		if ci == cardIndex {
			gs.Continuous[cardTeam] = append(gs.Continuous[cardTeam][:i], gs.Continuous[cardTeam][i+1:]...)
			break
		}
	}
	card.Zone = state.ZoneGraveyard
	gs.Grave[cardTeam] = append(gs.Grave[cardTeam], cardIndex)

	pq.RemoveSimItemByCardIndex(cardIndex)

	ClearPreselectsTargeting(gs, cardIndex)
	RemoveChainItemsForCard(gs, logList, UpdateID, cardIndex, executeAtWallTime)

	*logList = append(*logList, logs.RemoveContinuousLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.RemoveContinuous,
		TeamIndex:     cardTeam,
		CardIndex:     cardIndex,
	})
	*UpdateID++
}
