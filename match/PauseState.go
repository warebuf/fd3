package match

import (
	"fmt"

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
			gameState.PausePhase = state.PausePhaseSimulating
			gameState.SimPauseOffset += transitionAtWallTime - gameState.PauseStack[1]
			gameState.PauseStack[1] = 0
			scheduleEarliestSimItem(match)
			return
		case state.PausePhaseFullPause: // end indefinite pause
			if gameState.PauseStack[1] > 0 {
				gameState.PausePhase = state.PausePhaseSimPause
				gameState.FullPauseOffset += transitionAtWallTime - gameState.PauseStack[0]
				gameState.PauseStack[0] = 0
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

	case state.PausePhaseSimPause: // entering sim pause
		gameState.PausePhase = state.PausePhaseSimPause
		gameState.PauseStack[1] = transitionAtWallTime
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}

	case state.PausePhaseSimulating:
		fmt.Println("please do not use the function this way. just unpause by passing -1 into newPauseState")
	}
}
