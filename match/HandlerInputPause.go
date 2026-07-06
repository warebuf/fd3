package match

import (
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInputPause(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIndex := ie.TeamIndex

	executeTime := ie.WallTime.UnixMilli() + constants.LEAD
	//now := time.Now().UnixMilli()
	//fmt.Printf("handleInputPause: now=%d ieWallTime=%d executeTime=%d\n", now%1e7, ie.WallTime.UnixMilli()%1e7, executeTime%1e7)

	if gs.Phase < state.PhaseDraw {
		return
	}
	if gs.Phase == state.PhaseGameOver {
		return
	}
	if teamIndex < 0 || teamIndex >= len(gs.TimeBank) {
		return
	}

	if gs.PausePhase == state.PausePhaseSimulating || gs.PausePhase == state.PausePhaseSimPause {
		if gs.TimeBank[teamIndex] <= 0 {
			return
		}
		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(state.PausePhaseFullPause, executeTime, teamIndex)

		pauseEnd := &eventpq.PauseEndEvent{WallTime: time.UnixMilli(executeTime + gs.TimeBank[teamIndex] - constants.LEAD)}
		match.PauseEndItem = match.EventPQueue.CreateAndPushEventItem(pauseEnd, "PauseEndEvent")

		match.publish(executeTime, oldLogsSize)

	} else if gs.PausePhase == state.PausePhaseFullPause && gs.TeamPaused == teamIndex {

		if match.PauseEndItem != nil {
			match.EventPQueue.Remove(match.PauseEndItem)
			match.PauseEndItem = nil
		}

		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(-1, executeTime, teamIndex)
		match.processAutoSummonQueue(executeTime, -1)
		match.publish(executeTime, oldLogsSize)
	}

}
