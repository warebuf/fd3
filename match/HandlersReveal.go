package match

import "time"

func (match *Match) handleRevealEvent() {

	if len(match.RevealQueue) == 0 {
		PushPhase(match, time.Now().UnixMilli()+2000, 4)
		return
	}

	next := match.RevealQueue[0]
	match.RevealQueue = match.RevealQueue[1:]

	wallTime := time.Now().UnixMilli()
	oldLogsSize := len(match.EventLog)
	chainStarted := ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, next.TeamIdx, next.CardIndex, wallTime)
	if chainStarted {
		scheduleChainResolution(match)
	} else {
		scheduleReveal(match, wallTime+800)
	}
	match.publish(wallTime, oldLogsSize)

}
