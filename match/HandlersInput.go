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
	if defs.CardDefs[card.DefNo].RequiresStackTarget && len(gs.Stack) == 0 {
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

	if defs.CardDefs[card.DefNo].RequiresStackTarget && len(gs.Stack) == 0 {
		return
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	if gs.PausePhase == state.PausePhaseSimPause {
		oldLogsSize := len(match.EventLog)
		success := ScheduleDeclareCard(gs, &match.EventLog, &match.NextUpdateID, teamIndex, cardIndex, ie.WallTime.UnixMilli())
		if success {
			scheduleStackResolution(match)
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
		scheduleStackResolution(match)
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
				WallTimestamp: executeAtWallTime,
				UpdateID:      match.NextUpdateID,
				Type:          logs.AutoSummonIndex,
				TeamIndex:     t,
				CardIndex:     bestCard,
				AutoSummonIdx: 0,
			})
			match.NextUpdateID++

			match.TransitionPauseState(1, executeAtWallTime, -1)
			scheduleStackResolution(match)

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
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.AutoSummonIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		AutoSummonIdx: 0,
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
		WallTimestamp: executeAtWallTime,
		UpdateID:      match.NextUpdateID,
		Type:          logs.AutoSummonIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		AutoSummonIdx: order,
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

	def := defs.CardDefs[card.DefNo]
	if gs.Mana[teamIndex] < def.UseCost {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	gs.Mana[teamIndex] -= def.UseCost
	match.EventLog = append(match.EventLog, logs.SpendManaLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        def.UseCost,
	})
	match.NextUpdateID++

	effs := make([]state.Effect, len(def.UseEffects))
	for i, eff := range def.UseEffects {
		effs[i] = sim.FulfillEffectTemplate(eff, teamIndex, cardIndex)
	}
	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	stackItem := state.StackItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardID:              cardIndex,
		Type:                state.StackItemContinuousActivation,
		Effects:             effs,
	}
	gs.Stack = append(gs.Stack, stackItem)
	match.EventLog = append(match.EventLog, logs.AddStackLog{
		UpdateID:            match.NextUpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           stackItem.Type,
		StartAtActiveTime:   stackItem.StartAtActiveTime,
		PopAtActiveTime:     stackItem.PopAtActiveTime,
		ExecuteAtActiveTime: stackItem.ExecuteAtActiveTime,
		CardID:              stackItem.CardID,
		Effects:             stackItem.Effects,
	})
	match.NextUpdateID++

	match.TransitionPauseState(1, executeAtWallTime, -1)
	scheduleStackResolution(match)
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
