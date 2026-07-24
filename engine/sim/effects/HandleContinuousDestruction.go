package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// HandleContinuousDestruction removes a card from the continuous zone, cancels any
// scheduled items, clears any preselected targets pointing at it, and logs the removal.
func HandleContinuousDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	card := &gs.Cards[cardIndex]
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

	for i := range gs.Cards {
		for j := range gs.Cards[i].Actives {
			if gs.Cards[i].Actives[j].Target == cardIndex {
				gs.Cards[i].Actives[j].Target = -1
			}
		}
	}

	*logList = append(*logList, logs.RemoveContinuousLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.RemoveContinuous,
		TeamIndex:     cardTeam,
		CardIndex:     cardIndex,
	})
	*UpdateID++
}
