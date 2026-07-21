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
	if card.Zone != state.ZoneField || card.Team != teamIndex {
		return
	}

	// input events don't yet carry which Active slot to preselect for, so this always targets slot 0
	// (no current card has more than one Active ability)
	const activeIndex = 0
	if !isValidTarget(gs, card.Actives[activeIndex].TargetFlags, cardIndex, targetCardIndex, teamIndex) {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

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
	match.publish(executeAtWallTime, oldLogsSize)

}

func isValidTarget(gs *state.GameState, targetFlags int, sourceIndex int, targetIndex int, teamIndex int) bool {
	if targetFlags == 0 {
		return false
	}
	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return false
	}
	target := gs.Cards[targetIndex]
	if targetIndex == sourceIndex {
		return targetFlags&defs.TargetSelf != 0
	}
	switch {
	case target.Team == teamIndex && target.Zone == state.ZoneField:
		return targetFlags&defs.TargetOwnField != 0
	case target.Team != teamIndex && target.Zone == state.ZoneField:
		return targetFlags&defs.TargetOppField != 0
	case target.Team == teamIndex && target.Zone == state.ZoneHand:
		return targetFlags&defs.TargetOwnHand != 0
	case target.Team != teamIndex && target.Zone == state.ZoneHand:
		return targetFlags&defs.TargetOppHand != 0
	case target.Team == teamIndex && target.Zone == state.ZoneContinuous:
		return targetFlags&defs.TargetOwnCont != 0
	case target.Team != teamIndex && target.Zone == state.ZoneContinuous:
		return targetFlags&defs.TargetOppCont != 0
	}
	return false
}
