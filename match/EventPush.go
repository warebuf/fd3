package match

import (
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

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

func scheduleReactionStackResolution(match *Match) {

	// if no stack items, clean the pointer
	if len(match.GameState.ReactionStack) == 0 {
		if match.LatestReactionItem != nil {
			match.EventPQueue.Remove(match.LatestReactionItem)
			match.LatestReactionItem = nil
		}
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
