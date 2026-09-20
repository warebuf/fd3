package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectCounter(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, sourceCardIndex int, executeAtWallTime int64) {
	if len(gs.Chain) == 0 {
		return
	}

	target := gs.Chain[len(gs.Chain)-1]
	gs.Chain = gs.Chain[:len(gs.Chain)-1]

	if target.Type != state.ChainStructureAbility {
		card := &gs.Cards[target.CardIndex]
		switch target.Type {
		case state.ChainTrap:
			if card.Zone == state.ZoneTrap {
				for i, ci := range gs.Traps[card.Team] {
					if ci == target.CardIndex {
						gs.Traps[card.Team] = append(gs.Traps[card.Team][:i], gs.Traps[card.Team][i+1:]...)
						break
					}
				}
				card.Zone = state.ZoneGraveyard
				gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardIndex)
				destruction.RemoveChainItemsForCard(gs, logList, UpdateID, target.CardIndex, executeAtWallTime)
			}
		case state.ChainSummon, state.ChainEquip:
			card.Zone = state.ZoneGraveyard
			gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardIndex)
			destruction.RemoveChainItemsForCard(gs, logList, UpdateID, target.CardIndex, executeAtWallTime)
		case state.ChainUnitAbility:
			destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, target.CardIndex, executeAtWallTime)
		}
	}

	*logList = append(*logList, logs.CounterChainLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.CounterChain,
		SourceCardIndex: sourceCardIndex,
		TargetCardIndex: target.CardIndex,
	})
	*UpdateID++
}
