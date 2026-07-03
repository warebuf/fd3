package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectGainCharge(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardID int, executeAtWallTime int64) {
	card := &gs.Cards[sourceCardID]
	active := &card.Actives[eff.A] // eff.A: which Active slot this charges
	if active.Charges < active.MaxCharges {
		active.Charges++
		*logList = append(*logList, logs.AddChargeLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.AddCharge,
			TeamIndex:     card.Team,
			CardID:        sourceCardID,
			Charges:       active.Charges,
		})
		*UpdateID++
	}
}
