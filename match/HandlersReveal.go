package match

import "github.com/warebuf/fd3/match/eventpq"

func (match *Match) handleRevealEvent(ev *eventpq.RevealEvent) {

	wallTime := ev.T.UnixMilli()

	if len(match.RevealQueue) == 0 {
		PushPhase(match, wallTime+2000, 4)
		return
	}

	next := match.RevealQueue[0]
	match.RevealQueue = match.RevealQueue[1:]

	oldLogsSize := len(match.EventLog)
	chainStarted := ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, next.TeamIdx, next.CardIndex, wallTime)
	if chainStarted {
		scheduleChainResolution(match)
	} else {
		scheduleReveal(match, wallTime+800)
	}
	match.publish(wallTime, oldLogsSize)

}
