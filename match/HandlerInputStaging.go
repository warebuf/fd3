package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleStage(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
		return
	}

	if teamIndex >= 0 && teamIndex < len(match.GameState.Ready) && match.GameState.Ready[teamIndex] {
		return
	}

	card := match.GameState.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIndex {
		return
	}
	if card.StageOrder != 0 {
		return
	}

	order := nextStageOrder(match.GameState.Cards, teamIndex)
	match.GameState.Cards[cardIndex].StageOrder = order

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.StageLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.StageIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		StageOrder:    order,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleUnstage(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
		return
	}

	if teamIndex >= 0 && teamIndex < len(match.GameState.Ready) && match.GameState.Ready[teamIndex] {
		return
	}

	card := match.GameState.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIndex {
		return
	}
	if card.StageOrder == 0 {
		return
	}

	clearStageOrder(match.GameState.Cards, teamIndex, cardIndex)

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.StageLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.StageIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		StageOrder:    0,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleAutoSummon(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay && gs.Phase != state.PhaseReveal {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIndex {
		return
	}
	if card.AutoSummon != 0 {
		return
	}
	if card.StageOrder != 0 {
		return
	}

	// reject if card requires a stack target and there is nothing to target
	if defs.CardDefs[card.DefNo].RequiresStackTarget && len(gs.ReactionStack) == 0 {
		return
	}

	// during reveal, always queue — chaining during reveal requires a different input type
	if gs.Phase == state.PhaseReveal {
		match.queueAutoSummon(teamIndex, cardIndex, ie.WallTime.UnixMilli())
		return
	}

	// during full pause, only queue
	if gs.PausePhase == state.PausePhaseFullPause {
		match.queueAutoSummon(teamIndex, cardIndex, ie.WallTime.UnixMilli())
		return
	}

	// simulating or sim pause: queue, if autosummon size is 1, try to play immediately
	autosummonIndex := match.queueAutoSummon(teamIndex, cardIndex, ie.WallTime.UnixMilli())
	if autosummonIndex == 1 {
		oldLogsSize := len(match.EventLog)
		executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
		match.processAutoSummonQueue(executeAtWallTime, teamIndex)
		match.publish(executeAtWallTime, oldLogsSize)
	}

}

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

			success := schedule.ScheduleDeclareCard(
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

			match.TransitionPauseState(1, executeAtWallTime, -1)
			scheduleReactionStackResolution(match)

		}
	}

	scheduleEarliestSimItem(match)

}

func (match *Match) handleUnAutoSummon(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhasePlay && match.GameState.Phase != state.PhaseReveal {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
		return
	}

	card := match.GameState.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIndex {
		return
	}
	if card.AutoSummon == 0 {
		return
	}

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD

	clearAutoSummon(match.GameState.Cards, teamIndex, cardIndex)
	match.EventLog = append(match.EventLog, logs.AutoSummonLog{
		WallTimestamp:      executeAtWallTime,
		UpdateID:           match.NextUpdateID,
		Type:               logs.AutoSummonIndex,
		TeamIndex:          teamIndex,
		CardIndex:          cardIndex,
		AutoSummonPriority: 0,
	})
	match.NextUpdateID++

	match.processAutoSummonQueue(executeAtWallTime, teamIndex)
	match.publish(executeAtWallTime, oldLogsSize)

}

// returns the autosummon index assigned
func (match *Match) queueAutoSummon(teamIndex, cardIndex int, popAtWallTime int64) int {

	executeAtWallTime := popAtWallTime + constants.LEAD

	oldLogsSize := len(match.EventLog)
	order := nextAutoSummon(match.GameState.Cards, teamIndex)
	match.GameState.Cards[cardIndex].AutoSummon = order
	match.EventLog = append(match.EventLog, logs.AutoSummonLog{
		WallTimestamp:      executeAtWallTime,
		UpdateID:           match.NextUpdateID,
		Type:               logs.AutoSummonIndex,
		TeamIndex:          teamIndex,
		CardIndex:          cardIndex,
		AutoSummonPriority: order,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

	return order
}

func nextStageOrder(cards []state.Card, team int) int {
	next := 1
	for _, c := range cards {
		if c.Team == team && c.StageOrder >= next {
			next = c.StageOrder + 1
		}
	}
	return next
}

func clearStageOrder(cards []state.Card, team int, cardIndex int) {
	cleared := cards[cardIndex].StageOrder
	cards[cardIndex].StageOrder = 0
	for i := range cards {
		if cards[i].Team == team && cards[i].StageOrder > cleared {
			cards[i].StageOrder--
		}
	}
}

func nextAutoSummon(cards []state.Card, team int) int {
	next := 1
	for _, c := range cards {
		if c.Team == team && c.AutoSummon >= next {
			next = c.AutoSummon + 1
		}
	}
	return next
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
