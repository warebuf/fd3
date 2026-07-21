package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handlePauseEndEvent(pauseEndEvent *eventpq.PauseEndEvent) {
	if match.GameState.Phase == state.PhaseGameOver {
		return
	}
	wallTime := pauseEndEvent.WallTime.UnixMilli() + constants.LEAD
	if match.GameState.PausePhase == state.PausePhaseFullPause {
		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(state.PauseTransitionUnpause, wallTime, -1)
		match.processAutoSummonQueue(wallTime, -1)
		match.publish(wallTime, oldLogsSize)
	}
}
