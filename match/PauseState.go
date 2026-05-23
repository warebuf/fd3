package match

import (
	"fmt"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) TransitionPauseState(newPauseState int, transitionAtWallTime int64) {

	gameState := match.GameState

	if newPauseState >= 0 && newPauseState == gameState.PausePhase {
		return
	}

	// unpause the current pause
	if newPauseState < 0 {
		switch gameState.PausePhase {
		case state.PausePhaseSimulating: // already simulating
			return
		case state.PausePhaseSimPause: // end sim pause, enter simulating
			teamIdx := gameState.TeamPaused
			var timeRemaining int64
			if teamIdx >= 0 {
				timeRemaining = gameState.TimeBank[teamIdx]
			}
			gameState.PausePhase = state.PausePhaseSimulating
			gameState.SimPauseOffset += transitionAtWallTime - gameState.PauseStack[1]
			gameState.PauseStack[1] = 0
			match.EventLog = append(match.EventLog, logs.UnpauseLog{
				WallTimestamp: transitionAtWallTime,
				UpdateID:      match.NextUpdateID,
				Type:          logs.Unpause,
				PausePhase:    state.PausePhaseSimPause,
				TeamIndex:     teamIdx,
				TimeRemaining: timeRemaining,
			})
			match.NextUpdateID++
			scheduleEarliestSimItem(match)
			return
		case state.PausePhaseFullPause: // end indefinite pause
			teamIdx := gameState.TeamPaused
			var timeRemaining int64
			if teamIdx >= 0 {
				timeRemaining = gameState.TimeBank[teamIdx]
			}
			if gameState.PauseStack[1] > 0 {
				gameState.PausePhase = state.PausePhaseSimPause
				gameState.FullPauseOffset += transitionAtWallTime - gameState.PauseStack[0]
				gameState.PauseStack[0] = 0
				gameState.TeamPaused = -1
				match.EventLog = append(match.EventLog, logs.UnpauseLog{
					WallTimestamp: transitionAtWallTime,
					UpdateID:      match.NextUpdateID,
					Type:          logs.Unpause,
					PausePhase:    state.PausePhaseFullPause,
					TeamIndex:     teamIdx,
					TimeRemaining: timeRemaining,
				})
				match.NextUpdateID++
				scheduleChainResolution(match)
				if match.PendingPhase > 0 {
					PushPhaseEvent(match, match.PendingPhaseTime, match.PendingPhase)
				}
				if match.PendingRevealTime > 0 {
					pushRevealEvent(match, match.PendingRevealTime)
				}
				return
			}
			gameState.PausePhase = state.PausePhaseSimulating
			gameState.SimPauseOffset += transitionAtWallTime - gameState.PauseStack[0]
			gameState.FullPauseOffset += transitionAtWallTime - gameState.PauseStack[0]
			gameState.PauseStack[0] = 0
			gameState.TeamPaused = -1
			match.EventLog = append(match.EventLog, logs.UnpauseLog{
				WallTimestamp: transitionAtWallTime,
				UpdateID:      match.NextUpdateID,
				Type:          logs.Unpause,
				PausePhase:    state.PausePhaseFullPause,
				TeamIndex:     teamIdx,
				TimeRemaining: timeRemaining,
			})
			match.NextUpdateID++
			scheduleChainResolution(match)
			scheduleEarliestSimItem(match)
			if match.PendingPhase > 0 {
				PushPhaseEvent(match, match.PendingPhaseTime, match.PendingPhase)
			}
			if match.PendingRevealTime > 0 {
				pushRevealEvent(match, match.PendingRevealTime)
			}
			return
		}
	}

	switch newPauseState {
	case state.PausePhaseFullPause: // entering indefinite pause
		teamIdx := gameState.TeamPaused
		var timeRemaining int64
		if teamIdx >= 0 {
			timeRemaining = gameState.TimeBank[teamIdx]
		}
		gameState.PausePhase = state.PausePhaseFullPause
		gameState.PauseStack[0] = transitionAtWallTime
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}
		if match.ChainItem != nil {
			match.EventPQueue.Remove(match.ChainItem)
			match.ChainItem = nil
		}
		if match.PhaseItem != nil {
			match.EventPQueue.Remove(match.PhaseItem)
			match.PhaseItem = nil
		}
		if match.RevealItem != nil {
			match.EventPQueue.Remove(match.RevealItem)
			match.RevealItem = nil
		}
		match.EventLog = append(match.EventLog, logs.PauseLog{
			WallTimestamp: transitionAtWallTime,
			UpdateID:      match.NextUpdateID,
			Type:          logs.Pause,
			PausePhase:    state.PausePhaseFullPause,
			TeamIndex:     teamIdx,
			TimeRemaining: timeRemaining,
		})
		match.NextUpdateID++

	case state.PausePhaseSimPause: // entering sim pause
		teamIdx := gameState.TeamPaused
		var timeRemaining int64
		if teamIdx >= 0 {
			timeRemaining = gameState.TimeBank[teamIdx]
		}
		gameState.PausePhase = state.PausePhaseSimPause
		gameState.PauseStack[1] = transitionAtWallTime
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}
		match.EventLog = append(match.EventLog, logs.PauseLog{
			WallTimestamp: transitionAtWallTime,
			UpdateID:      match.NextUpdateID,
			Type:          logs.Pause,
			PausePhase:    state.PausePhaseSimPause,
			TeamIndex:     teamIdx,
			TimeRemaining: timeRemaining,
		})
		match.NextUpdateID++

	case state.PausePhaseSimulating:
		fmt.Println("please do not use the function this way. just unpause by passing -1 into newPauseState")
	}
}
