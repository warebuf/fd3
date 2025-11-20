package match

import (
	"fmt"
	"time"
)

func ScheduleSummonEvent(timestamp time.Time, match *Match, clientID string, unitIdx uint) {

	// check if client is in the whitelist
	clientIdx, ok := match.Whitelist[clientID]
	if !ok {
		return
	}

	// check if the summon is valid

	fmt.Println(timestamp, "summon", clientIdx, unitIdx)

	ie := &InputEvent{
		T:    timestamp,
		Type: "Summon",
	}

	match.InputEventChan <- ie

}
