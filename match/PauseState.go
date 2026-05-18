package match

import "fmt"

func (match *Match) TransitionPauseState(newPauseState int, transitionTime int64) {

	gameState := match.GameState

	if newPauseState < 0 {
		switch gameState.PausePhase {
		case 2: // already simulating
			return
		case 1: // end sim pause, enter simulating
			gameState.PausePhase = 2
			gameState.SimPauseOffset += transitionTime - gameState.PauseStack[1]
			gameState.PauseStack[1] = 0
			scheduleEarliestSimItem(match)
			return
		case 0: // end indefinite pause
			if gameState.PauseStack[1] > 0 {
				gameState.PausePhase = 1
				gameState.FullPauseOffset += transitionTime - gameState.PauseStack[0]
				gameState.PauseStack[0] = 0
				scheduleChainResolution(match)
				if match.PendingPhase > 0 {
					PushPhase(match, match.PendingPhaseTime, match.PendingPhase)
				}
				return
			}
			gameState.PausePhase = 2
			gameState.SimPauseOffset += transitionTime - gameState.PauseStack[0]
			gameState.FullPauseOffset += transitionTime - gameState.PauseStack[0]
			gameState.PauseStack[0] = 0
			scheduleChainResolution(match)
			scheduleEarliestSimItem(match)
			if match.PendingPhase > 0 {
				PushPhase(match, match.PendingPhaseTime, match.PendingPhase)
			}
			return
		}
	}

	switch newPauseState {
	case 0: // entering indefinite pause
		gameState.PausePhase = 0
		gameState.PauseStack[0] = transitionTime
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

	case 1: // entering sim pause
		gameState.PausePhase = 1
		gameState.PauseStack[1] = transitionTime
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}

	case 2:
		fmt.Println("please do not use the function this way. just unpause by passing -1 into newPauseState")
	}
}
