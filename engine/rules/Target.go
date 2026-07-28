package rules

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// checks if targetIndex is legal given the targetFlags, from the perspective of teamIndex
func IsValidTarget(gs *state.GameState, targetFlags int, sourceIndex int, targetIndex int, teamIndex int) bool {
	if targetFlags == 0 {
		return false
	}
	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return false
	}
	target := gs.Cards[targetIndex]
	if targetIndex == sourceIndex {
		return targetFlags&defs.TargetSelf != 0
	}
	switch {
	case target.Team == teamIndex && target.Zone == state.ZoneField:
		return targetFlags&defs.TargetOwnField != 0
	case target.Team != teamIndex && target.Zone == state.ZoneField:
		return targetFlags&defs.TargetOppField != 0
	case target.Team == teamIndex && target.Zone == state.ZoneHand:
		return targetFlags&defs.TargetOwnHand != 0
	case target.Team != teamIndex && target.Zone == state.ZoneHand:
		return targetFlags&defs.TargetOppHand != 0
	case target.Team == teamIndex && target.Zone == state.ZoneContinuous:
		return targetFlags&defs.TargetOwnCont != 0
	case target.Team != teamIndex && target.Zone == state.ZoneContinuous:
		return targetFlags&defs.TargetOppCont != 0
	}
	return false
}
