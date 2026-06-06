package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectCounter(gs *state.GameState, logList *[]any, UpdateID *int, sourceCardID int, executeAtWallTime int64) {
	if len(gs.Stack) == 0 {
		return
	}

	target := gs.Stack[len(gs.Stack)-1]
	gs.Stack = gs.Stack[:len(gs.Stack)-1]

	card := &gs.Cards[target.CardID]
	if card.Zone == state.ZoneStack {
		card.Zone = state.ZoneGraveyard
		gs.Grave[card.Team] = append(gs.Grave[card.Team], target.CardID)
	}

	*logList = append(*logList, logs.CounterStackLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.CounterStack,
		SourceCardID:  sourceCardID,
		TargetCardID:  target.CardID,
	})
	*UpdateID++
}
