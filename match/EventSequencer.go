package match

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/summon"
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

	for ie := range match.InputEventChan {
		fmt.Println("inputEvent", ie)

		// remove all from the gameState array not within the next 200 ms
		curr := match.CurrentState
		deadline := ie.T.Add(200 * time.Millisecond)
		for curr < len(match.GameStates) && deadline.After(match.GameStates[curr].Timestamp) {
			curr += 1
		}
		newGameStatesArr := match.GameStates[0:curr]

		// create a new gameState with this applied
		newState := summon.BasicSummon(match.UnitModels, match.GameStates[len(match.GameStates)-1], deadline, ie.TeamIdx, ie.UnitIdx)
		newGameStatesArr = append(newGameStatesArr, newState)
		match.GameStates = newGameStatesArr

		// broadcast it
		match.BroadcastLastGameState()

		// TODO: schedule autoscheduler
	}

}
