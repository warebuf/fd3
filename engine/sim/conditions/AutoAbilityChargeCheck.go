package conditions

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// TODO: a ZoneField check should be decided by the ability's con
func AutoAbilityChargeCheck(gs *state.GameState, cardID int, activeIndex int) bool {
	card := gs.Cards[cardID]
	if card.Zone != state.ZoneField {
		return false
	}

	ability := defs.CardDefs[card.DefNo].Active[activeIndex]
	if !ability.AutoActivate {
		return false
	}

	// if any effects exist
	if ability.Effects[0].Type == 0 {
		return false
	}

	if card.Actives[activeIndex].Charges <= 0 {
		return false
	}

	return CheckAbilityConditions(gs, cardID, activeIndex)
}
