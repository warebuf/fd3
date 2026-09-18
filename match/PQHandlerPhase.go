package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleNewPhaseEvent(newPhase *eventpq.NewPhaseEvent) {

	// skip stale phase events (e.g. draw phase timer after early ready-up)
	if match.GameState.Phase >= newPhase.Phase {
		return
	}

	match.PendingPhase = 0
	match.PendingPhaseTime = 0

	switch newPhase.Phase {
	case state.PhaseDraw:
		match.handleEnterDrawPhase(newPhase)
	case state.PhasePlay:
		match.handleEnterPlayPhase(newPhase)
	case state.PhaseGameOver:
		match.handleEnterGameOverPhase(newPhase)
	}

}

// entering draw phase, we are in pause phase 2 here since we want to draw cards
func (match *Match) handleEnterDrawPhase(newPhase *eventpq.NewPhaseEvent) {

	oldLogsSize := len(match.EventLog)

	// change the phase
	phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
	drawEndActiveTime := phaseChangeWallTime - match.GameState.FullPauseOffset + constants.DRAWS_COMPLETE_AT + constants.STAGING_WINDOW
	match.GameState.Phase = state.PhaseDraw
	match.GameState.PhaseEndInActiveTime = drawEndActiveTime
	match.EventLog = append(match.EventLog, logs.PhaseLog{
		UpdateID:             match.NextUpdateID,
		WallTimestamp:        phaseChangeWallTime,
		Type:                 logs.Phase,
		Phase:                state.PhaseDraw,
		PhaseEndInActiveTime: drawEndActiveTime,
	})
	match.NextUpdateID += 1

	// initialize ready flags; both teams readying up skips straight to play
	numTeams := len(match.GameState.Hand)
	match.GameState.Ready = make([]bool, numTeams)

	// schedule draws: 5 cards per team, staggered 250ms apart
	for i := 0; i < 5; i++ {
		schedule.ScheduleSystemDraw(match.GameState, &match.SimulationPQ, 0, phaseChangeWallTime+500+int64(i)*250)
	}
	for i := 0; i < 5; i++ {
		schedule.ScheduleSystemDraw(match.GameState, &match.SimulationPQ, 1, phaseChangeWallTime+2000+int64(i)*250)
	}

	// schedule play phase after draws complete
	SchedulePhaseEvent(match, drawEndActiveTime, state.PhasePlay)

	match.publish(phaseChangeWallTime, oldLogsSize)

	// schedule draw cards
	scheduleEarliestSimItem(match)
}

// entering play
func (match *Match) handleEnterPlayPhase(newPhase *eventpq.NewPhaseEvent) {

	oldLogsSize := len(match.EventLog)

	phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
	match.GameState.Phase = state.PhasePlay
	match.GameState.PhaseEndInActiveTime = 0
	match.EventLog = append(match.EventLog, logs.PhaseLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: phaseChangeWallTime,
		Type:          logs.Phase,
		Phase:         state.PhasePlay,
	})
	match.NextUpdateID += 1

	match.TransitionPauseState(state.PauseTransitionUnpause, phaseChangeWallTime, -1)

	// begin mana regen for structures
	startSimTime := phaseChangeWallTime - match.GameState.SimPauseOffset
	for team := range match.GameState.Structures {
		for i := range match.GameState.Structures[team] {
			schedule.ScheduleStructureManaRegen(match.GameState, &match.SimulationPQ, team, i, startSimTime)
		}
	}

	// apply any available auto plays
	match.processAutoSummonQueue(phaseChangeWallTime, -1)

	scheduleEarliestSimItem(match)

	match.publish(phaseChangeWallTime, oldLogsSize)
}

// entering game over
func (match *Match) handleEnterGameOverPhase(newPhase *eventpq.NewPhaseEvent) {

	oldLogsSize := len(match.EventLog)
	phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
	match.GameState.Phase = state.PhaseGameOver
	match.GameState.PhaseEndInActiveTime = 0

	deadCount := 0
	for team := range match.GameState.Structures {
		if match.GameState.AllStructuresDead(team) {
			deadCount++
		}
	}
	if deadCount >= 2 {
		match.GameState.Winner = -2
	}

	match.EventLog = append(match.EventLog, logs.GameOverLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: phaseChangeWallTime,
		Type:          logs.GameOver,
		Winner:        match.GameState.Winner,
	})
	match.NextUpdateID += 1

	// cancel all pending scheduled events
	if match.EarliestSimItem != nil {
		match.EventPQueue.Remove(match.EarliestSimItem)
		match.EarliestSimItem = nil
	}
	if match.LatestChainItem != nil {
		match.EventPQueue.Remove(match.LatestChainItem)
		match.LatestChainItem = nil
	}
	if match.PhaseItem != nil {
		match.EventPQueue.Remove(match.PhaseItem)
		match.PhaseItem = nil
	}
	if match.PauseEndItem != nil {
		match.EventPQueue.Remove(match.PauseEndItem)
		match.PauseEndItem = nil
	}

	match.publish(phaseChangeWallTime, oldLogsSize)
}
