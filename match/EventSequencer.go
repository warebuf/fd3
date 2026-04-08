package match

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/simulate"
	"github.com/warebuf/fd3/engine/simulate/microevents"
	"github.com/warebuf/fd3/engine/state"
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
		case "Effect":
			oldLogsSize := len(match.EventLog)
			simulate.ManageStackPQ(match.GameState, match.PQ, &match.EventLog)
			match.BroadcastLogs(oldLogsSize)
		case "Input":
			ie := next.Event.(*InputEvent)
			fmt.Println("received Input", ie)

			// TODO: notably there is little to no validation

			// TODO: if phase is not 2, do not accept inputs (not sure whether to do it here or before)

			// remove all autoscheduler events (at most only 1 exists in current design)
			match.EventPQueue.RemoveAllAutoSchedulers()

			// increment the state from the last confirmed GameState
			nextTimestamp := ie.T.Add(200 * time.Millisecond)
			incrementedState := IncrementState(match.GameState, nextTimestamp)

			if ie.Type == "Summon" {

				// add the summon to the priority queue and modify the game state
				newEvents, success := simulate.AddSummonToPQ(match.GameState, match.PQ, ie.UnitIdx, nextTimestamp, match.NextEffectIndex)
				if success {
					// increment the effect counter
					match.NextEffectIndex += 1

					// publish the new state
					match.PublishedState = engine.DeepCopyGameState(match.GameState.Timestamp, match.GameState)

					// append the new actions
					match.EventLog = append(match.EventLog, newEvents...)

					// share the new action
					if len(newEvents) > 0 {
						match.BroadcastEventLog(newEvents)
					}
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

		case "NewPhase":

			newPhase := next.Event.(*NewPhaseEvent)
			fmt.Println("received newPhase", newPhase)

			if newPhase.Phase == 1 { // entering draw phase

				// saving the old log count for broadcasting the new logs
				oldLogsSize := len(match.EventLog)

				// increment the phase
				match.GameState.Phase = 1
				match.EventLog = append(match.EventLog, logs.Log{
					Timestamp: time.Now().UnixMilli(),
					Type:      logs.PhaseLog,
					Value:     1,
				})

				// schedule the draw event
				currentTime := time.Now().UnixMilli()
				microevents.ScheduleDrawCard(match.PQ, match.NextPQIndex, 0, currentTime+5000) // 5.0 seconds from now
				match.NextPQIndex += 1
				microevents.ScheduleDrawCard(match.PQ, match.NextPQIndex, 0, currentTime+5500) // 5.5 seconds from now
				match.NextPQIndex += 1
				microevents.ScheduleDrawCard(match.PQ, match.NextPQIndex, 0, currentTime+6000) // 6.0 seconds from now
				match.NextPQIndex += 1
				microevents.ScheduleDrawCard(match.PQ, match.NextPQIndex, 1, currentTime+6500) // 6.5 seconds from now
				match.NextPQIndex += 1
				microevents.ScheduleDrawCard(match.PQ, match.NextPQIndex, 1, currentTime+7000) // 7.0 seconds from now
				match.NextPQIndex += 1
				microevents.ScheduleDrawCard(match.PQ, match.NextPQIndex, 1, currentTime+7500) // 7.5 seconds from now
				match.NextPQIndex += 1

				// schedule phase 2 10s from now
				PushNewPhase(match, time.UnixMilli(currentTime).Add(10*time.Second), 2)

				// update the game state, published state
				match.GameState.GameStateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				// broadcast new logs
				match.BroadcastLogs(oldLogsSize)

			} else if newPhase.Phase == 2 { // entering game start
				// increment the phase (these can be broadcasted early)
				// update the game state, published state
				// schedule EventPQ effects
				// broadcast new logs
			} else if newPhase.Phase == 3 { // entering game over
				// end game (not sure what this requires tbh)
				// update the game state, published state
				// broadcast new logs
			}
		}

	}

	scheduleStackPQ(match) // set the event for the next stackPQ item
	match.setPQTimer()     // set timer for next event or set it to nil

}

func PushNewPhase(match *Match, t time.Time, phase int) {
	match.NextSeq++
	newPhase := &NewPhaseEvent{
		T:     t.Add(-200 * time.Millisecond), // needs to be scheduled 200 ms early
		Phase: phase,
	}
	item := &EventItem{
		Event: newPhase,
		Type:  "NewPhase",
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

func scheduleStackPQ(match *Match) {

	// get the timestamp of earliest in StackPQ
	earliestStack := match.PQ.Peek()

	// if no effect events, remove the EffectEvent from the EventPQ
	if earliestStack == nil {
		if match.EarliestEffect != nil {
			heap.Remove(&match.EventPQueue, match.EarliestEffect.Index)
			match.EarliestEffect = nil
		}
		return
	}

	// if EffectEvent exists, update and reheapify in EventPQ
	earliest := time.UnixMilli(earliestStack.ResolveTime)
	if match.EarliestEffect != nil {
		effEv, ok := match.EarliestEffect.Event.(*EffectEvent)
		if !ok {
			return
		}
		effEv.T = earliest
		heap.Fix(&match.EventPQueue, match.EarliestEffect.Index)
		return
	}

	// otherwise add EffectEvent to EventPQ
	ev := &EffectEvent{
		T: earliest,
	}
	item := &EventItem{
		Event: ev,
		Type:  "Effect",
		seq:   int64(match.NextSeq),
	}
	match.NextSeq++
	heap.Push(&match.EventPQueue, item)
	match.EarliestEffect = item

}

func handleDueStackPQ(match *Match) {

}
