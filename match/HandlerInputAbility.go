package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInstantPlay(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
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

	if defs.CardDefs[card.DefNo].RequiresStackTarget && len(gs.ReactionStack) == 0 {
		return
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	if gs.PausePhase == state.PausePhaseSimPause {
		oldLogsSize := len(match.EventLog)
		success := ScheduleDeclareCard(gs, &match.EventLog, &match.NextUpdateID, teamIndex, cardIndex, ie.WallTime.UnixMilli())
		if success {
			scheduleReactionStackResolution(match)
			match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
		}
		return
	}

	// simulating: try to play immediately, no queue fallback
	oldLogsSize := len(match.EventLog)
	success := ScheduleDeclareCard(
		gs,
		&match.EventLog,
		&match.NextUpdateID,
		teamIndex,
		cardIndex,
		ie.WallTime.UnixMilli(),
	)

	if success {
		match.TransitionPauseState(1, ie.WallTime.UnixMilli()+constants.LEAD, -1)
		scheduleReactionStackResolution(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
	}

}

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
	active := card.Actives[activeIndex]

	if active.Pending >= active.Charges {
		return
	}

	def := defs.CardDefs[card.DefNo].Active[activeIndex]
	if gs.Mana[teamIndex] < def.Cost {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	gs.Mana[teamIndex] -= def.Cost
	match.EventLog = append(match.EventLog, logs.SpendManaLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        def.Cost,
	})
	match.NextUpdateID++

	var effs []state.Effect
	for _, eff := range def.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(eff, teamIndex, cardIndex))
	}
	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		Type:                state.ReactionContinuousAbility,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	gs.Cards[cardIndex].Actives[activeIndex].Pending++
	match.EventLog = append(match.EventLog, logs.AddStackLog{
		UpdateID:            match.NextUpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	match.NextUpdateID++

	match.TransitionPauseState(1, executeAtWallTime, -1)
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

	// input events don't yet carry which Active slot to use, so this always targets slot 0
	// (no current card has more than one Active ability)
	const activeIndex = 0
	active := card.Actives[activeIndex]
	if active.Pending >= active.Charges {
		return
	}

	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]

	var effs []state.Effect
	for _, eff := range activeDef.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(eff, teamIndex, cardIndex))
	}
	if len(effs) == 0 {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		Type:                state.ReactionUnitAbility,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	gs.Cards[cardIndex].Actives[activeIndex].Pending++
	match.EventLog = append(match.EventLog, logs.AddStackLog{
		UpdateID:            match.NextUpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	match.NextUpdateID++

	match.TransitionPauseState(1, executeAtWallTime, -1)
	scheduleReactionStackResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}

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
		TargetCardIndex: targetCardIndex,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

}

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
