package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions/autoability"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectChargeTarget(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team
	targetIndex := eff.A

	// at the end of the function, play the card's autoability if possible
	defer func() {
		for activeIndex := range gs.Cards[targetIndex].Actives {
			if autoability.AutoAbilityChargeCheck(gs, targetIndex, activeIndex) {
				chain.PushAbilityOnChain(gs, logList, UpdateID, targetIndex, activeIndex, executeAtWallTime)
			}
		}
	}()

	if !isValidPersistentTargetAnyTeam(gs, targetIndex) {
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
		TeamIndex:     card.Team, // the target's owner, since the target may be an enemy card
		CardIndex:     targetIndex,
		Charges:       card.Charges,
	})
	*UpdateID++
}
