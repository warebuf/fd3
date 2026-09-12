package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInputSkipChain(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(gs.SkipChain) {
		return
	}
	if len(gs.Chain) == 0 {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause { // active time is frozen, nothing to collapse
		return
	}
	if gs.SkipChain[teamIndex] {
		return
	}

	gs.SkipChain[teamIndex] = true

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.SkipChainLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.SkipChain,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++

	if allSkipped(gs) {
		match.collapseChainWindow(executeAtWallTime)
	}

	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleInputUnskipChain(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(gs.SkipChain) {
		return
	}
	if !gs.SkipChain[teamIndex] {
		return
	}

	// once every team has skipped, the collapse has already been committed; too late to unskip
	if allSkipped(gs) {
		return
	}

	gs.SkipChain[teamIndex] = false

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.UnskipChainLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.UnskipChain,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

}

func allSkipped(gs *state.GameState) bool {
	for _, skip := range gs.SkipChain {
		if !skip {
			return false
		}
	}
	return true
}

// pulls the top chain item's window in so it resolves at executeAtWallTime instead of waiting out CHAIN_WINDOW
func (match *Match) collapseChainWindow(executeAtWallTime int64) {
	gs := match.GameState
	top := &gs.Chain[len(gs.Chain)-1]
	now := executeAtWallTime - gs.FullPauseOffset
	top.PopAtActiveTime = now - constants.LEAD
	top.ExecuteAtActiveTime = now

	match.EventLog = append(match.EventLog, logs.RefreshChainLog{
		UpdateID:            match.NextUpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.RefreshChain,
		CardIndex:           top.CardIndex,
		StartAtActiveTime:   top.StartAtActiveTime,
		PopAtActiveTime:     top.PopAtActiveTime,
		ExecuteAtActiveTime: top.ExecuteAtActiveTime,
	})
	match.NextUpdateID++

	scheduleChainResolution(match)
}
