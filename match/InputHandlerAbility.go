package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleUseContinuous(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneContinuous || card.Team != teamIndex {
		return
	}

	// input events don't yet carry which Active slot to use, so this always targets slot 0
	// (no current card has more than one Active ability)
	const activeIndex = 0
	def := defs.CardDefs[card.DefNo].Active[activeIndex]

	if !def.Unlimited && card.PendingCharges >= card.Charges {
		return
	}

	if gs.Mana[teamIndex] < def.Cost {
		return
	}

	if !conditions.CheckActiveAbilityConditions(gs, cardIndex, activeIndex) {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	if !reaction.PushAbilityOnStack(gs, &match.EventLog, &match.NextUpdateID, cardIndex, activeIndex, executeAtWallTime) {
		return
	}

	match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
	scheduleReactionStackResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleUnitAbility(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneField || card.Team != teamIndex {
		return
	}

	// HARDCODE: ALWAYS TARGETS SLOT 0
	const activeIndex = 0
	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]
	if !activeDef.Unlimited && card.PendingCharges >= card.Charges {
		return
	}

	if gs.Mana[teamIndex] < activeDef.Cost {
		return
	}

	if !conditions.CheckActiveAbilityConditions(gs, cardIndex, activeIndex) {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	if !reaction.PushAbilityOnStack(gs, &match.EventLog, &match.NextUpdateID, cardIndex, activeIndex, executeAtWallTime) {
		return
	}

	match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
	scheduleReactionStackResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}
