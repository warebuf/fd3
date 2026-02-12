package match

import (
	"container/heap"
	"fmt"
	"math"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions"
	"github.com/warebuf/fd3/engine/summon"
	"github.com/warebuf/fd3/engine/target"
	"github.com/warebuf/fd3/utility/sort"
)

// so we have to balance between client inputs and autoevents
// every client input allows us to predict the next n+1 autoevent
// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// this one loop handles pushing inputs into a priority queue
// and taking out Events, Inputs, and AutoSchedulers from it

func (match *Match) EventScheduler() {

	// on init
	match.SetCurrentState()

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
			autoEventTime := next.Event.Timestamp().Add(200 * time.Millisecond)
			incrementedState := IncrementState(&match.GameState, autoEventTime)

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
				pushCurrentEvent(match, autoEventTime)
			}

			// schedule next AutoScheduler
			nextAutoEventTime, success := CalcNextAutoEvent(abilityState)
			if success {
				pushAutoScheduler(match, nextAutoEventTime)
			}

		case "Input":
			fmt.Println("received Input")
			ie := next.Event.(*InputEvent)
			fmt.Println("ie", ie)

			// TODO: notably there is little to no validation

			// remove all autoscheduler events (at most only 1 exists in current design)
			match.EventPQueue.RemoveAllAutoSchedulers()

			// increment the state from the last confirmed GameState
			nextTimestamp := ie.T.Add(200 * time.Millisecond)
			incrementedState := IncrementState(match.GameStates[len(match.GameStates)-1], nextTimestamp)

			if ie.Type == "Summon" {

				// apply the event
				newState, newActions := summon.BasicSummon(match.UnitModels, incrementedState, nextTimestamp, ie.TeamIdx, ie.UnitIdx)
				fmt.Println("newActions", newActions)

				// set the match game state to newState

				// append the new action / schedule the action resolution

				// share the new action

				// schedule autoscheduler
				nextAutoEventTime, success := CalcNextAutoEvent(newState)
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
					pushCurrentEvent(match, nextTimestamp)
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
					pushCurrentEvent(match, nextTimestamp)
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

func pushCurrentEvent(match *Match, timestamp time.Time) {
	match.NextSeq++
	currentEvent := &CurrentEvent{
		T: timestamp,
	}
	item := &EventItem{
		Event: currentEvent,
		Type:  "Current",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

// AutoScheduler happens 200 ms earlier
func pushAutoScheduler(match *Match, nextAutoEvent time.Time) {
	match.NextSeq++
	currentEvent := &CurrentEvent{
		T: nextAutoEvent.Add(-200 * time.Millisecond),
	}
	item := &EventItem{
		Event: currentEvent,
		Type:  "AutoScheduler",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

// increment currentState to most recent state
func (match *Match) SetCurrentState() {
	for match.CurrentState+1 < len(match.GameStates) && time.Now().After(match.GameStates[match.CurrentState].Timestamp) {
		match.CurrentState += 1
	}
}

func IncrementState(state *engine.GameState, nextTimestamp time.Time) *engine.GameState {

	newState := engine.CopyGameState(nextTimestamp, state)
	increment := int(nextTimestamp.UnixMilli() - state.Timestamp.UnixMilli())

	for i := 0; i < len(state.Zone); i++ {
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
	}

	return newState
}

// if we return a negative number, no autoevent found
func CalcNextAutoEvent(state *engine.GameState) (time.Time, bool) {

	timeToAutoEvent := math.MaxInt

	for i := 0; i < len(state.Zone); i++ {
		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.FunctionStartIdx[i]; j < state.FunctionStartIdx[i]+state.FunctionCount[i]; j++ {
				if state.FunctionType[j] == 0 { // is an auto function
					timeLeft := int(state.FunctionSPD[j] - state.FunctionCharge[j])
					if timeLeft < timeToAutoEvent && CheckValidFunction(state, i, state.Functions[j]) {
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

func HandleDueFunctions(incrementedState *engine.GameState, dueUnits []int, dueAbilities []int) (*engine.GameState, bool) {

	newState := engine.CopyGameState(incrementedState.Timestamp, incrementedState)
	modified := false

	prio := sortFunctionsByPrio(incrementedState, dueAbilities)

	for i := 0; i < len(prio); i++ {

		idx := prio[i]

		dueUnit := dueUnits[idx]
		dueAbility := dueAbilities[idx]
		abilityNo := newState.Functions[dueAbility]

		// could use an array of functions but will use a switch statement for now
		switch abilityNo {
		case 0:
			occurred := functions.A00000(newState, dueUnit)
			if occurred {
				modified = true
				newState.FunctionCharge[dueAbility] = 0
			}
		case 1:
			occurred := functions.A00001(newState, dueUnit)
			if occurred {
				modified = true
				newState.FunctionCharge[dueAbility] = 0
				defenders := []int{}
				attackers := []int{}
				functions := []int{}
				for i := 0; i < len(newState.Zone); i++ {
					if newState.Zone[i] == 1 {
						defenders = append(defenders, i)
						attackers = append(attackers, dueUnit)
						functions = append(functions, dueAbility)
					}
				}
				ApplyCounterChain(newState, defenders, attackers, functions, 2)
			}
		case 2:
			occurred := functions.A00002(newState, dueUnit)
			if occurred {
				newState.FunctionCharge[dueAbility] = 0
				modified = true
			}
		case 3:
			targetIdx := functions.A00003(newState, dueUnit)
			if targetIdx >= 0 {
				modified = true
				newState.FunctionCharge[dueAbility] = 0
				defenders := []int{targetIdx}
				attackers := []int{dueUnit}
				functions := []int{dueAbility}
				ApplyCounterChain(newState, defenders, attackers, functions, 50)
			}
		case 4:
			fmt.Println("cannot handle counterattack here, there is a bug")
		case 5:
			targetIdx := functions.A00005(newState, dueUnit)
			if targetIdx >= 0 {
				modified = true
				newState.FunctionCharge[dueAbility] = 0
				defenders := []int{targetIdx}
				attackers := []int{dueUnit}
				functions := []int{dueAbility}
				ApplyCounterChain(newState, defenders, attackers, functions, 50)
			}
		case 6:
			targetIdx := functions.A00006(newState, dueUnit)
			if targetIdx >= 0 {
				modified = true
				newState.FunctionCharge[dueAbility] = 0
				defenders := []int{targetIdx}
				attackers := []int{dueUnit}
				functions := []int{dueAbility}
				ApplyCounterChain(newState, defenders, attackers, functions, 100)
			}
		}

	}

	// mark dead as dead
	for i := 0; i < len(newState.HP); i++ {
		if newState.HP[i] <= 0 {
			newState.Zone[i] = 2
		}
	}

	// iterate through targets, if dead, reset target
	for i := 0; i < len(newState.FunctionTarget); i++ {
		if newState.FunctionTarget[i] >= 0 {
			if newState.Zone[newState.FunctionTarget[i]] == 2 {
				newState.FunctionTarget[i] = -1
			}
		}
	}

	return newState, modified
}

func sortFunctionsByPrio(state *engine.GameState, dueFunctions []int) []int {

	funcNo := make([]int, len(dueFunctions))
	for i := 0; i < len(dueFunctions); i++ {
		funcNo[i] = state.Functions[dueFunctions[i]]
	}

	prio := make([]int, len(funcNo))

	// heal has prio over damage
	for i := 0; i < len(funcNo); i++ {
		switch funcNo[i] {
		case 0:
			prio[i] = 0
		case 1:
			prio[i] = 1
		case 2:
			prio[i] = 0
		case 3:
			prio[i] = 1
		case 4: // counter, should not be sorted because cannot be activated by itself
			prio[i] = 2
		case 5:
			prio[i] = 1
		case 6:
			prio[i] = 1
		}
	}

	sorted := sort.SortedIndexes(prio)

	return sorted

}

func CheckValidFunction(state *engine.GameState, casterIdx int, funcNo int) bool {
	switch funcNo {
	case 0:
		return functions.A00000Valid(state, casterIdx)
	case 1:
		return functions.A00001Valid(state, casterIdx)
	case 2:
		return functions.A00002Valid(state, casterIdx)
	case 3:
		return functions.A00003Valid(state, casterIdx)
	case 4: // counter, should never be valid to be initiated by itself
		return false
	case 5:
		return functions.A00005Valid(state, casterIdx)
	case 6:
		return functions.A00006Valid(state, casterIdx)
	}
	return false
}

// three types of implementations
// (1) reflect and mult, (2) cancel and return a fixed amount, (3) reflect only up to a certain amount
// actually, you can think of a bunch more, reflect and boost, reflect to a teammate, etc
func ApplyCounterChain(state *engine.GameState, defenderArr []int, attackerArr []int, functionArr []int, dmg int) {

	currentDmg := dmg

	for j := 0; j < len(defenderArr); j++ {

		currDefender := defenderArr[j]
		currAttacker := attackerArr[j]
		currFunction := functionArr[j]

		success := true

		// check all abilities for this unit
		for i := state.FunctionStartIdx[currDefender]; i < state.FunctionStartIdx[currDefender]+state.FunctionCount[currDefender]; i++ {
			switch state.Functions[i] {
			case 4: // reflect and mult 1x
				if functions.A00004CounterValid(state, i) {
					currentDmg = functions.A00004(state, currDefender, currAttacker, currentDmg)
					state.FunctionCharge[i] = 0
					defenderArr = append(defenderArr, currAttacker)
					attackerArr = append(attackerArr, currDefender)
					functionArr = append(functionArr, i)
					success = false
				}
			default:
			}
		}

		//
		if success {
			HandlePostEffect(state, defenderArr, attackerArr, functionArr, currAttacker, currDefender, currFunction)
		}

	}

}

func HandlePostEffect(state *engine.GameState, defenderArr []int, attackerArr []int, functionArr []int, attackerIdx int, defenderIdx int, functionIdx int) {

	functionNo := state.Functions[functionIdx]

	switch functionNo {
	case 5:
		functions.A00005Success(state, attackerIdx)
	case 6:
		functions.A00005Success(state, attackerIdx)
	default:
	}

}
