package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInputReady(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(match.GameState.Ready) {
		return
	}

	if match.GameState.Ready[teamIndex] {
		return
	}

	match.GameState.Ready[teamIndex] = true

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.ReadyLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.Ready,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

	allReady := true
	for _, r := range match.GameState.Ready {
		if !r {
			allReady = false
			break
		}
	}
	if allReady {
		SchedulePhaseEvent(match, executeAtWallTime-match.GameState.FullPauseOffset, state.PhaseReveal)
	}

}

func (match *Match) handleInputUnready(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(match.GameState.Ready) {
		return
	}

	if !match.GameState.Ready[teamIndex] {
		return
	}

	allWereReady := true
	for _, r := range match.GameState.Ready {
		if !r {
			allWereReady = false
			break
		}
	}

	match.GameState.Ready[teamIndex] = false

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.UnreadyLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.Unready,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++

	if allWereReady {
		SchedulePhaseEvent(match, match.GameState.PhaseEndInActiveTime, state.PhaseReveal)
	}

	match.publish(executeAtWallTime, oldLogsSize)

}
