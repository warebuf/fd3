package match

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/debug"
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
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
		if time.Now().Before(next.Event.GetWallTime()) {
			break
		}

		heap.Pop(&match.EventPQueue)

		switch next.Type {
		case "SimPQ":
			match.EarliestSimItem = nil
			match.handleSimPQEvent(next.Event.(*eventpq.SimPQEvent))
		case "StackResolution":
			match.LatestReactionItem = nil
			match.handleRStackResolution()
		case "Input":
			match.handleInputEvent(next.Event.(*eventpq.InputEvent))
		case "NewPhase":
			match.PhaseItem = nil
			match.handleNewPhaseEvent(next.Event.(*eventpq.NewPhaseEvent))
		case "PauseEndEvent":
			match.PauseEndItem = nil
			match.handlePauseEndEvent(next.Event.(*eventpq.PauseEndEvent))
		case "Reveal":
			match.RevealItem = nil
			match.handleRevealEvent(next.Event.(*eventpq.RevealEvent))
		}

	}

	match.setPQTimer() // set timer for next event or set it to nil

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
	fireTime := first.Event.GetWallTime()
	delay := fireTime.Sub(now)
	if delay < 0 {
		delay = 0
	}
	match.EventPQTimer = time.NewTimer(delay)

}

func (match *Match) handleSimPQEvent(simPQEvent *eventpq.SimPQEvent) {
	if match.GameState.Phase == state.PhaseGameOver {
		return
	}
	oldLogsSize := len(match.EventLog)
	wallTime := simPQEvent.WallTime.UnixMilli() + constants.LEAD
	sim.ManageSimulationPQ(match.GameState, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID)
	if (match.GameState.Winner >= 0 || match.GameState.Winner == -2) && match.GameState.Phase != state.PhaseGameOver {
		PushPhaseEvent(match, wallTime-match.GameState.FullPauseOffset, state.PhaseGameOver)
	} else {
		if len(match.GameState.ReactionStack) > 0 {
			match.TransitionPauseState(1, wallTime, -1)
		}
		scheduleRStackResolution(match)
		match.processAutoSummonQueue(wallTime, -1)
		scheduleEarliestSimItem(match)
	}
	match.publish(wallTime, oldLogsSize)
}

func (match *Match) handlePauseEndEvent(pauseEndEvent *eventpq.PauseEndEvent) {
	if match.GameState.Phase == state.PhaseGameOver {
		return
	}
	wallTime := pauseEndEvent.WallTime.UnixMilli() + constants.LEAD
	if match.GameState.PausePhase == state.PausePhaseFullPause {
		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(-1, wallTime, -1)
		match.processAutoSummonQueue(wallTime, -1)
		match.publish(wallTime, oldLogsSize)
	}
}

// publishes the latest updates
func (match *Match) publish(wallTime int64, oldLogsSize int) {
	if len(match.EventLog) <= oldLogsSize {
		return
	}
	match.GameState.WallTimestamp = wallTime
	match.GameState.UpdateID = match.NextUpdateID - 1
	match.PublishedState = state.DeepCopyGameState(match.GameState)
	stateCopy := state.DeepCopyGameState(match.GameState)
	for _, log := range match.EventLog[oldLogsSize:] {
		match.DebugTrace = append(match.DebugTrace, debug.TraceEntry{Log: log, State: stateCopy})
	}
	//debug.WriteTrace(match.ID, match.DebugTrace)
	match.BroadcastLogs(oldLogsSize)
}

func PushPhaseEvent(match *Match, activeTime int64, phase int) {

	// clear previous phase events
	if match.PhaseItem != nil {
		match.EventPQueue.Remove(match.PhaseItem)
	}

	newPhase := &eventpq.NewPhaseEvent{
		WallTime: time.UnixMilli(activeTime + match.GameState.FullPauseOffset - constants.LEAD),
		Phase:    phase,
	}
	match.PendingPhase = phase
	match.PendingPhaseTime = activeTime
	match.PhaseItem = match.EventPQueue.CreateAndPushEventItem(newPhase, "NewPhase")
}

func pushInputEvent(match *Match, ev *eventpq.InputEvent) {
	ev.WallTime = time.Now() // relabel the timestamp when putting it into PQ
	match.EventPQueue.CreateAndPushEventItem(ev, "Input")
}

func pushRevealEvent(match *Match, activeTime int64) {
	match.PendingRevealTime = activeTime
	ev := &eventpq.RevealEvent{WallTime: time.UnixMilli(activeTime + match.GameState.FullPauseOffset - constants.LEAD)}
	match.RevealItem = match.EventPQueue.CreateAndPushEventItem(ev, "Reveal")
}

func scheduleEarliestSimItem(match *Match) {

	if match.GameState.PausePhase != state.PausePhaseSimulating {
		return
	}

	// get the timestamp of earliest in the SimulationPQ
	earliestSimItem := match.SimulationPQ.Peek()

	// if no earliest item, just clean everything
	if earliestSimItem == nil {
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}
		return
	}

	// popTime is denominated in sim time
	earliest := time.UnixMilli(earliestSimItem.PopAtSimTime + match.GameState.SimPauseOffset)

	// if the simItem exists and is in the EventPQ, update and reheapify
	if match.EarliestSimItem != nil {
		simPQEv, _ := match.EarliestSimItem.Event.(*eventpq.SimPQEvent)
		simPQEv.WallTime = earliest
		match.EventPQueue.Fix(match.EarliestSimItem)
		return
	}

	// otherwise schedule the simItem in the EventPQ
	ev := &eventpq.SimPQEvent{WallTime: earliest}
	match.EarliestSimItem = match.EventPQueue.CreateAndPushEventItem(ev, "SimPQ")

}

func scheduleRStackResolution(match *Match) {

	// if no stack items, clean the pointer
	if len(match.GameState.ReactionStack) == 0 {
		return
	}

	// get the top reaction stack item
	topReactionItem := match.GameState.ReactionStack[len(match.GameState.ReactionStack)-1]

	// pop time is denominated in active time
	earliest := time.UnixMilli(topReactionItem.PopAtActiveTime + match.GameState.FullPauseOffset)

	// if the stack item exists and is in the EventPQ, update and reheapify
	if match.LatestReactionItem != nil {
		effEv, _ := match.LatestReactionItem.Event.(*eventpq.StackResolutionEvent)
		effEv.WallTime = earliest
		match.EventPQueue.Fix(match.LatestReactionItem)
		return
	}

	// otherwise add it to EventPQ
	ev := &eventpq.StackResolutionEvent{WallTime: earliest}
	match.LatestReactionItem = match.EventPQueue.CreateAndPushEventItem(ev, "StackResolution")
}
