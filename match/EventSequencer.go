package match

import (
	"fmt"
)

// so we have to balance between client inputs and autoevents
// every client input allows us to predict the next n+1 autoevent
// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// I have one loop that handles
// pushing InputEventChan into the priority queue
// taking out events from the priority queue
//    - handling incrementing currentEvents
//    - handling broadcasting InputEvents
//    - handling broadcasting AutoSchedulers

func (match *Match) EventScheduler() {

	for {
		select {
		case ie := <-match.InputEventChan:

			fmt.Println("ie", ie)

		case <-match.EventPQTimer.C:
		}
	}

	// for ie := range match.InputEventChan {
	// 	fmt.Println("inputEvent", ie.T.UnixMilli(), ie.Type, ie.TeamIdx, ie.UnitIdx)

	// 	// remove all from the gameState array not within the next 200 ms, assume at least 1 GameState exists
	// 	deadline := ie.T.Add(200 * time.Millisecond)
	// 	for {
	// 		back := match.GameStates[len(match.GameStates)-1]
	// 		if deadline.Before(back.Timestamp) {
	// 			match.GameStates = match.GameStates[:len(match.GameStates)-1] // pop, O(1) btw
	// 		} else {
	// 			break
	// 		}
	// 	}

	// 	// create a new gameState with this applied
	// 	newState := summon.BasicSummon(match.UnitModels, match.GameStates[len(match.GameStates)-1], deadline, ie.TeamIdx, ie.UnitIdx)
	// 	match.GameStates = append(match.GameStates, newState)

	// 	// broadcast it
	// 	match.BroadcastLastGameState() // TODO: this does not have to be a state, can be smaller event object, the client can create the state off of it

	// 	// TODO: schedule autoscheduler

	// 	// let UpdateCurrentState know that GameStates arr has changed
	// 	select {
	// 	case match.GameStatesChanged <- struct{}{}:
	// 	default: // drop notification if one is already pending
	// 	}
	// }

}

// we have two goroutines because of the following case:
// imagine we have an event every 1 ms
// if we put the currentState update and the scheduling in the same goroutine/for loop
// there is a very small chance that UpdateCurrentState continuously runs and scheduling does not happen

// TODO: should probably mutex GameState right
// func (match *Match) UpdateCurrentState() {

// 	// init
// 	match.setCurrTimer()

// 	for {

// 		// if no currentTimer
// 		if match.CurrentStateTimer == nil {
// 			<-match.GameStatesChanged
// 			match.setCurrTimer()
// 			continue
// 		}

// 		// if we have currentTimer
// 		select {
// 		case <-match.GameStatesChanged:
// 			match.setCurrTimer()
// 		case <-match.CurrentStateTimer.C:
// 			match.setCurrTimer()
// 		}
// 	}
// }

// func (match *Match) setCurrTimer() {

// 	// increment currentState to most recent state
// 	for match.CurrentState+1 < len(match.GameStates) && time.Now().After(match.GameStates[match.CurrentState].Timestamp) {
// 		match.CurrentState += 1
// 	}

// 	// stop existing timer and set to nil
// 	if match.CurrentStateTimer != nil {
// 		if !match.CurrentStateTimer.Stop() {
// 			select {
// 			case <-match.CurrentStateTimer.C: // non-blocking drain channel
// 			default:
// 			}
// 		}
// 		match.CurrentStateTimer = nil
// 	}

// 	// set timer
// 	if match.CurrentState+1 < len(match.GameStates) {
// 		d := time.Until(match.GameStates[match.CurrentState+1].Timestamp)
// 		match.CurrentStateTimer = time.NewTimer(d)
// 	}
// }
