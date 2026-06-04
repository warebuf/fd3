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
	if card.AutoSummon != 0 {
		return
	}

	// try to play immediately
	oldLogsSize := len(match.EventLog)
	chainStarted, err := ScheduleDeclareCard(match.GameState, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, teamIndex, cardIndex, ie.WallTime.UnixMilli())
	if err {
		return
	}

	if len(match.EventLog) > oldLogsSize {
		if chainStarted {
			match.TransitionPauseState(1, ie.WallTime.UnixMilli()+constants.LEAD, -1)
			scheduleChainResolution(match)
		} else {
			scheduleEarliestSimItem(match)
		}
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
		return
	}

	// conditions not met — enqueue
	nextAutoSummon := 1
	for _, c := range match.GameState.Cards {
		if c.Team == teamIndex && c.AutoSummon >= nextAutoSummon {
			nextAutoSummon = c.AutoSummon + 1
		}
	}
	match.GameState.Cards[cardIndex].AutoSummon = nextAutoSummon

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
		match.publish(executeTime, oldLogsSize)
	}

}
