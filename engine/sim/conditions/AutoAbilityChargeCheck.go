package conditions

import (
	"github.com/warebuf/fd3/engine/state"
)

func AutoAbilityChargeCheck(gs *state.GameState, cardIndex int, activeIndex int) bool {
	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneField {
		return false
	}

	active := card.Actives[activeIndex]
	if !active.AutoAbility {
		return false
	}

	if active.Charges <= 0 {
		return false
	}

	return CheckAbilityConditions(gs, cardIndex, activeIndex)
}
