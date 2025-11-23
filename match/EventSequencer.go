package match

import (
	"fmt"
	"time"
)

// so we have to balance between client inputs and autoevents

// every client input allows us to predict the next n+1 autoevent

// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// we can work out all cases to write the EventSequencer / EventScheduler

// we can have an Event data structure, with inputs, autoevents, and autoschedulers
// we can also have a GameState data structure

// soo...
// inputs can look pretty simple, pretty much just {userIdx, unitIdx, actionType}, just need to do a verification step
// event types can be pretty simple too, some sort of {ability, unitIdx}

func (match *Match) EventScheduler() {

	// increment currentState to most recent state
	for match.CurrentState+1 < len(match.GameStates) && time.Now().After(match.GameStates[match.CurrentState].Timestamp) {
		match.CurrentState += 1
	}
	fmt.Println("CurrentState", match.CurrentState)

	for ie := range match.InputEventChan {
		fmt.Println("inputEvent", ie.T.UnixMilli(), ie.Type)
		// remove all from the gameState array not within the next 200 ms
		

		// create a new gameState with this applied
		// broadcast it
		// TODO: schedule autoscheduler
	}

}
