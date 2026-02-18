package match

import (
	"container/heap"
	"fmt"
	"math"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/simulate"
	"github.com/warebuf/fd3/engine/summon"
	"github.com/warebuf/fd3/engine/target"
)

// so we have to balance between client inputs and autoevents
// every client input allows us to predict the next n+1 autoevent
// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// this one loop handles pushing inputs into a priority queue
// and taking out Events, Inputs, and AutoSchedulers from it

func (match *Match) EventScheduler() {

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

			//increment
			autoEventTime := next.Event.Timestamp().Add(200 * time.Millisecond) // HARDCODED b/c autoschedulers are scheduled 200 ms before the autoevent
			incrementedState := IncrementState(match.GameState, autoEventTime)

			// figure out which abilities are due
			dueUnits := []int{} // matches length of dueAbilities
			dueFunc := []int{}
			for i := 0; i < len(incrementedState.FunctionStartIdx); i++ {
				for j := 0; j < incrementedState.FunctionCount[i]; j++ {
					funcIdx := incrementedState.FunctionStartIdx[i] + j
					if incrementedState.FunctionType[funcIdx] == 0 { // only check auto functions
						isValid := CheckValidFunction(incrementedState, i, incrementedState.Functions[funcIdx])
						if incrementedState.FunctionCharge[funcIdx] >= incrementedState.FunctionSPD[funcIdx] && isValid {
							dueFunc = append(dueFunc, funcIdx)
							dueUnits = append(dueUnits, i)
						}
					}
				}
			}

			// activate abilities
			abilityState, modified := HandleDueFunctions(incrementedState, dueUnits, dueFunc)
			if modified {
				match.GameStates = append(match.GameStates, abilityState)
				match.BroadcastGameState(abilityState)
			}

			// schedule next AutoScheduler
			nextAutoEventTime, success := CalcNextAutoEvent(abilityState)
			if success {
				pushAutoScheduler(match, nextAutoEventTime)
			}

		case "Input":
			ie := next.Event.(*InputEvent)
			fmt.Println("received Input", ie)

			// TODO: notably there is little to no validation

			// remove all autoscheduler events (at most only 1 exists in current design)
			match.EventPQueue.RemoveAllAutoSchedulers()

			// increment the state from the last confirmed GameState
			nextTimestamp := ie.T.Add(200 * time.Millisecond)
			incrementedState := IncrementState(match.GameState, nextTimestamp)

			if ie.Type == "Summon" {

				// apply the event
				newActions := summon.BasicSummon(match.UnitModels, incrementedState, nextTimestamp, ie.TeamIdx, ie.UnitIdx)
				fmt.Println("newActions", newActions)

				// publish the new state
				match.PublishedState = engine.CopyGameState(match.GameState.Timestamp, match.GameState)

				// append the new actions
				match.ActionLog = append(match.ActionLog, newActions...)

				// share the new action
				if len(newActions) > 0 {
					match.BroadcastActions(newActions)
				}

				// schedule autoscheduler
				nextAutoEventTime, success := CalcNextAutoEvent(match.GameState)
				if success {
					pushAutoScheduler(match, nextAutoEventTime)
				}

			} else if ie.Type == "Active" {

				// check whether the active is truly due, otherwise we can skip
				dueUnits := []int{} // matches length of dueAbilities
				dueFuncs := []int{}
				if CheckValidFunction(incrementedState, ie.UnitIdx, incrementedState.Functions[ie.FuncIdx]) {
					dueFuncs = append(dueFuncs, ie.FuncIdx)
					dueUnits = append(dueUnits, ie.UnitIdx)
				} else {
					continue
				}

				// check for other functions that are due
				for i := 0; i < len(incrementedState.FunctionStartIdx); i++ {
					for j := 0; j < incrementedState.FunctionCount[i]; j++ {
						funcIdx := incrementedState.FunctionStartIdx[i] + j
						if incrementedState.FunctionType[funcIdx] == 0 { // only check auto functions
							isValid := CheckValidFunction(incrementedState, i, incrementedState.Functions[funcIdx])
							if incrementedState.FunctionCharge[funcIdx] >= incrementedState.FunctionSPD[funcIdx] && isValid {
								dueFuncs = append(dueFuncs, funcIdx)
								dueUnits = append(dueUnits, i)
							}
						}
					}
				}

				// activate abilities
				abilityState, modified := HandleDueFunctions(incrementedState, dueUnits, dueFuncs)
				if modified {
					match.GameStates = append(match.GameStates, abilityState)
					match.BroadcastGameState(abilityState)
				}

				// schedule next AutoScheduler
				nextAutoEventTime, success := CalcNextAutoEvent(abilityState)
				if success {
					pushAutoScheduler(match, nextAutoEventTime)
				}

			} else if ie.Type == "Target" {
				// can change this to an event instead of full gamestate
				modifiedState, modified := target.SetTargets(incrementedState, nextTimestamp, ie.UnitIdx, ie.FuncIdx, ie.Target1, ie.Target2)
				if modified {
					match.GameStates = append(match.GameStates, modifiedState)
					match.BroadcastGameState(modifiedState)
				}
				// schedule autoscheduler
				nextAutoEventTime, success := CalcNextAutoEvent(modifiedState)
				if success {
					pushAutoScheduler(match, nextAutoEventTime)
				}
			}

		case "NewRound":
			fmt.Println("received NewRound")

			// remove all autoscheduler events
			match.EventPQueue.RemoveAllAutoSchedulers()

			// copy state
			roundTimestamp := next.Event.Timestamp().Add(200 * time.Millisecond)
			newState := engine.CopyGameState(roundTimestamp, match.GameStates[len(match.GameStates)-1])

			// increment the round
			newState.Round += 1
			match.GameStates = append(match.GameStates, newState)
			match.BroadcastGameState(newState)
			pushCurrentEvent(match, roundTimestamp)

			// sim any first round summons (add a summon condition, an enumeration/ioata in the game state)
			prevState := newState
			modified := false
			for i := 0; i < len(newState.Zone); i++ {
				if newState.Zone[i] == 3 {
					modified = true
					prevState = summon.BasicSummon(match.UnitModels, prevState, roundTimestamp, newState.Team[i], i)
				}
			}
			if modified {
				match.GameStates = append(match.GameStates, prevState)
				match.BroadcastGameState(prevState)
				pushCurrentEvent(match, roundTimestamp)
			}

			// schedule autoscheduler
			nextAutoEventTime, success := CalcNextAutoEvent(prevState)
			if success {
				pushAutoScheduler(match, nextAutoEventTime)
			}
		}

	}

	// set timer for next event or set it to nil
	match.setPQTimer()

}

