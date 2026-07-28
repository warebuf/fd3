package conditions

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// when mana is added, check if any field/continuous cards can autoability
func AutoAbilityManaCheck(gs *state.GameState, teamIndex int) ([]int, []int) {
	var cardIndexes []int
	var activeIndexes []int

	var candidates []int
	candidates = append(candidates, gs.Field[teamIndex]...)
	candidates = append(candidates, gs.Continuous[teamIndex]...)

	for i := 0; i < len(candidates); i++ {
		cardIndex := candidates[i]
		card := gs.Cards[cardIndex]
		def := defs.CardDefs[card.DefNo]

		for activeIndex := 0; activeIndex < len(def.Active); activeIndex++ {
			ability := def.Active[activeIndex]
			active := card.Actives[activeIndex]

			if !active.AutoAbility {
				continue
			}

			if !ability.Unlimited && card.Charges <= 0 {
				continue
			}

			if !ability.Unlimited && card.PendingCharges >= card.Charges {
				continue
			}

			if gs.Mana[teamIndex] < ability.Cost {
				continue
			}

			if CheckAbilityConditions(gs, cardIndex, activeIndex) {
				cardIndexes = append(cardIndexes, cardIndex)
				activeIndexes = append(activeIndexes, activeIndex)
			}
		}
	}

	return cardIndexes, activeIndexes
}
