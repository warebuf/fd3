package match

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/randomFunctions"
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
		if time.Now().Before(next.Event.Timestamp()) {
			break
		}

		heap.Pop(&match.EventPQueue)

		switch next.Type {
		case "Effect":
			match.handleEffectEvent(next.Event.(*eventpq.EffectEvent))
		case "Input":
			match.handleInputEvent(next.Event.(*eventpq.InputEvent))
		case "NewPhase":
			match.handleNewPhaseEvent(next.Event.(*eventpq.NewPhaseEvent))
		case "PauseEndEvent":
			pauseEndEvent := next.Event.(*eventpq.PauseEndEvent)
			match.TransitionPauseState(-1, pauseEndEvent.T.UnixMilli())
		}

	}

	match.setPQTimer() // set timer for next event or set it to nil

}

func (match *Match) handleEffectEvent(effEvent *eventpq.EffectEvent) {

	// simulate
	oldLogsSize := len(match.EventLog)
	randomFunctions.ManageSimulationPQ(match.GameState, &match.EventLog, &match.NextUpdateID)

	// update the published state and publish the new logs
	match.publish(effEvent.T.UnixMilli()+constants.LEAD, oldLogsSize)

}

// publish updates WallTimestamp, snapshots GameState into PublishedState, and broadcasts new logs.
func (match *Match) publish(wallTime int64, oldLogsSize int) {
	match.GameState.WallTimestamp = wallTime
	match.GameState.UpdateID = match.NextUpdateID
	match.NextUpdateID += 1
	match.PublishedState = state.DeepCopyGameState(match.GameState)
	match.BroadcastLogs(oldLogsSize)
}

func PushNewPhase(match *Match, t int64, phase int) {
	newPhase := &eventpq.NewPhaseEvent{
		T:     time.UnixMilli(t - constants.LEAD), // schedule 200 ms early
		Phase: phase,
	}
	match.EventPQueue.Enqueue(newPhase, "NewPhase")
}

func pushInputEvent(match *Match, ev *eventpq.InputEvent) {
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

func scheduleEarliestSimItem(match *Match) {

	// get the timestamp of earliest in the SimulationPQ
	earliestSimItem := match.GameState.SimulationPQ.Peek()

	// if no earliest item, just clean everything
	if earliestSimItem == nil {
		if match.EarliestSimItem != nil && match.EarliestSimItem.Index > 0 { // EarliestEffect exists AND is in EventPQ (indicated by index > 0)
			match.EventPQueue.Remove(match.EarliestSimItem)
		}
		match.EarliestSimItem = nil
		return
	}

	// popTime is denominated in sim time, so we convert sim time to wall time by adding accumulated pause offset
	earliest := time.UnixMilli(earliestSimItem.PopTime + match.GameState.SimPauseOffset)

	// if the simItem exists and is in the EventPQ, update and reheapify in EventPQ
	if match.EarliestSimItem != nil && match.EarliestSimItem.Index > 0 {
		// update and reheapify
		effEv, _ := match.EarliestSimItem.Event.(*eventpq.EffectEvent)
		effEv.T = earliest
		match.EventPQueue.Fix(match.EarliestSimItem)
		return
	}

	// otherwise schedule the simItem in the EventPQ
	ev := &eventpq.EffectEvent{T: earliest}
	match.EarliestSimItem = match.EventPQueue.Enqueue(ev, "Effect")

}
