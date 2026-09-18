package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleToggleAutoAbility(ie *eventpq.InputEvent) {

	gs := match.GameState
	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex

	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Team != teamIndex {
		return
	}
	if card.Zone != state.ZoneField && card.Zone != state.ZoneContinuous {
		return
	}

	// input events don't yet carry which Active slot to toggle, so this always targets slot 0
	// (no current card has more than one Active ability)
	const activeIndex = 0

	if !card.Actives[activeIndex].CanAutoFire {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	on := !gs.Cards[cardIndex].Actives[activeIndex].AutoAbility
	gs.Cards[cardIndex].Actives[activeIndex].AutoAbility = on
	match.EventLog = append(match.EventLog, logs.ToggleAutoAbilityLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.ToggleAutoAbility,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		AbilityIndex:  activeIndex,
		On:            on,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleToggleStructureAutoAbility(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIndex := ie.TeamIndex

	if teamIndex < 0 || teamIndex >= len(gs.Structures) {
		return
	}

	structureIndex := ie.StructureIndex
	if structureIndex < 0 || structureIndex >= len(gs.Structures[teamIndex]) {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	on := !gs.Structures[teamIndex][structureIndex].AutoAbility
	gs.Structures[teamIndex][structureIndex].AutoAbility = on
	match.EventLog = append(match.EventLog, logs.ToggleStructureAutoAbilityLog{
		UpdateID:       match.NextUpdateID,
		WallTimestamp:  executeAtWallTime,
		Type:           logs.ToggleStructureAutoAbility,
		TeamIndex:      teamIndex,
		StructureIndex: structureIndex,
		On:             on,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

}
