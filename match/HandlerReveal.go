package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleRevealEvent(ev *eventpq.RevealEvent) {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	executeAtWallTime := ev.WallTime.UnixMilli() + constants.LEAD

	if len(match.RevealQueue) == 0 {
		match.PendingRevealTime = 0
		SchedulePhaseEvent(match, executeAtWallTime-match.GameState.FullPauseOffset, state.PhasePlay)
		return
	}

	next := match.RevealQueue[0]
	match.RevealQueue = match.RevealQueue[1:]

	oldLogsSize := len(match.EventLog)
	success := reaction.ScheduleDeclareCard(
		match.GameState,
		&match.EventLog,
		&match.NextUpdateID,
		next.TeamIndex,
		next.CardIndex,
		ev.WallTime.UnixMilli(),
	)

	if success {
		match.PendingRevealTime = 0            // clear previous reveal
		scheduleReactionStackResolution(match) // no need to sim pause, already sim paused
	} else {
		pushRevealEvent(match, executeAtWallTime-match.GameState.FullPauseOffset+constants.REACTION_WINDOW) // make a reveal same amount of time as 1 stack item
	}
	match.publish(executeAtWallTime, oldLogsSize)

}
