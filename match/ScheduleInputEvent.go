package match

import (
	"time"

	"github.com/warebuf/fd3/match/eventpq"
)

func ScheduleInputEvent(timestamp time.Time, match *Match, eventType string, clientID string, cardIndex int, targetCardIndex int) {

	// check if client is in the whitelist
	clientIndex, ok := match.Whitelist[clientID]
	if !ok {
		return
	}

	// this function should only be doing rate limit checks, not sim checks

	// can check that unitIndex is non negative
	if cardIndex < 0 {
		return
	}

	ie := &eventpq.InputEvent{
		WallTime:        timestamp,
		Type:            eventType,
		TeamIndex:       int(clientIndex),
		CardIndex:       cardIndex,
		TargetCardIndex: targetCardIndex,
	}

	match.InputEventChan <- ie

}
