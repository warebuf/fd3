package match

import (
	"container/heap"
	"fmt"
	"math"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilities"
	"github.com/warebuf/fd3/engine/summon"
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

			//increment
			autoEventTime := next.Event.Timestamp().Add(200 * time.Millisecond)
			incrementedState := IncrementState(match.GameStates[len(match.GameStates)-1], autoEventTime)

			// figure out which abilities are due
			dueUnits := []int{} // matches length of dueAbilities
			dueAbilities := []int{}
			for i := 0; i < len(incrementedState.AbilityStartIdx); i++ {
				for j := 0; j < incrementedState.AbilityCount[i]; j++ {
					abilityIdx := incrementedState.AbilityStartIdx[i] + j
					isValid := CheckValidAbility(incrementedState, i, incrementedState.Abilities[abilityIdx])
					if incrementedState.AbilityCharge[abilityIdx] >= incrementedState.AbilitySPD[abilityIdx] && isValid {
						dueAbilities = append(dueAbilities, abilityIdx)
						dueUnits = append(dueUnits, i)
					}
				}
			}

			// activate abilities
			abilityState, modified := HandleDueAbilities(incrementedState, dueUnits, dueAbilities)
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

		case "Current":
			match.CurrentState += 1
			fmt.Println("moved to Current State", match.CurrentState, "/", len(match.GameStates)-1)

		case "Input":
			fmt.Println("received Input")
			ie := next.Event.(*InputEvent)

			// remove all autoscheduler events (at most only 1 exists in current design)
			match.EventPQueue.RemoveFirstAutoScheduler()

			// increment the state from the last confirmed GameState
			nextTimestamp := ie.T.Add(200 * time.Millisecond)
			incrementedState := IncrementState(match.GameStates[len(match.GameStates)-1], nextTimestamp)

			// apply the event and create a new GameState (TODO: broadcast an event instead of a full GameState)
			newState := summon.BasicSummon(match.UnitModels, incrementedState, nextTimestamp, ie.TeamIdx, ie.UnitIdx)
			match.GameStates = append(match.GameStates, newState)
			match.BroadcastGameState(newState)
			pushCurrentEvent(match, nextTimestamp) // add the currentEvent to the PQ

			// schedule autoscheduler
			nextAutoEventTime, success := CalcNextAutoEvent(newState)
			if success {
				pushAutoScheduler(match, nextAutoEventTime)
			}

		case "NewRound":
			fmt.Println("received NewRound")

			// remove all autoscheduler events (at most only 1 exists in current design)
			match.EventPQueue.RemoveFirstAutoScheduler()

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
				if newState.Zone[i] == 0 && newState.SummonCondition[i] == engine.SummonTypeStart {
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
	increment := nextTimestamp.UnixMilli() - state.Timestamp.UnixMilli()

	for i := 0; i < len(state.Zone); i++ {
		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.AbilityStartIdx[i]; j < state.AbilityStartIdx[i]+state.AbilityCount[i]; j++ {

				if state.AbilitySPD[j]-increment < state.AbilityCharge[j] { // we've reached past the charge
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
func CalcNextAutoEvent(state *engine.GameState) (time.Time, bool) {

	timeToAutoEvent := math.MaxInt

	for i := 0; i < len(state.Zone); i++ {

		if state.HP[i] > 0 && state.Zone[i] == 1 {
			for j := state.AbilityStartIdx[i]; j < state.AbilityStartIdx[i]+state.AbilityCount[i]; j++ {
				timeLeft := int(state.AbilitySPD[j] - state.AbilityCharge[j])
				if timeLeft < timeToAutoEvent && CheckValidAbility(state, i, state.Abilities[j]) {
					timeToAutoEvent = timeLeft
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

func HandleDueAbilities(incrementedState *engine.GameState, dueUnits []int, dueAbilities []int) (*engine.GameState, bool) {

	newState := engine.CopyGameState(incrementedState.Timestamp, incrementedState)
	modified := false

	prio := sortAbilitiesByPrio(incrementedState, dueAbilities)

	for i := 0; i < len(prio); i++ {

		idx := prio[i]

		dueUnit := dueUnits[idx]
		dueAbility := dueAbilities[idx]
		abilityNo := newState.Abilities[dueAbility]

		// could use an array of functions but will use a switch statement for now
		switch abilityNo {
		case 0:
			occurred := abilities.A00000(newState, dueUnit)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 1:
			occurred := abilities.A00001(newState)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 2:
			occurred := abilities.A00002(newState, dueUnit)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 3:
			occurred := abilities.A00003(newState, dueUnit)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 4:
			occurred := abilities.A00004(newState, dueUnit)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		case 5:
			occurred := abilities.A00005(newState, dueUnit)
			if occurred {
				modified = true
				newState.AbilityCharge[dueAbility] = 0
			}
		}

	}

	// mark dead as dead
	for i := 0; i < len(newState.HP); i++ {
		if newState.HP[i] <= 0 {
			newState.Zone[i] = 2
		}
	}

	return newState, modified
}

func sortAbilitiesByPrio(state *engine.GameState, dueAbilities []int) []int {

	abilityNo := make([]int, len(dueAbilities))
	for i := 0; i < len(dueAbilities); i++ {
		abilityNo[i] = state.Abilities[dueAbilities[i]]
	}

	prio := make([]int, len(abilityNo))

	// heal has prio over damage
	for i := 0; i < len(abilityNo); i++ {
		switch abilityNo[i] {
		case 0:
			prio[i] = 0
		case 1:
			prio[i] = 1
		case 2:
			prio[i] = 1
		case 3:
			prio[i] = 0
		case 4:
			prio[i] = 0
		case 5:
			prio[i] = 1
		}
	}

	sorted := sort.SortedIndexes(prio)

	return sorted

}

func CheckValidAbility(state *engine.GameState, casterIdx int, abilityNo int) bool {
	switch abilityNo {
	case 0:
		return abilities.A00000Valid(state, casterIdx)
	case 1:
		return abilities.A00001Valid(state, casterIdx)
	case 2:
		return abilities.A00002Valid(state, casterIdx)
	case 3:
		return abilities.A00003Valid(state, casterIdx)
	case 4:
		return abilities.A00004Valid(state, casterIdx)
	case 5:
		return abilities.A00005Valid(state, casterIdx)
	}
	return false
}
