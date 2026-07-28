package effects

import "github.com/warebuf/fd3/engine/state"

// helper, check if card is still in the correct zone
func isValidPersistentTarget(gs *state.GameState, teamIndex int, targetIndex int) bool {
	for _, ci := range gs.Field[teamIndex] {
		if ci == targetIndex {
			return true
		}
	}
	for _, ci := range gs.Continuous[teamIndex] {
		if ci == targetIndex {
			return true
		}
	}
	return false
}
