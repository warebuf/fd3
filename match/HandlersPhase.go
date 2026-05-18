package match

import (
	"sort"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
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
	case 1: // entering draw phase, we are in pause phase 2 here since we want to draw cards
		oldLogsSize := len(match.EventLog)

		// change the phase
		phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
		match.GameState.Phase = 1
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:  match.NextUpdateID,
			Timestamp: phaseChangeTimestamp,
			Type:      logs.Phase,
			Phase:     1,
		})
		match.NextUpdateID += 1

		// schedule draws: 5 cards per team, staggered 250ms apart
		for i := 0; i < 5; i++ {
			ScheduleDrawCard(match.GameState, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+500+int64(i)*250)
		}
		for i := 0; i < 5; i++ {
			ScheduleDrawCard(match.GameState, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2000+int64(i)*250)
		}

		// schedule board setup phase after draws complete
		PushPhase(match, phaseChangeTimestamp+3500, 2)

		match.publish(phaseChangeTimestamp, oldLogsSize)

		// schedule draw cards
		scheduleEarliestSimItem(match)

	case 2: // entering board set up phase, enter pause phase 1

		oldLogsSize := len(match.EventLog)

		phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
		match.GameState.Phase = 2
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:  match.NextUpdateID,
			Timestamp: phaseChangeTimestamp,
			Type:      logs.Phase,
			Phase:     2,
		})
		match.NextUpdateID += 1

		match.TransitionPauseState(1, phaseChangeTimestamp)

		// initialize ready flags
		numTeams := len(match.GameState.Hand)
		match.GameState.Ready = make([]bool, numTeams)

		// schedule reveal in 90 seconds; can be triggered early by both players readying up
		PushPhase(match, phaseChangeTimestamp+90000, 3)

		match.publish(phaseChangeTimestamp, oldLogsSize)

	case 3: // entering reveal, remain pause phase 1

		oldLogsSize := len(match.EventLog)

		phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
		match.GameState.Phase = 3
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:  match.NextUpdateID,
			Timestamp: phaseChangeTimestamp,
			Type:      logs.Phase,
			Phase:     3,
		})
		match.NextUpdateID += 1

		// collect staged cards per team, sorted by AutoSummon order
		numTeams := len(match.GameState.Hand)
		stagedByTeam := make([][]int, numTeams)
		for i, c := range match.GameState.Cards {
			if c.AutoSummon > 0 {
				stagedByTeam[c.Team] = append(stagedByTeam[c.Team], i)
			}
		}
		for teamIdx := range stagedByTeam {
			sort.Slice(stagedByTeam[teamIdx], func(a, b int) bool {
				return match.GameState.Cards[stagedByTeam[teamIdx][a]].AutoSummon < match.GameState.Cards[stagedByTeam[teamIdx][b]].AutoSummon
			})
		}

		// build reveal queue: alternating teams in AutoSummon order
		maxStaged := 0
		for _, staged := range stagedByTeam {
			if len(staged) > maxStaged {
				maxStaged = len(staged)
			}
		}
		match.RevealQueue = match.RevealQueue[:0]
		for i := 0; i < maxStaged; i++ {
			for teamIdx, staged := range stagedByTeam {
				if i >= len(staged) {
					continue
				}
				cardIndex := staged[i]
				match.GameState.Cards[cardIndex].AutoSummon = 0
				match.RevealQueue = append(match.RevealQueue, RevealEntry{TeamIdx: teamIdx, CardIndex: cardIndex})
			}
		}
		match.GameState.Ready = make([]bool, len(match.GameState.Ready))

		// schedule first reveal; phase 4 is pushed by handleRevealEvent when queue empties
		scheduleReveal(match, phaseChangeTimestamp+500)

		match.publish(phaseChangeTimestamp, oldLogsSize)

	case 4: // entering play

		oldLogsSize := len(match.EventLog)

		phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
		match.GameState.Phase = 4
		match.EventLog = append(match.EventLog, logs.PhaseLog{
			UpdateID:  match.NextUpdateID,
			Timestamp: phaseChangeTimestamp,
			Type:      logs.Phase,
			Phase:     4,
		})
		match.NextUpdateID += 1

		// enter simulating state
		match.TransitionPauseState(2, phaseChangeTimestamp)

		// update the published state and broadcast to clients
		match.publish(phaseChangeTimestamp, oldLogsSize)

	case 5: // entering game over
		// TODO: end game logic

	}

}
