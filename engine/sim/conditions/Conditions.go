package conditions

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// checks whether an ability can be played in the current state
func CheckAbilityConditions(gs *state.GameState, cardID int, activeIndex int) bool {
	card := gs.Cards[cardID]
	ability := defs.CardDefs[card.DefNo].Active[activeIndex]

	for i := 0; i < len(ability.Cons); i++ {
		c := ability.Cons[i]
		if c == 0 {
			return true // no gaps: first unused slot means no more conditions to check
		}
		switch c {
		case defs.ConditionManaAtLeast:
			if gs.Mana[card.Team] < ability.Cost {
				return false
			}
		case defs.ConditionChargeAtLeastOne:
			if card.Actives[activeIndex].Charges < 1 {
				return false
			}
		default:
			return false // unknown/unimplemented condition type - fail closed
		}
	}
	return true
}
