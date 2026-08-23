package destruction

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// removes an equip card from its host's team, clears ability timers, clears preselects pointing at it
func HandleEquipDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	card := &gs.Cards[cardIndex]
	if card.Zone != state.ZoneEquip {
		return
	}
	cardTeam := card.Team

	for i, ci := range gs.Equip[cardTeam] {
		if ci == cardIndex {
			gs.Equip[cardTeam] = append(gs.Equip[cardTeam][:i], gs.Equip[cardTeam][i+1:]...)
			break
		}
	}
	card.Zone = state.ZoneGraveyard
	if !card.IsToken {
		gs.Grave[cardTeam] = append(gs.Grave[cardTeam], cardIndex)
	}

	pq.RemoveSimItemByCardIndex(cardIndex)
	card.OnTick = [3]state.TickAbility{}
	card.EquipHostIndex = -1

	ClearPreselectsTargeting(gs, cardIndex)
	RemoveChainItemsForCard(gs, logList, UpdateID, cardIndex, executeAtWallTime)

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
