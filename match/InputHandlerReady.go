package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInputReady(ie *eventpq.InputEvent) {

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

	match.trySkipStaging(executeAtWallTime)

	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleInputUnready(ie *eventpq.InputEvent) {

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(match.GameState.Ready) {
		return
	}

	if !match.GameState.Ready[teamIndex] {
		return
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
	match.publish(executeAtWallTime, oldLogsSize)

}
