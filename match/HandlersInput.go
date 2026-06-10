package match

import (
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
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

	nextStageOrder := 1
	for _, c := range match.GameState.Cards {
		if c.Team == teamIndex && c.StageOrder >= nextStageOrder {
			nextStageOrder = c.StageOrder + 1
		}
	}
	match.GameState.Cards[cardIndex].StageOrder = nextStageOrder

	oldLogsSize := len(match.EventLog)
	match.EventLog = append(match.EventLog, logs.StageLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.StageIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		StageOrder:    nextStageOrder,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

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

	cleared := match.GameState.Cards[cardIndex].StageOrder
	match.GameState.Cards[cardIndex].StageOrder = 0
	for i := range match.GameState.Cards {
		if match.GameState.Cards[i].Team == teamIndex && match.GameState.Cards[i].StageOrder > cleared {
			match.GameState.Cards[i].StageOrder--
		}
	}

	oldLogsSize := len(match.EventLog)
	match.EventLog = append(match.EventLog, logs.StageLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.StageIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		StageOrder:    0,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

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

	if gs.Phase == state.PhaseReveal {
		oldStackSize := len(gs.Stack)
		oldLogsSize := len(match.EventLog)
		result := ScheduleDeclareCard(gs, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, teamIndex, cardIndex, ie.WallTime.UnixMilli(), true)
		if result != DeclareChainStarted {
			return
		}
		if oldStackSize == 0 {
			if match.RevealItem != nil {
				match.EventPQueue.Remove(match.RevealItem)
				match.RevealItem = nil
			}
			match.PendingRevealTime = 0
		}
		scheduleStackResolution(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
		return
	}

	// try to play immediately
	oldLogsSize := len(match.EventLog)
	result := ScheduleDeclareCard(
		gs,
		&match.SimulationPQ,
		&match.AbilityIDCounter,
		&match.EventLog,
		&match.NextUpdateID,
		teamIndex,
		cardIndex,
		ie.WallTime.UnixMilli(),
		gs.PausePhase == state.PausePhaseSimPause,
	)

	switch result {
	case DeclareErrBadInput:
		return
	case DeclareSuccess:
		scheduleEarliestSimItem(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
		return
	case DeclareChainStarted:
		match.TransitionPauseState(1, ie.WallTime.UnixMilli()+constants.LEAD, -1)
		scheduleStackResolution(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
		return
	case DeclareConditionsNotMet:
		nextAutoSummon := 1
		for _, c := range gs.Cards {
			if c.Team == teamIndex && c.AutoSummon >= nextAutoSummon {
				nextAutoSummon = c.AutoSummon + 1
			}
		}
		gs.Cards[cardIndex].AutoSummon = nextAutoSummon

		match.EventLog = append(match.EventLog, logs.AutoSummonLog{
			WallTimestamp: ie.WallTime.UnixMilli(),
			UpdateID:      match.NextUpdateID,
			Type:          logs.AutoSummonIndex,
			TeamIndex:     teamIndex,
			CardIndex:     cardIndex,
			AutoSummonIdx: nextAutoSummon,
		})
		match.NextUpdateID++
		match.publish(ie.WallTime.UnixMilli(), oldLogsSize)
		return
	}

}

func (match *Match) processAutoSummonQueue(wallTime int64) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}

	for teamIndex := 0; teamIndex < len(gs.Hand); teamIndex++ {
		for {
			if gs.PausePhase != state.PausePhaseSimulating {
				return
			}

			// find the highest priority (lowest AutoSummon value) queued card in hand for this team
			bestCard := -1
			bestPriority := 0
			for i, card := range gs.Cards {
				if card.Team == teamIndex && card.Zone == state.ZoneHand && card.AutoSummon > 0 {
					if bestCard == -1 || card.AutoSummon < bestPriority {
						bestCard = i
						bestPriority = card.AutoSummon
					}
				}
			}
			if bestCard == -1 {
				break
			}

			result := ScheduleDeclareCard(
				gs,
				&match.SimulationPQ,
				&match.AbilityIDCounter,
				&match.EventLog,
				&match.NextUpdateID,
				teamIndex,
				bestCard,
				wallTime-constants.LEAD,
				false,
			)

			if result == DeclareConditionsNotMet {
				break
			}

			// clear from autosummon queue (played successfully or stale bad-input entry)
			cleared := gs.Cards[bestCard].AutoSummon
			gs.Cards[bestCard].AutoSummon = 0
			for i := range gs.Cards {
				if gs.Cards[i].Team == teamIndex && gs.Cards[i].AutoSummon > cleared {
					gs.Cards[i].AutoSummon--
				}
			}
			match.EventLog = append(match.EventLog, logs.AutoSummonLog{
				WallTimestamp: wallTime,
				UpdateID:      match.NextUpdateID,
				Type:          logs.AutoSummonIndex,
				TeamIndex:     teamIndex,
				CardIndex:     bestCard,
				AutoSummonIdx: 0,
			})
			match.NextUpdateID++

			if result == DeclareChainStarted {
				match.TransitionPauseState(1, wallTime, -1)
				scheduleStackResolution(match)
				return
			}
			// DeclareSuccess or DeclareErrBadInput: continue to next card in queue
		}
	}

	scheduleEarliestSimItem(match)

}

func (match *Match) handleUnAutoSummon(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhasePlay {
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

	cleared := match.GameState.Cards[cardIndex].AutoSummon
	match.GameState.Cards[cardIndex].AutoSummon = 0
	for i := range match.GameState.Cards {
		if match.GameState.Cards[i].Team == teamIndex && match.GameState.Cards[i].AutoSummon > cleared {
			match.GameState.Cards[i].AutoSummon--
		}
	}

	oldLogsSize := len(match.EventLog)
	match.EventLog = append(match.EventLog, logs.AutoSummonLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.AutoSummonIndex,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		AutoSummonIdx: 0,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

}

func (match *Match) handleInputReady(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	if match.GameState.PausePhase == state.PausePhaseFullPause {
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
	match.EventLog = append(match.EventLog, logs.ReadyLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.Ready,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

	allReady := true
	for _, r := range match.GameState.Ready {
		if !r {
			allReady = false
			break
		}
	}
	if allReady {
		PushPhaseEvent(match, ie.WallTime.UnixMilli()-match.GameState.FullPauseOffset, state.PhaseReveal)
	}

}

func (match *Match) handleInputUnready(ie *eventpq.InputEvent) {

	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	if match.GameState.PausePhase == state.PausePhaseFullPause {
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
	match.EventLog = append(match.EventLog, logs.UnreadyLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.Unready,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

}

func (match *Match) handleInputPause(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIndex := ie.TeamIndex

	executeTime := ie.WallTime.UnixMilli() + constants.LEAD

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
		match.processAutoSummonQueue(executeTime)
		match.publish(executeTime, oldLogsSize)
	}

}
