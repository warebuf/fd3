package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// HandleUnitDestruction removes a unit from the field, clears its ability timers,
// clears any preselected targets pointing at it, and logs the destruction.
func HandleUnitDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardID int, attackerTeam int, executeAtWallTime int64) {
	card := &gs.Cards[cardID]
	cardTeam := card.Team

	// remove from field
	for i, ci := range gs.Field[cardTeam] {
		if ci == cardID {
			gs.Field[cardTeam] = append(gs.Field[cardTeam][:i], gs.Field[cardTeam][i+1:]...)
			break
		}
	}
	card.Zone = state.ZoneGraveyard
	gs.Grave[cardTeam] = append(gs.Grave[cardTeam], cardID)

	// cancel scheduled abilities
	pq.RemoveSimItemByCardID(cardID)
	card.AbilityTimers = []state.AbilityTimer{}

	// clear any preselected targets pointing at this card
	for i := range gs.Cards {
		for j := range gs.Cards[i].Actives {
			if gs.Cards[i].Actives[j].Target == cardID {
				gs.Cards[i].Actives[j].Target = -1
			}
		}
	}

	*logList = append(*logList, logs.DestroyCardLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.DestroyCard,
		TeamIndex:       attackerTeam,
		CardIndex:       cardID,
		TargetAll:       false,
		TargetTeamIndex: cardTeam,
		TargetCardIndex: cardID,
	})
	*UpdateID++
}
