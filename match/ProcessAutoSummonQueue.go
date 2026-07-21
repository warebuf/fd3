package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) processAutoSummonQueue(executeAtWallTime int64, targetTeam int) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}

	if gs.PausePhase != state.PausePhaseSimulating {
		return
	}

	startTeam := 0
	endTeam := len(gs.Hand)
	if targetTeam >= 0 {
		startTeam = targetTeam
		endTeam = targetTeam + 1
	}

	for t := startTeam; t < endTeam; t++ {
		for {
			bestCard := -1
			bestPriority := 0
			for _, i := range gs.Hand[t] {
				if card := gs.Cards[i]; card.AutoSummon > 0 {
					if bestCard == -1 || card.AutoSummon < bestPriority {
						bestCard = i
						bestPriority = card.AutoSummon
					}
				}
			}
			if bestCard == -1 {
				break
			}

			success := reaction.ScheduleDeclareCard(
				gs,
				&match.EventLog,
				&match.NextUpdateID,
				t,
				bestCard,
				executeAtWallTime-constants.LEAD,
			)

			if success == false {
				break
			}

			clearAutoSummon(gs.Cards, t, bestCard)
			match.EventLog = append(match.EventLog, logs.AutoSummonLog{
				WallTimestamp:      executeAtWallTime,
				UpdateID:           match.NextUpdateID,
				Type:               logs.AutoSummonIndex,
				TeamIndex:          t,
				CardIndex:          bestCard,
				AutoSummonPriority: 0,
			})
			match.NextUpdateID++

			match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
			scheduleReactionStackResolution(match)

		}
	}

	scheduleEarliestSimItem(match)

}

func clearAutoSummon(cards []state.Card, team int, cardIndex int) {
	cleared := cards[cardIndex].AutoSummon
	cards[cardIndex].AutoSummon = 0
	for i := range cards {
		if cards[i].Team == team && cards[i].AutoSummon > cleared {
			cards[i].AutoSummon--
		}
	}
}
