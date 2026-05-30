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
	case "hand":
		match.handleInputHand(ie)
	case "undoHand":
		match.handleInputUndoHand(ie)
	case "ready":
		match.handleInputReady(ie)
	}

}

func (match *Match) handleInputHand(ie *eventpq.InputEvent) {

	// do not accept if we are not in phase 2 (board set up) or 4 (play)
	if match.GameState.Phase != state.PhaseBoardSetup && match.GameState.Phase != state.PhasePlay {
		return
	}

	cardIndex := ie.CardIndex
	teamIdx := ie.TeamIdx
	if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
		return
	}

	card := match.GameState.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIdx {
		return
	}

	// we are denying all inputs during a chain/reaction, but later on we should only accept chains
	if match.GameState.Phase == state.PhasePlay && match.GameState.PausePhase != state.PausePhaseSimulating {
		return
	}

	if match.GameState.Phase == state.PhaseBoardSetup { // if we are in phase 2 (board set up), stage the card

		if card.AutoSummon != 0 {
			return
		}

		// stage the card
		nextAutoSummon := 1
		for _, c := range match.GameState.Cards {
			if c.Team == teamIdx && c.AutoSummon >= nextAutoSummon {
				nextAutoSummon = c.AutoSummon + 1
			}
		}
		match.GameState.Cards[cardIndex].AutoSummon = nextAutoSummon

		oldLogsSize := len(match.EventLog)
		match.EventLog = append(match.EventLog, logs.AutoSummonLog{
			WallTimestamp: ie.WallTime.UnixMilli(),
			UpdateID:      match.NextUpdateID,
			Type:          logs.AutoSummonIndex,
			TeamIndex:     teamIdx,
			CardIndex:     cardIndex,
			AutoSummonIdx: nextAutoSummon,
		})
		match.NextUpdateID++
		match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

	} else if match.GameState.Phase == state.PhasePlay { // if we are in phase 4 (play), declare the card

		oldLogsSize := len(match.EventLog)
		chainStarted := ScheduleDeclareCard(match.GameState, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, ie.TeamIdx, ie.CardIndex, ie.WallTime.UnixMilli())
		if chainStarted {
			match.TransitionPauseState(1, ie.WallTime.UnixMilli()+constants.LEAD, -1)
			scheduleChainResolution(match)
		} else {
			scheduleEarliestSimItem(match)
		}
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)

	}

}

func (match *Match) handleInputUndoHand(ie *eventpq.InputEvent) {

	// only accept if we are in phase 2 (board set up)
	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	cardIndex := ie.CardIndex
	teamIdx := ie.TeamIdx

	// do not accept a card that does not exist
	if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
		return
	}

	// card must be in hand, must match the user's team, must be staged
	card := match.GameState.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIdx || card.AutoSummon == 0 {
		return
	}

	// unstage the card, shift all subsequent staged cards
	cleared := match.GameState.Cards[cardIndex].AutoSummon
	match.GameState.Cards[cardIndex].AutoSummon = 0
	for i := range match.GameState.Cards {
		if match.GameState.Cards[i].Team == teamIdx && match.GameState.Cards[i].AutoSummon > cleared {
			match.GameState.Cards[i].AutoSummon--
		}
	}

	// create the log, update game state, broadcast the log
	oldLogsSize := len(match.EventLog)
	match.EventLog = append(match.EventLog, logs.AutoSummonLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.AutoSummonIndex,
		TeamIndex:     teamIdx,
		CardIndex:     cardIndex,
		AutoSummonIdx: 0,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

}

func (match *Match) handleInputReady(ie *eventpq.InputEvent) {

	// only accept if we are in stage 2 (board set up)
	if match.GameState.Phase != state.PhaseBoardSetup {
		return
	}

	// do not accept while in full pause
	if match.GameState.PausePhase == state.PausePhaseFullPause {
		return
	}

	// do not accept if not acceptable team index
	teamIdx := ie.TeamIdx
	if teamIdx < 0 || teamIdx >= len(match.GameState.Ready) {
		return
	}

	// if the player is already ready, do not accept
	if match.GameState.Ready[teamIdx] {
		return
	}

	// update ready arr
	match.GameState.Ready[teamIdx] = true

	// create the log and broadcast
	oldLogsSize := len(match.EventLog)
	match.EventLog = append(match.EventLog, logs.ReadyLog{
		WallTimestamp: ie.WallTime.UnixMilli(),
		UpdateID:      match.NextUpdateID,
		Type:          logs.Ready,
		TeamIndex:     teamIdx,
	})
	match.NextUpdateID++
	match.publish(ie.WallTime.UnixMilli(), oldLogsSize)

	// if all ready, increment the phase to board set up
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

func (match *Match) handleInputPause(ie *eventpq.InputEvent) {

	gs := match.GameState
	teamIdx := ie.TeamIdx

	executeTime := ie.WallTime.UnixMilli() + constants.LEAD

	if gs.Phase < state.PhaseDraw {
		return
	}
	if gs.Phase == state.PhaseGameOver {
		return
	}
	if teamIdx < 0 || teamIdx >= len(gs.TimeBank) {
		return
	}

	if gs.PausePhase == state.PausePhaseSimulating || gs.PausePhase == state.PausePhaseSimPause {
		// entering pause
		if gs.TimeBank[teamIdx] <= 0 {
			return
		}
		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(state.PausePhaseFullPause, executeTime, teamIdx)

		pauseEnd := &eventpq.PauseEndEvent{WallTime: time.UnixMilli(executeTime + gs.TimeBank[teamIdx] - constants.LEAD)}
		match.PauseEndItem = match.EventPQueue.CreateAndPushEventItem(pauseEnd, "PauseEndEvent")

		match.publish(executeTime, oldLogsSize)

	} else if gs.PausePhase == state.PausePhaseFullPause && gs.TeamPaused == teamIdx { // early unpause by the same team

		if match.PauseEndItem != nil {
			match.EventPQueue.Remove(match.PauseEndItem)
			match.PauseEndItem = nil
		}

		oldLogsSize := len(match.EventLog)
		match.TransitionPauseState(-1, executeTime, teamIdx)
		match.publish(executeTime, oldLogsSize)
	}

}
