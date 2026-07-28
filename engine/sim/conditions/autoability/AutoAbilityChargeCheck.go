package autoability

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/state"
)

// when a charge is added to a card, use this function to check if an autoability should/can fire
func AutoAbilityChargeCheck(gs *state.GameState, cardIndex int, activeIndex int) bool {
	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneField {
		return false
	}

	ability := defs.CardDefs[card.DefNo].Active[activeIndex]
	active := card.Actives[activeIndex]
	if !active.AutoAbility {
		return false
	}

	if !ability.Unlimited && card.Charges <= 0 {
		return false
	}

	if !ability.Unlimited && card.PendingCharges >= card.Charges {
		return false
	}

	if gs.Mana[card.Team] < ability.Cost {
		return false
	}

	return conditions.CheckAbilityConditions(gs, cardIndex, activeIndex) // check if the ability mets it's conditions
}
