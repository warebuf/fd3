package match

import (
	"container/heap"
	"fmt"
	"time"
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

func pushInputEvent(match *Match, ev *InputEvent) {
	match.NextSeq++
	ev.T = time.Now() // relabel the timestamp when putting it into PQ
	item := &EventItem{
		Event: ev,
		Type:  "Input",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

func (match *Match) EventScheduler() {

	// on init
	match.setCurrentState()

	for {

		// clear due PQ events
		match.handleDuePQEvents()

		// check if timer is set
		var timerC <-chan time.Time
		if match.EventPQTimer != nil {
			timerC = match.EventPQTimer.C
		}

		// wait on input or PQ timer
		select {
		case ie := <-match.InputEventChan:
			fmt.Println("pushed ie", ie)
			pushInputEvent(match, ie)

		case <-timerC:
			fmt.Println("timer fired")
			match.handleDuePQEvents()
		}
	}

}

func (match *Match) setCurrentState() {
	// increment currentState to most recent state
	for match.CurrentState+1 < len(match.GameStates) && time.Now().After(match.GameStates[match.CurrentState].Timestamp) {
		match.CurrentState += 1
	}
}

func (match *Match) handleDuePQEvents() {
	for {

		// return if empty
		if match.EventPQueue.Len() == 0 {
			return
		}

		next := match.EventPQueue[0]

		// no more due events
		if time.Now().Before(next.Event.Timestamp()) {
			return
		}

		heap.Pop(&match.EventPQueue) // pop the event

		// handle the event
		switch next.Type {
		case "Input":
			ie := next.Event.(*InputEvent)
			fmt.Println("ie", ie)

			// think you have to remove all autoschedulers and their associated autoevents

			// 	// create a new gameState with this applied
			// 	newState := summon.BasicSummon(match.UnitModels, match.GameStates[len(match.GameStates)-1], deadline, ie.TeamIdx, ie.UnitIdx)
			// 	match.GameStates = append(match.GameStates, newState)

			// 	// broadcast it
			// 	match.BroadcastLastGameState() // TODO: this does not have to be a state, can be smaller event object, the client can create the state off of it

			// 	// TODO: schedule autoscheduler

		}

	}
}
