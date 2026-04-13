package match

import (
	"time"
)

func ScheduleInputEvent(timestamp time.Time, match *Match, eventType string, clientID string, cardIndex int) {

	// check if client is in the whitelist
	clientIdx, ok := match.Whitelist[clientID]
	if !ok {
		return
	}

	// this function should only be doing rate limit checks, not sim checks

	// can check that unitIdx is non negative
	if cardIndex < 0 {
		return
	}

	ie := &InputEvent{
		T:         timestamp,
		Type:      eventType,
		TeamIdx:   int(clientIdx),
		CardIndex: cardIndex,
	}

	match.InputEventChan <- ie

}
