package destruction

import "github.com/warebuf/fd3/engine/state"

// clears any preselected Active target pointing at cardIndex, e.g. after it's destroyed
func ClearPreselectsTargeting(gs *state.GameState, cardIndex int) {
	for i := range gs.Cards {
		for j := range gs.Cards[i].Actives {
			if gs.Cards[i].Actives[j].Target == cardIndex {
				gs.Cards[i].Actives[j].Target = -1
			}
		}
	}
}
