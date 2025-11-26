package match

import (
	"container/heap"
	"fmt"
	"math"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/summon"
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
			pushInputEvent(match, ie)

		case <-timerC:
			fmt.Println("timer fired")
			match.handleDuePQEvents()
		}
	}

}

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

func pushCurrentEvent(match *Match, timestamp time.Time) {
	currentEvent := &CurrentEvent{
		T: timestamp,
	}

	match.NextSeq++
	item := &EventItem{
		Event: currentEvent,
		Type:  "Current",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

// AutoScheduler happens 200 ms earlier
func pushAutoScheduler(match *Match, nextAutoEvent time.Time) {
	currentEvent := &CurrentEvent{
		T: nextAutoEvent.Add(-200 * time.Millisecond),
	}

	match.NextSeq++
	item := &EventItem{
		Event: currentEvent,
		Type:  "AutoScheduler",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
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
			break
		}

		heap.Pop(&match.EventPQueue) // pop the event

		// handle the event
		switch next.Type {
		case "AutoScheduler":
			fmt.Println("received AutoScheduler")

		case "Current":
			match.CurrentState += 1
			fmt.Println("received Current ->", match.CurrentState)

		case "Input":
			fmt.Println("received Input")
			ie := next.Event.(*InputEvent)

			// remove all autoschedulers and their associated autoevents

			// increment the state from the last confirmed GameState
			nextTimestamp := ie.T.Add(200 * time.Millisecond)
			incrementedState := incrementState(match.GameStates[len(match.GameStates)-1], nextTimestamp)
			match.BroadcastGameState(incrementedState)
			pushCurrentEvent(match, nextTimestamp)

			// apply the event and create a new GameState (TODO: broadcast an event instead of a full GameState)
			newState := summon.BasicSummon(match.UnitModels, incrementedState, nextTimestamp, ie.TeamIdx, ie.UnitIdx)
			match.GameStates = append(match.GameStates, newState)
			match.BroadcastGameState(newState)
			pushCurrentEvent(match, nextTimestamp) // add the currentEvent to the PQ

			// schedule autoscheduler
			nextAutoEvent, success := calcNextAutoEvent(newState)
			if success {
				pushAutoScheduler(match, nextAutoEvent)
			}

		}

	}

	// set timer for next event or set it to nil
}

// increment currentState to most recent state
func (match *Match) setCurrentState() {
	for match.CurrentState+1 < len(match.GameStates) && time.Now().After(match.GameStates[match.CurrentState].Timestamp) {
		match.CurrentState += 1
	}
}

func incrementState(state *engine.GameState, nextTimestamp time.Time) *engine.GameState {

	newState := engine.CopyGameState(nextTimestamp, state)
	increment := nextTimestamp.UnixMilli() - state.Timestamp.UnixMilli()
	fmt.Println("since last state", increment)

	for i := 0; i < len(state.Zone); i++ {
		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.AbilityStartIdx[i]; j < state.AbilityStartIdx[i]+state.AbilityCount[i]; j++ {

				if state.AbilitySPD[j]-increment < state.AbilityCharge[j] { // we've reached past the charge
					//fmt.Println("we've reached past the charge???", state.AbilityCharge[j]+increment)
					newState.AbilityCharge[j] = newState.AbilitySPD[j]
				} else {
					newState.AbilityCharge[j] += increment
				}
			}
		}
	}

	return newState
}

// if we return a negative number, no autoevent found
func calcNextAutoEvent(state *engine.GameState) (time.Time, bool) {

	timeToAutoEvent := math.MaxInt

	for i := 0; i < len(state.Zone); i++ {

		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.AbilityStartIdx[i]; j < state.AbilityStartIdx[i]+state.AbilityCount[i]; j++ {
				timeLeft := int(state.AbilitySPD[j] - state.AbilityCharge[j])
				if timeLeft < timeToAutoEvent {
					timeToAutoEvent = timeLeft
				}

			}
		}

	}

	fmt.Println("next AutoEvent in", timeToAutoEvent)

	if timeToAutoEvent == math.MaxInt {
		return time.Time{}, false
	}

	return state.Timestamp.Add(time.Millisecond * time.Duration(timeToAutoEvent)), true

}

func handleAutoScheduler() {

}
