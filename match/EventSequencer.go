package match

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/randomFunctions"
	"github.com/warebuf/fd3/engine/state"
)

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

		next := match.EventPQueue.Peek()

		// no more due events
		if time.Now().Before(next.Event.Timestamp()) {
			break
		}

		heap.Pop(&match.EventPQueue)

		// handle the event
		switch next.Type {

		case "Effect":

			// simulate
			oldLogsSize := len(match.EventLog)
			randomFunctions.ManageStackPQ(match.GameState, &match.EventLog, &match.NextUpdateID)

			// update the published state
			match.GameState.Timestamp = next.Event.Timestamp().Add(constants.LEAD * time.Millisecond)
			match.GameState.UpdateID = match.NextUpdateID
			match.NextUpdateID += 1
			match.PublishedState = state.DeepCopyGameState(match.GameState)

			// publish the new logs
			match.BroadcastLogs(oldLogsSize)

		case "Input":
			ie := next.Event.(*InputEvent)

			switch ie.Type {
			case "FullPause":
				// set GameState
			}

		case "NewPhase":

			newPhase := next.Event.(*NewPhaseEvent)

		case "PauseEnd":

		}

	}

	match.setPQTimer() // set timer for next event or set it to nil

}

// if isEnd is false, we are starting a new PauseState
func (match *Match) UpdatePauseState(PauseState int, isEnd bool) {

	// for the pause state we are exiting

	// if the pause state is 3: simulating
	// remove the simulationPQ item from EventPQ
	// there might be more items we remove here? idk

	// if the pause state is 2: simulation pause
	// increment the correct offsets
	// remove the Chain item from EventPQ

	// if the pause state is 1: animation pause
	// increment the correct offsets
	// remove the PauseEnd item to EventPQ if it exists

	// if the pause state is 0: full pause
	// increment the correct offsets
	// remove the PauseEnd item to EventPQ if it exists

	// for the pause state we are entering

	// if the pause state is 3: simulating
	// add the simulationPQ item from EventPQ

	// if the pause state is 2: simulation pause
	// update the pause state related members in game state
	// add the Chain item from EventPQ

	// if the pause state is 1: animation pause
	// update the pause state related members in game state
	// add a PauseEnd item to EventPQ

	// if the pause state is 0: full pause
	// increment the correct offsets
	// add a PauseEnd item to EventPQ

}

// UpdatePauseState handles entering or exiting a pause state.
// isEnd=true: exiting pauseState. isEnd=false: entering pauseState.
// endTime is the wall ms when this pause should auto-end (only used for states 0 and 1).
func (match *Match) UpdatePauseState(pauseState int, isEnd bool, endTime int64) {
	now := time.Now().UnixMilli()
	gs := match.GameState

	if isEnd {
		switch pauseState {
		case 3: // simulating
			if match.EarliestSimItem != nil {
				match.EventPQueue.Remove(match.EarliestSimItem)
				match.EarliestSimItem = nil
			}

		case 2: // sim pause
			if gs.PauseStack[2] != 0 {
				gs.SimPauseOffset += now - gs.PauseStack[2]
				gs.PauseStack[2] = 0
			}
			if match.ChainItem != nil {
				match.EventPQueue.Remove(match.ChainItem)
				match.ChainItem = nil
			}

		case 1: // animation pause
			if gs.PauseStack[1] != 0 {
				gs.FullPauseOffset += now - gs.PauseStack[1]
				gs.PauseStack[1] = 0
			}
			if match.PauseEndItem != nil {
				match.EventPQueue.Remove(match.PauseEndItem)
				match.PauseEndItem = nil
			}

		case 0: // full pause
			if gs.PauseStack[0] != 0 {
				gs.FullPauseOffset += now - gs.PauseStack[0]
				gs.PauseStack[0] = 0
			}
			if match.PauseEndItem != nil {
				match.EventPQueue.Remove(match.PauseEndItem)
				match.PauseEndItem = nil
			}
		}
	} else {
		switch pauseState {
		case 3: // simulating
			scheduleSimulationItem(match)

		case 2: // sim pause
			gs.PauseStack[2] = now
			scheduleChainItem(match)

		case 1: // animation pause
			gs.PauseStack[1] = now
			ev := &PauseEndEvent{T: time.UnixMilli(endTime)}
			match.PauseEndItem = match.EventPQueue.Enqueue(ev, "PauseEnd")

		case 0: // full pause
			gs.PauseStack[0] = now
			ev := &PauseEndEvent{T: time.UnixMilli(endTime)}
			match.PauseEndItem = match.EventPQueue.Enqueue(ev, "PauseEnd")
		}
	}

	match.setPQTimer()
}

func PushNewPhase(match *Match, t int64, phase int) {
	newPhase := &NewPhaseEvent{
		T:     time.UnixMilli(t - constants.LEAD), // schedule 200 ms early
		Phase: phase,
	}
	match.EventPQueue.Enqueue(newPhase, "NewPhase")
}

func pushInputEvent(match *Match, ev *InputEvent) {
	ev.T = time.Now() // relabel the timestamp when putting it into PQ
	match.EventPQueue.Enqueue(ev, "Input")
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
	first := match.EventPQueue.Peek()
	if first == nil {
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

func scheduleSimulationItem(match *Match) {

	// get the timestamp of earliest in StackPQ
	earliestStack := match.GameState.SimulationPQ.Peek()

	// if no effect events, remove the EffectEvent from the EventPQ
	if earliestStack == nil {
		if match.EarliestSimItem != nil && match.EarliestSimItem.Index > 0 { // EarliestEffect exists AND is in EventPQ (indicated by index > 0)
			match.EventPQueue.Remove(match.EarliestSimItem)
		}
		match.EarliestSimItem = nil
		return
	}

	// if EffectEvent exists and is in the EventPQ, update and reheapify in EventPQ
	earliest := time.UnixMilli(earliestStack.PopTime)
	if match.EarliestSimItem != nil && match.EarliestSimItem.Index > 0 {
		// update and reheapify
		effEv, _ := match.EarliestSimItem.Event.(*EffectEvent)
		effEv.T = earliest
		match.EventPQueue.Fix(match.EarliestSimItem)
		return
	}

	// otherwise add EffectEvent to EventPQ
	ev := &EffectEvent{T: earliest}
	match.EarliestSimItem = match.EventPQueue.Enqueue(ev, "Effect")

}
