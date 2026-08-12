package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions/autoability"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectChargeSelf(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	card := &gs.Cards[sourceCardIndex]
	if card.Charges < card.MaxCharges {
		card.Charges++
		*logList = append(*logList, logs.AddChargeLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.AddCharge,
			TeamIndex:     card.Team,
			CardIndex:     sourceCardIndex,
			Charges:       card.Charges,
		})
		*UpdateID++
	}

	// Charges is shared across the card, so any Active slot could now be auto-fireable
	for activeIndex := range gs.Cards[sourceCardIndex].Actives {
		if autoability.AutoAbilityChargeCheck(gs, sourceCardIndex, activeIndex) {
			chain.PushAbilityOnChain(gs, logList, UpdateID, sourceCardIndex, activeIndex, executeAtWallTime)
		}
	}
}
