package conditions

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

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

	return CheckAbilityConditions(gs, cardIndex, activeIndex)
}
