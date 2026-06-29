package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectGainCharge(gs *state.GameState, logList *[]any, UpdateID *int, sourceCardID int, executeAtWallTime int64) {
	card := &gs.Cards[sourceCardID]
	if card.Charges < card.MaxCharges {
		card.Charges++
		*logList = append(*logList, logs.AddChargeLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.AddCharge,
			TeamIndex:     card.Team,
			CardID:        sourceCardID,
			Charges:       card.Charges,
		})
		*UpdateID++
	}
}
