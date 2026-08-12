package match

import (
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func SchedulePhaseEvent(match *Match, activeTime int64, phase int) {

	wallTime := time.UnixMilli(activeTime + match.GameState.FullPauseOffset - constants.LEAD)
	match.PendingPhase = phase
	match.PendingPhaseTime = activeTime

	// if a phase event is already pending, update and reheapify it in place
	if match.PhaseItem != nil {
		newPhase, _ := match.PhaseItem.Event.(*eventpq.NewPhaseEvent)
		newPhase.WallTime = wallTime
		newPhase.Phase = phase
		match.EventPQueue.Fix(match.PhaseItem)
		return
	}

	// otherwise schedule a new phase event in the EventPQ
	newPhase := &eventpq.NewPhaseEvent{WallTime: wallTime, Phase: phase}
	match.PhaseItem = match.EventPQueue.CreateAndPushEventItem(newPhase, "NewPhase")
}

func (match *Match) trySkipStaging(executeAtWallTime int64) {
	gs := match.GameState
	if gs.Phase != state.PhaseDraw {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	for _, ready := range gs.Ready {
		if !ready {
			return
		}
	}

	activeTime := executeAtWallTime - gs.FullPauseOffset
	SchedulePhaseEvent(match, activeTime, state.PhasePlay)
}

func pushInputEvent(match *Match, ev *eventpq.InputEvent) {
	ev.WallTime = time.Now() // relabel the timestamp when putting it into PQ
	match.EventPQueue.CreateAndPushEventItem(ev, "Input")
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

func scheduleChainResolution(match *Match) {

	// if no chain items, clean the pointer
	if len(match.GameState.Chain) == 0 {
		if match.LatestChainItem != nil {
			match.EventPQueue.Remove(match.LatestChainItem)
			match.LatestChainItem = nil
		}
		return
	}

	// get the top chain item
	topChainItem := match.GameState.Chain[len(match.GameState.Chain)-1]

	// pop time is denominated in active time
	earliest := time.UnixMilli(topChainItem.PopAtActiveTime + match.GameState.FullPauseOffset)

	// if the chain item exists and is in the EventPQ, update and reheapify
	if match.LatestChainItem != nil {
		effEv, _ := match.LatestChainItem.Event.(*eventpq.ChainResolutionEvent)
		effEv.WallTime = earliest
		match.EventPQueue.Fix(match.LatestChainItem)
		return
	}

	// otherwise add it to EventPQ
	ev := &eventpq.ChainResolutionEvent{WallTime: earliest}
	match.LatestChainItem = match.EventPQueue.CreateAndPushEventItem(ev, "ChainResolution")
}
