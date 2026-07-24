package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/rules"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

// TODO: not sure about this func

func (match *Match) handlePreselect(ie *eventpq.InputEvent) {

	gs := match.GameState
	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	targetCardIndex := ie.TargetCardIndex

	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Team != teamIndex {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	switch card.Zone {
	case state.ZoneField:

		activeIndex := ie.AbilityIndex
		if activeIndex < 0 || activeIndex >= len(card.Actives) {
			return
		}
		if !rules.IsValidTarget(gs, card.Actives[activeIndex].TargetFlags, cardIndex, targetCardIndex, teamIndex) {
			return
		}

		gs.Cards[cardIndex].Actives[activeIndex].Target = targetCardIndex
		match.EventLog = append(match.EventLog, logs.PreselectLog{
			UpdateID:        match.NextUpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.Preselect,
			TeamIndex:       teamIndex,
			CardIndex:       cardIndex,
			AbilityIndex:    activeIndex,
			TargetCardIndex: targetCardIndex,
		})
		match.NextUpdateID++

	case state.ZoneHand:

		const onPlayIndex = 0
		if !rules.IsValidTarget(gs, defs.CardDefs[card.DefNo].OnPlay[onPlayIndex].TargetFlags, cardIndex, targetCardIndex, teamIndex) {
			return
		}

		gs.Cards[cardIndex].OnPlayTarget = targetCardIndex
		match.EventLog = append(match.EventLog, logs.PreselectLog{
			UpdateID:        match.NextUpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.Preselect,
			TeamIndex:       teamIndex,
			CardIndex:       cardIndex,
			AbilityIndex:    -1,
			TargetCardIndex: targetCardIndex,
		})
		match.NextUpdateID++

	default:
		return
	}

	match.publish(executeAtWallTime, oldLogsSize)

}
