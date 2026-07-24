package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

// IsValidTarget checks whether targetIndex is a legal target for a source card
// with the given TargetFlags bitmask, from the perspective of teamIndex.
func IsValidTarget(gs *state.GameState, targetFlags int, sourceIndex int, targetIndex int, teamIndex int) bool {
	if targetFlags == 0 {
		return false
	}
	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return false
	}
	target := gs.Cards[targetIndex]
	if targetIndex == sourceIndex {
		return targetFlags&TargetSelf != 0
	}
	switch {
	case target.Team == teamIndex && target.Zone == state.ZoneField:
		return targetFlags&TargetOwnField != 0
	case target.Team != teamIndex && target.Zone == state.ZoneField:
		return targetFlags&TargetOppField != 0
	case target.Team == teamIndex && target.Zone == state.ZoneHand:
		return targetFlags&TargetOwnHand != 0
	case target.Team != teamIndex && target.Zone == state.ZoneHand:
		return targetFlags&TargetOppHand != 0
	case target.Team == teamIndex && target.Zone == state.ZoneContinuous:
		return targetFlags&TargetOwnCont != 0
	case target.Team != teamIndex && target.Zone == state.ZoneContinuous:
		return targetFlags&TargetOppCont != 0
	}
	return false
}
