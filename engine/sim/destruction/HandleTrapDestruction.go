package destruction

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// removes a set trap, cancels the timer still counting it down, clears preselects pointing at it
func HandleTrapDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	card := &gs.Cards[cardIndex]
	if card.Zone != state.ZoneTrap {
		return
	}

	if RedirectDestructionToEquip(gs, pq, logList, UpdateID, cardIndex, executeAtWallTime) {
		return
	}

	cardTeam := card.Team

	for i, ci := range gs.Traps[cardTeam] {
		if ci == cardIndex {
			gs.Traps[cardTeam] = append(gs.Traps[cardTeam][:i], gs.Traps[cardTeam][i+1:]...)
			break
		}
	}
	card.Zone = state.ZoneGraveyard
	if !card.IsToken {
		gs.Grave[cardTeam] = append(gs.Grave[cardTeam], cardIndex)
	}

	pq.RemoveSimItemByCardIndex(cardIndex)
	card.SetStartAtSimTime = 0
	card.ReadyAtSimTime = 0
	card.Ready = false

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
