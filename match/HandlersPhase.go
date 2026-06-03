package match

import (
	"sort"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleNewPhaseEvent(newPhase *eventpq.NewPhaseEvent) {

	// skip stale phase events (e.g. board setup timer after early ready-up)
	if match.GameState.Phase >= newPhase.Phase {
		return
	}

	match.PendingPhase = 0
	match.PendingPhaseTime = 0

	switch newPhase.Phase {
	case state.PhaseDraw: // entering draw phase, we are in pause phase 2 here since we want to draw cards
		oldLogsSize := len(match.EventLog)

		// change the phase
		phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
		match.GameState.Phase = state.PhaseDraw
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: phaseChangeWallTime,
			Type:          logs.Phase,
			Phase:         state.PhaseDraw,
		})
		match.NextUpdateID += 1

		// schedule draws: 5 cards per team, staggered 250ms apart
		for i := 0; i < 5; i++ {
			ScheduleDrawCard(match.GameState, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeWallTime+500+int64(i)*250)
		}
		for i := 0; i < 5; i++ {
			ScheduleDrawCard(match.GameState, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeWallTime+2000+int64(i)*250)
		}

		// schedule board setup phase after draws complete
		PushPhaseEvent(match, phaseChangeWallTime-match.GameState.FullPauseOffset+3500, state.PhaseBoardSetup)

		match.publish(phaseChangeWallTime, oldLogsSize)

		// schedule draw cards
		scheduleEarliestSimItem(match)

	case state.PhaseBoardSetup: // entering board set up phase, enter pause phase 1

		oldLogsSize := len(match.EventLog)

		phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
		match.GameState.Phase = state.PhaseBoardSetup
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: phaseChangeWallTime,
			Type:          logs.Phase,
			Phase:         state.PhaseBoardSetup,
		})
		match.NextUpdateID += 1

		match.TransitionPauseState(1, phaseChangeWallTime, -1)

		// initialize ready flags
		numTeams := len(match.GameState.Hand)
		match.GameState.Ready = make([]bool, numTeams)

		// schedule reveal in 90 seconds; can be triggered early by both players readying up
		PushPhaseEvent(match, phaseChangeWallTime-match.GameState.FullPauseOffset+90000, state.PhaseReveal)

		match.publish(phaseChangeWallTime, oldLogsSize)

	case state.PhaseReveal: // entering reveal, remain in pause phase 1

		oldLogsSize := len(match.EventLog)

		phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
		match.GameState.Phase = state.PhaseReveal
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: phaseChangeWallTime,
			Type:          logs.Phase,
			Phase:         state.PhaseReveal,
		})
		match.NextUpdateID += 1

		// collect staged cards into stagedByTeam
		numTeams := len(match.GameState.Hand)
		stagedByTeam := make([][]int, numTeams)
		for i := 0; i < len(match.GameState.Cards); i++ {
			if match.GameState.Cards[i].StageOrder > 0 {
				stagedByTeam[match.GameState.Cards[i].Team] = append(stagedByTeam[match.GameState.Cards[i].Team], i)
			}
		}
		for teamIndex := 0; teamIndex < len(stagedByTeam); teamIndex++ {
			sort.Slice(stagedByTeam[teamIndex], func(a, b int) bool { // sort by ascending
				return match.GameState.Cards[stagedByTeam[teamIndex][a]].StageOrder < match.GameState.Cards[stagedByTeam[teamIndex][b]].StageOrder
			})
		}

		// get the max size
		maxStaged := 0
		for teamIndex := 0; teamIndex < len(stagedByTeam); teamIndex++ {
			if len(stagedByTeam[teamIndex]) > maxStaged {
				maxStaged = len(stagedByTeam[teamIndex])
			}
		}

		// interweave the reveals
		match.RevealQueue = match.RevealQueue[:0] // set RevealQueue to empty
		for i := 0; i < maxStaged; i++ {
			for teamIndex := 0; teamIndex < len(stagedByTeam); teamIndex++ {
				if i >= len(stagedByTeam[teamIndex]) {
					continue
				}
				cardIndex := stagedByTeam[teamIndex][i]
				match.RevealQueue = append(match.RevealQueue, RevealEntry{TeamIndex: teamIndex, CardIndex: cardIndex})
			}
		}

		// set all ready flags back to false
		match.GameState.Ready = make([]bool, len(match.GameState.Ready))

		pushRevealEvent(match, phaseChangeWallTime-match.GameState.FullPauseOffset+500)

		match.publish(phaseChangeWallTime, oldLogsSize)

	case state.PhasePlay: // entering play

		oldLogsSize := len(match.EventLog)

		phaseChangeTimestamp := newPhase.WallTime.UnixMilli() + constants.LEAD
		match.GameState.Phase = state.PhasePlay
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: phaseChangeTimestamp,
			Type:          logs.Phase,
			Phase:         state.PhasePlay,
		})
		match.NextUpdateID += 1

		for i := range match.GameState.Cards {
			match.GameState.Cards[i].StageOrder = 0
		}

		match.TransitionPauseState(-1, phaseChangeTimestamp, -1)

		// begin mana regen for structures
		startSimTime := phaseChangeTimestamp - match.GameState.SimPauseOffset
		for i := range match.GameState.Structures {
			sim.ScheduleStructureManaRegen(match.GameState, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, i, startSimTime)
		}
		scheduleEarliestSimItem(match)

		match.publish(phaseChangeTimestamp, oldLogsSize)

	case state.PhaseGameOver: // entering game over
		oldLogsSize := len(match.EventLog)
		phaseChangeWallTime := newPhase.WallTime.UnixMilli() + constants.LEAD
		match.GameState.Phase = state.PhaseGameOver

		deadCount := 0
		for _, s := range match.GameState.Structures {
			if s.HP <= 0 {
				deadCount++
			}
		}
		if deadCount >= 2 {
			match.GameState.Winner = -2
		}

		match.EventLog = append(match.EventLog, logs.GameOverLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: phaseChangeWallTime,
			Type:          logs.GameOver,
			Winner:        match.GameState.Winner,
		})
		match.NextUpdateID += 1

		// cancel all pending scheduled events
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
		if match.PauseEndItem != nil {
			match.EventPQueue.Remove(match.PauseEndItem)
			match.PauseEndItem = nil
		}

		match.publish(phaseChangeWallTime, oldLogsSize)

	}

}
