package match

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) handleStackResolution() {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	gs := match.GameState

	if len(gs.Stack) == 0 {
		return
	}

	// pop the top of the stack
	stackTop := gs.Stack[len(gs.Stack)-1]
	gs.Stack = gs.Stack[:len(gs.Stack)-1]

	wallTime := stackTop.ExecuteAtActiveTime + gs.FullPauseOffset
	oldLogsSize := len(match.EventLog)

	fmt.Printf("[stack pop] %s cardID=%d name=%q executeAt=%d wallTime=%d\n", time.Now().Format("15:04:05.000"), stackTop.CardID, gs.Cards[stackTop.CardID].Name, stackTop.ExecuteAtActiveTime%1000000, wallTime%1000000)

	// apply effects
	for _, eff := range stackTop.Effects {
		sim.ApplyEffect(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, eff, stackTop.CardID, wallTime)
	}

	// move spell to graveyard
	card := &gs.Cards[stackTop.CardID]
	if card.Zone == state.ZoneStack {
		card.Zone = state.ZoneGraveyard
		gs.Grave[card.Team] = append(gs.Grave[card.Team], stackTop.CardID)
	}

	match.EventLog = append(match.EventLog, logs.RemoveStackLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:          logs.RemoveStack,
		CardID:        stackTop.CardID,
	})
	match.NextUpdateID++

	if len(gs.Stack) == 0 {
		if gs.Phase == state.PhaseReveal {
			pushRevealEvent(match, wallTime-gs.FullPauseOffset+800)
		} else {
			match.TransitionPauseState(-1, wallTime, -1)
		}
	} else { // the new top gets a refreshed reaction window
		newTopIdx := len(gs.Stack) - 1
		freshStart := wallTime - gs.FullPauseOffset
		gs.Stack[newTopIdx].StartAtActiveTime = freshStart
		gs.Stack[newTopIdx].PopAtActiveTime = freshStart + constants.REACTION_WINDOW - constants.LEAD
		gs.Stack[newTopIdx].ExecuteAtActiveTime = freshStart + constants.REACTION_WINDOW

		match.EventLog = append(match.EventLog, logs.RefreshStackLog{
			UpdateID:            match.NextUpdateID,
			WallTimestamp:       wallTime,
			Type:                logs.RefreshStack,
			CardID:              gs.Stack[newTopIdx].CardID,
			StartAtActiveTime:   gs.Stack[newTopIdx].StartAtActiveTime,
			PopAtActiveTime:     gs.Stack[newTopIdx].PopAtActiveTime,
			ExecuteAtActiveTime: gs.Stack[newTopIdx].ExecuteAtActiveTime,
		})
		match.NextUpdateID++

		scheduleStackResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
