package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleAutoSummon(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhaseDraw && gs.Phase != state.PhasePlay {
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

	// reject if card requires a stack target and there is nothing to target
	if defs.CardDefs[card.DefNo].RequiresStackTarget && len(gs.ReactionStack) == 0 {
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

func (match *Match) handleUnAutoSummon(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseDraw && match.GameState.Phase != state.PhasePlay {
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

	match.trySkipStaging(executeAtWallTime)

	match.publish(executeAtWallTime, oldLogsSize)

	return order
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
