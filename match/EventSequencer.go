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

	// init
	fmt.Println("now", time.Now().UnixMilli())
	fmt.Println("state0", match.GameStates[match.CurrentState].Timestamp.UnixMilli())
	fmt.Println("Abilities", match.GameStates[match.CurrentState].Abilities)
	fmt.Println("AbilityCharge", match.GameStates[match.CurrentState].AbilityCharge)
	fmt.Println("AbilityCount", match.GameStates[match.CurrentState].AbilityCount)
	fmt.Println("AbilitySPD", match.GameStates[match.CurrentState].AbilitySPD)
	fmt.Println("AbilityStartIdx", match.GameStates[match.CurrentState].AbilityStartIdx)

	//var minTime int64 = math.MaxInt64
	for i := 0; i < len(match.GameStates[match.CurrentState].Zone); i++ {

		if match.GameStates[match.CurrentState].Zone[i] == 1 && match.GameStates[match.CurrentState].HP[i] > 0 {

		}

	}

	for ie := range match.InputEventChan {
		fmt.Println("inputEvent", ie)
		// add the event to the array
		// broadcast
		// calculate the n+1 event?
	}

}
