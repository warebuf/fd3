package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectCounter(gs *state.GameState, logList *[]any, UpdateID *int, sourceCardID int, executeAtWallTime int64) {
	if len(gs.Chain) == 0 {
		return
	}

	target := gs.Chain[len(gs.Chain)-1]
	gs.Chain = gs.Chain[:len(gs.Chain)-1]

	card := &gs.Cards[target.CardID]
	if card.Zone == state.ZoneChain {
		card.Zone = state.ZoneGraveyard
		gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardID)
	}

	*logList = append(*logList, logs.CounterChainLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.CounterChain,
		SourceCardID:  sourceCardID,
		TargetCardID:  target.CardID,
	})
	*UpdateID++
}
