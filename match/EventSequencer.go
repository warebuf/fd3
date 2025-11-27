package match

import (
	"container/heap"
	"fmt"
	"math"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilities"
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

			autoEventTime := next.Event.Timestamp().Add(200 * time.Millisecond)

			//increment
			incrementedState := incrementState(match.GameStates[len(match.GameStates)-1], autoEventTime)
			match.BroadcastGameState(incrementedState)
			pushCurrentEvent(match, autoEventTime)

			// figure out which abilities are due
			dueUnits := []int{} // matches length of dueAbilities
			dueAbilities := []int{}
			for i := 0; i < len(incrementedState.AbilityStartIdx); i++ {
				for j := 0; j < incrementedState.AbilityCount[i]; j++ {
					abilityIdx := incrementedState.AbilityStartIdx[i] + j
					if incrementedState.AbilityCharge[abilityIdx] >= incrementedState.AbilitySPD[abilityIdx] {
						dueAbilities = append(dueAbilities, abilityIdx)
						dueUnits = append(dueUnits, i)
					}
				}
			}

			// activate abilities
			abilityState, modified := handleDueAbilities(incrementedState, dueUnits, dueAbilities)
			if modified {
				match.GameStates = append(match.GameStates, abilityState)
				match.BroadcastGameState(abilityState)
				pushCurrentEvent(match, autoEventTime)
			}

			// schedule next AutoScheduler
			nextAutoEventTime, success := calcNextAutoEvent(abilityState, true)
			if success {
				pushAutoScheduler(match, nextAutoEventTime)
			}

		case "Current":
			match.CurrentState += 1
			fmt.Println("received Current ->", match.CurrentState, len(match.GameStates))

		case "Input":
			fmt.Println("received Input")
			ie := next.Event.(*InputEvent)

			// remove all autoschedulers

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
			nextAutoEventTime, success := calcNextAutoEvent(newState, true)
			if success {
				pushAutoScheduler(match, nextAutoEventTime)
			}

		}

	}

	// set timer for next event or set it to nil
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
func calcNextAutoEvent(state *engine.GameState, ignoreDue bool) (time.Time, bool) {

	timeToAutoEvent := math.MaxInt

	for i := 0; i < len(state.Zone); i++ {

		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.AbilityStartIdx[i]; j < state.AbilityStartIdx[i]+state.AbilityCount[i]; j++ {
				timeLeft := int(state.AbilitySPD[j] - state.AbilityCharge[j])
				if timeLeft == 0 && ignoreDue {
					continue
				} else if timeLeft < timeToAutoEvent {
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

// could use an array of functions but will use a switch statement for now

func handleDueAbilities(incrementedState *engine.GameState, dueUnits []int, dueAbilities []int) (*engine.GameState, bool) {

	fmt.Println("dueUnits", dueUnits)
	fmt.Println("dueAbilities", dueAbilities)

	newState := engine.CopyGameState(incrementedState.Timestamp, incrementedState)
	modified := false

	for i := 0; i < len(dueAbilities); i++ {

		dueUnit := dueUnits[i]
		dueAbility := dueAbilities[i]
		abilityNo := newState.Abilities[dueAbility]

		switch abilityNo {
		case 0:
			fmt.Println("used A00000")
			occurred := abilities.A00000(newState, dueUnit)
			fmt.Println("occurred", occurred)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 1:
			fmt.Println("used A00001")
			occurred := abilities.A00001(newState)
			fmt.Println("occurred", occurred)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 2:
			fmt.Println("used A00002")
			occurred := abilities.A00002(newState, dueUnit)
			fmt.Println("occurred", occurred)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		}

	}

	return newState, modified
}
