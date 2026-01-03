package match

import (
	"time"
)

func ScheduleInputEvent(timestamp time.Time, match *Match, eventType string, clientID string, unitIdx int, funcIdx int, target1 int, target2 int) {

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
		Type:    eventType,
		TeamIdx: int(clientIdx),
		UnitIdx: unitIdx,
		FuncIdx: funcIdx,
		Target1: target1,
		Target2: target2,
	}

	match.InputEventChan <- ie

}
