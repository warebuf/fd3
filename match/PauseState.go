package match

import (
	"fmt"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) TransitionPauseState(newPauseState int, transitionAtWallTime int64, teamPaused int) {

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

			// create the log (sim pause always ended by system not user)
			match.EventLog = append(match.EventLog, logs.UnpauseLog{
				WallTimestamp: transitionAtWallTime,
				UpdateID:      match.NextUpdateID,
				Type:          logs.Unpause,
				TimeRemaining: 0,
			})
			match.NextUpdateID++

			// update the states (no need to update time bank)
			gameState.PausePhase = state.PausePhaseSimulating
			gameState.SimPauseOffset += transitionAtWallTime - gameState.SimPauseStartedAt
			gameState.TeamPaused = -1
			gameState.SimPauseStartedAt = 0

			// reschedule unpaused events
			scheduleEarliestSimItem(match)

		case state.PausePhaseFullPause: // end indefinite pause

			// create the log
			timeRemaining := int64(0)
			if gameState.TeamPaused >= 0 {
				timeElapsed := transitionAtWallTime - gameState.FullPauseStartedAt
				gameState.TimeBank[gameState.TeamPaused] = gameState.TimeBank[gameState.TeamPaused] - timeElapsed
				if gameState.TimeBank[gameState.TeamPaused] < 0 {
					gameState.TimeBank[gameState.TeamPaused] = 0
				}
				timeRemaining = gameState.TimeBank[gameState.TeamPaused]
			}
			match.EventLog = append(match.EventLog, logs.UnpauseLog{
				WallTimestamp: transitionAtWallTime,
				UpdateID:      match.NextUpdateID,
				Type:          logs.Unpause,
				TimeRemaining: timeRemaining,
			})
			match.NextUpdateID++

			// if we full paused amid a sim paused, enter sim paused
			if gameState.SimPauseStartedAt > 0 {
				gameState.PausePhase = state.PausePhaseSimPause
				gameState.FullPauseOffset += transitionAtWallTime - gameState.FullPauseStartedAt
				gameState.FullPauseStartedAt = 0
				gameState.TeamPaused = -1

				scheduleStackResolution(match)
				if match.PendingPhase > 0 {
					PushPhaseEvent(match, match.PendingPhaseTime, match.PendingPhase)
				}
				if match.PendingRevealTime > 0 {
					pushRevealEvent(match, match.PendingRevealTime)
				}

			} else {
				gameState.PausePhase = state.PausePhaseSimulating
				gameState.SimPauseOffset += transitionAtWallTime - gameState.FullPauseStartedAt
				gameState.FullPauseOffset += transitionAtWallTime - gameState.FullPauseStartedAt
				gameState.FullPauseStartedAt = 0
				gameState.TeamPaused = -1

				scheduleStackResolution(match)
				scheduleEarliestSimItem(match)
				if match.PendingPhase > 0 {
					PushPhaseEvent(match, match.PendingPhaseTime, match.PendingPhase)
				}
				if match.PendingRevealTime > 0 {
					pushRevealEvent(match, match.PendingRevealTime)
				}
			}

		}

	}

	switch newPauseState {
	case state.PausePhaseFullPause: // entering indefinite pause

		// appending the log
		timeRemaining := int64(0)
		if teamPaused >= 0 {
			timeRemaining = gameState.TimeBank[teamPaused]
		}
		match.EventLog = append(match.EventLog, logs.PauseLog{
			WallTimestamp: transitionAtWallTime,
			UpdateID:      match.NextUpdateID,
			Type:          logs.Pause,
			PausePhase:    state.PausePhaseFullPause,
			TeamIndex:     teamPaused,
			TimeStart:     transitionAtWallTime,
			TimeRemaining: timeRemaining,
		})
		match.NextUpdateID++

		// changing the states
		gameState.PausePhase = state.PausePhaseFullPause
		gameState.FullPauseStartedAt = transitionAtWallTime
		gameState.TeamPaused = teamPaused

		// cancelling events that do not occur in a full pause
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}
		if match.StackItem != nil {
			match.EventPQueue.Remove(match.StackItem)
			match.StackItem = nil
		}
		if match.PhaseItem != nil {
			match.EventPQueue.Remove(match.PhaseItem)
			match.PhaseItem = nil
		}
		if match.RevealItem != nil {
			match.EventPQueue.Remove(match.RevealItem)
			match.RevealItem = nil
		}

	case state.PausePhaseSimPause: // entering sim pause

		// appending the log
		match.EventLog = append(match.EventLog, logs.PauseLog{
			WallTimestamp: transitionAtWallTime,
			UpdateID:      match.NextUpdateID,
			Type:          logs.Pause,
			PausePhase:    state.PausePhaseSimPause,
			TeamIndex:     teamPaused,
			TimeStart:     transitionAtWallTime,
			TimeRemaining: 0, // TODO: add the correct amount of time for the current sim pause
		})
		match.NextUpdateID++

		// changing the states
		gameState.PausePhase = state.PausePhaseSimPause
		gameState.SimPauseStartedAt = transitionAtWallTime
		gameState.TeamPaused = teamPaused

		// cancelling events that do not occur in a sim pause
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}

	case state.PausePhaseSimulating:
		fmt.Println("please do not use the function this way. just unpause by passing -1 into newPauseState")
		return
	}

}
