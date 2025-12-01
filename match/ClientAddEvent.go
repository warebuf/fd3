package match

import (
	"time"
)

func ScheduleSummonEvent(timestamp time.Time, match *Match, clientID string, unitIdx int) {

	// check if client is in the whitelist
	clientIdx, ok := match.Whitelist[clientID]
	if !ok {
		return
	}

	// match has not started
	if match.GameStates[match.CurrentState].Round == 0 {
		return
	}

	// check if the summon is valid

	ie := &InputEvent{
		T:       timestamp,
		Type:    "Summon",
		TeamIdx: int(clientIdx),
		UnitIdx: unitIdx,
	}

	match.InputEventChan <- ie

}
