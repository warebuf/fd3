package conditions

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

func AutoAbilityChargeCheck(gs *state.GameState, cardID int, activeIndex int) bool {
	card := gs.Cards[cardID]
	if card.Zone != state.ZoneField {
		return false
	}

	active := card.Actives[activeIndex]
	if !active.AutoAbility {
		return false
	}

	ability := defs.CardDefs[card.DefNo].Active[activeIndex]
	// if any effects exist
	if ability.Effects[0].Type == 0 {
		return false
	}

	if active.Charges <= 0 {
		return false
	}

	return CheckAbilityConditions(gs, cardID, activeIndex)
}