func PushNewRound(match *Match) {
	match.NextSeq++
	newRound := &NewRoundEvent{
		T: match.NextRoundTime.Add(-200 * time.Millisecond),
	}
	item := &EventItem{
		Event: newRound,
		Type:  "NewRound",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
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

func pushAutoScheduler(match *Match, nextAutoEvent time.Time) {
	match.NextSeq++
	currentEvent := &CurrentEvent{
		T: nextAutoEvent.Add(-200 * time.Millisecond), // HARDCODED to schedule autoschedules 200 ms earlier
	}
	item := &EventItem{
		Event: currentEvent,
		Type:  "AutoScheduler",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

func IncrementState(state *engine.GameState, nextTimestamp time.Time) *engine.GameState {

	newState := engine.CopyGameState(nextTimestamp, state)
	increment := int(nextTimestamp.UnixMilli() - state.Timestamp.UnixMilli())

	for i := 0; i < len(state.Zone); i++ {

		// increment the on-board functions
		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.FunctionStartIdx[i]; j < state.FunctionStartIdx[i]+state.FunctionCount[i]; j++ {
				if state.FunctionType[j] == 0 {
					if state.FunctionSPD[j]-increment < state.FunctionCharge[j] { // we've reached past the charge
						newState.FunctionCharge[j] = newState.FunctionSPD[j]
					} else {
						newState.FunctionCharge[j] += increment
					}
				}
			}
		}

		// increment the units being summoned
		if state.HP[i] > 0 && state.Zone[i] == 3 {
			state.SummonCharge[i] += increment
		}

	}

	return newState
}

func CalcNextAutoEvent(state *engine.GameState) (time.Time, bool) {

	timeToAutoEvent := math.MaxInt // in milliseconds

	for i := 0; i < len(state.Zone); i++ {

		// checking for summon resolutions
		if state.HP[i] > 0 && state.Zone[i] == 3 { // presummon zone
			timeLeft := 2000 - state.SummonCharge[i] // HARDCODED to 2 seconds
			if timeLeft < timeToAutoEvent {
				timeToAutoEvent = timeLeft
			}
		}

		// iterating through functions
		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.FunctionStartIdx[i]; j < state.FunctionStartIdx[i]+state.FunctionCount[i]; j++ {
				if state.FunctionType[j] == 0 { // is an auto function
					timeLeft := int(state.FunctionSPD[j] - state.FunctionCharge[j])
					if timeLeft < timeToAutoEvent && simulate.ValidFunction(state, i, state.Functions[j]) {
						timeToAutoEvent = timeLeft
					}
				}
			}
		}

	}

	if timeToAutoEvent == math.MaxInt {
		return time.Time{}, false
	}

	return state.Timestamp.Add(time.Millisecond * time.Duration(timeToAutoEvent)), true
}

func (match *Match) setPQTimer() {

	// clear timer
	if match.EventPQTimer != nil {
		if !match.EventPQTimer.Stop() {
			select {
			case <-match.EventPQTimer.C: // drain the channel
			default:
			}
		}
	}

	// if no events, return
	first, ok := match.EventPQueue.Peek()
	if !ok {
		match.EventPQTimer = nil
		return
	}

	// set timer
	now := time.Now()
	fireTime := first.Event.Timestamp()
	delay := fireTime.Sub(now)
	if delay < 0 {
		delay = 0
	}
	match.EventPQTimer = time.NewTimer(delay)

}
