package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

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
		// input events don't yet carry which Active slot to preselect for, so this always targets slot 0
		// (no current card has more than one Active ability)
		const activeIndex = 0
		if !defs.IsValidTarget(gs, card.Actives[activeIndex].TargetFlags, cardIndex, targetCardIndex, teamIndex) {
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
		// preselecting the target for this card's OnPlay ability, chosen before it's played
		// (same "no card has more than one such ability" assumption as the Active case above)
		const onPlayIndex = 0
		if !defs.IsValidTarget(gs, defs.CardDefs[card.DefNo].OnPlay[onPlayIndex].TargetFlags, cardIndex, targetCardIndex, teamIndex) {
			return
		}

		gs.Cards[cardIndex].OnPlayTarget = targetCardIndex
		match.EventLog = append(match.EventLog, logs.PreselectLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.Preselect,
			TeamIndex:     teamIndex,
			CardIndex:     cardIndex,
			// -1 mirrors ReactionItem.ActiveIndex's convention: this preselect is for
			// an OnPlay ability, not an Active-ability slot
			AbilityIndex:    -1,
			TargetCardIndex: targetCardIndex,
		})
		match.NextUpdateID++

	default:
		return
	}

	match.publish(executeAtWallTime, oldLogsSize)

}
