package match

import (
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInputEvent(ie *eventpq.InputEvent) {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	switch ie.Type {
	case "fullPause":
		match.handleInputPause(ie)
	case "stage":
		match.handleStage(ie)
	case "unstage":
		match.handleUnstage(ie)
	case "autosummon":
		match.handleAutoSummon(ie)
	case "unautosummon":
		match.handleUnAutoSummon(ie)
	case "ready":
		match.handleInputReady(ie)
	case "unready":
		match.handleInputUnready(ie)
	case "active":
		match.handleUseContinuous(ie)
	case "instant":
		match.handleInstantPlay(ie)
	case "preselect":
		match.handlePreselect(ie)
	case "unitAbility":
		match.handleUnitAbility(ie)
	case "toggleAutoAbility":
		match.handleToggleAutoAbility(ie)
	}

}

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

			success := ScheduleDeclareCard(
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

func (match *Match) handleInputReady(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(match.GameState.Ready) {
		return
	}

	if match.GameState.Ready[teamIndex] {
		return
	}

	match.GameState.Ready[teamIndex] = true

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.ReadyLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.Ready,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++
	match.publish(executeAtWallTime, oldLogsSize)

	allReady := true
	for _, r := range match.GameState.Ready {
		if !r {
			allReady = false
			break
		}
	}
	if allReady {
		PushPhaseEvent(match, executeAtWallTime-match.GameState.FullPauseOffset, state.PhaseReveal)
	}

}

func (match *Match) handleInputUnready(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(match.GameState.Ready) {
		return
	}

	if !match.GameState.Ready[teamIndex] {
		return
	}

	allWereReady := true
	for _, r := range match.GameState.Ready {
		if !r {
			allWereReady = false
			break
		}
	}

	match.GameState.Ready[teamIndex] = false

	oldLogsSize := len(match.EventLog)
	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	match.EventLog = append(match.EventLog, logs.UnreadyLog{
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.Unready,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++

	if allWereReady {
		PushPhaseEvent(match, match.GameState.PhaseEndInActiveTime, state.PhaseReveal)
	}

	match.publish(executeAtWallTime, oldLogsSize)

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
		effs = append(effs, sim.FulfillEffectTemplate(eff, teamIndex, cardIndex))
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

func (match *Match) handleInputPause(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIndex := ie.TeamIndex

	executeTime := ie.WallTime.UnixMilli() + constants.LEAD
	//now := time.Now().UnixMilli()
	//fmt.Printf("handleInputPause: now=%d ieWallTime=%d executeTime=%d\n", now%1e7, ie.WallTime.UnixMilli()%1e7, executeTime%1e7)

	if gs.Phase < state.PhaseDraw {
		return
	}
	if gs.Phase == state.PhaseGameOver {
		return
	}
	if teamIndex < 0 || teamIndex >= len(gs.TimeBank) {
		return
	}

	if gs.PausePhase == state.PausePhaseSimulating || gs.PausePhase == state.PausePhaseSimPause {
		if gs.TimeBank[teamIndex] <= 0 {
			return
		}
		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(state.PausePhaseFullPause, executeTime, teamIndex)

		pauseEnd := &eventpq.PauseEndEvent{WallTime: time.UnixMilli(executeTime + gs.TimeBank[teamIndex] - constants.LEAD)}
		match.PauseEndItem = match.EventPQueue.CreateAndPushEventItem(pauseEnd, "PauseEndEvent")

		match.publish(executeTime, oldLogsSize)

	} else if gs.PausePhase == state.PausePhaseFullPause && gs.TeamPaused == teamIndex {

		if match.PauseEndItem != nil {
			match.EventPQueue.Remove(match.PauseEndItem)
			match.PauseEndItem = nil
		}

		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(-1, executeTime, teamIndex)
		match.processAutoSummonQueue(executeTime, -1)
		match.publish(executeTime, oldLogsSize)
	}

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
		effs = append(effs, sim.FulfillEffectTemplate(eff, teamIndex, cardIndex))
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
