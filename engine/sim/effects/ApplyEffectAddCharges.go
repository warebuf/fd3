package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectAddCharges(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team
	targetIndex := eff.A

	// Charges is shared across the card, so any Active slot on the target could now be auto-fireable
	defer func() {
		for activeIndex := range gs.Cards[targetIndex].Actives {
			if conditions.AutoAbilityChargeCheck(gs, targetIndex, activeIndex) {
				reaction.PushAbilityOnStack(gs, logList, UpdateID, targetIndex, activeIndex, executeAtWallTime)
			}
		}
	}()

	valid := false
	for _, ci := range gs.Field[teamIndex] {
		if ci == targetIndex {
			valid = true
			break
		}
	}
	if !valid {
		for _, ci := range gs.Continuous[teamIndex] {
			if ci == targetIndex {
				valid = true
				break
			}
		}
	}
	if !valid {
		return // preselected target is gone or was never set, ability fizzles
	}

	switch eff.C {
	case 16: // Kindling's DefNo
		// fizzles once any of the caster's own cards already has 2+ charges - a one-time kickstart, not a repeatable pump
		for _, ci := range gs.Field[teamIndex] {
			if gs.Cards[ci].Charges >= 2 {
				return
			}
		}
		for _, ci := range gs.Continuous[teamIndex] {
			if gs.Cards[ci].Charges >= 2 {
				return
			}
		}
	}

	card := &gs.Cards[targetIndex]
	newCharges := card.Charges + eff.B
	if newCharges > card.MaxCharges {
		newCharges = card.MaxCharges
	}
	if newCharges <= card.Charges {
		return
	}
	card.Charges = newCharges

	*logList = append(*logList, logs.AddChargeLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.AddCharge,
		TeamIndex:     teamIndex,
		CardIndex:     targetIndex,
		Charges:       card.Charges,
	})
	*UpdateID++
}
