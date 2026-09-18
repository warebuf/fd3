package conditions

import (
	"github.com/warebuf/fd3/engine/state"
)

// checks whether an active ability can be played in the current state
func CheckActiveAbilityConditions(gs *state.GameState, cardIndex int, activeIndex int) bool {
	card := gs.Cards[cardIndex]
	active := card.Actives[activeIndex]

	// each ability has 3 slots for conditions
	for i := 0; i < len(active.Cons); i++ {
		c := active.Cons[i]
		if c == 0 {
			return true // no gaps: first unused slot means no more conditions to check
		}
		switch c {
		case state.ConditionManaAtLeast:
			if gs.Mana[card.Team] < active.Cost {
				return false
			}
		case state.ConditionChargeAtLeastOne:
			if card.Charges < 1 {
				return false
			}
		case state.ConditionTargetOwnedBySelf:
			target := active.Target
			if target < 0 || target >= len(gs.Cards) || gs.Cards[target].OwnerTeam != card.Team {
				return false
			}
		default:
			return false // unknown/unimplemented condition type - fail closed
		}
	}
	return true
}
