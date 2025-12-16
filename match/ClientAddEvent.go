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

	// this function should only be doing rate limit checks, not sim checks

	// can check that unitIdx is non negative
	if unitIdx < 0 {
		return
	}

	ie := &InputEvent{
		T:       timestamp,
		Type:    "Summon",
		TeamIdx: int(clientIdx),
		UnitIdx: unitIdx,
	}

	match.InputEventChan <- ie

}
