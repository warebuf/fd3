package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleSimPQEvent(simPQEvent *eventpq.SimPQEvent) {
	if match.GameState.Phase == state.PhaseGameOver {
		return
	}
	oldLogsSize := len(match.EventLog)
	wallTime := simPQEvent.WallTime.UnixMilli() + constants.LEAD
	sim.ManageSimulationPQ(match.GameState, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID)
	if (match.GameState.Winner >= 0 || match.GameState.Winner == -2) && match.GameState.Phase != state.PhaseGameOver {
		SchedulePhaseEvent(match, wallTime-match.GameState.FullPauseOffset, state.PhaseGameOver)
	} else {
		if len(match.GameState.ReactionStack) > 0 {
			match.TransitionPauseState(state.PausePhaseSimPause, wallTime, -1)
		}
		scheduleReactionStackResolution(match)
		match.processAutoSummonQueue(wallTime, -1)
		scheduleEarliestSimItem(match)
	}
	match.publish(wallTime, oldLogsSize)
}
