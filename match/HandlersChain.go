package match

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/randomFunctions"
	"github.com/warebuf/fd3/engine/state"
)

// right now chain size should only be 1
// we have to edit this file down the line to handle multiple chains

// NOTE VERY INCOMPLETE, JUST A WIP

func (match *Match) handleChainResolution() {

	gs := match.GameState

	if len(gs.Chain) == 0 {
		return
	}

	// pop the top of the chain
	chainTop := gs.Chain[len(gs.Chain)-1]
	gs.Chain = gs.Chain[:len(gs.Chain)-1]
	match.ChainItem = nil

	wallTime := chainTop.OccurTime + gs.FullPauseOffset
	oldLogsSize := len(match.EventLog)

	// apply effects
	for _, eff := range chainTop.Effects {
		randomFunctions.ApplyEffect(gs, &match.EventLog, &match.NextUpdateID, eff, chainTop.CardID, wallTime)
	}

	// move spell to graveyard
	card := &gs.Cards[chainTop.CardID]
	if card.Zone == state.ZoneChain {
		card.Zone = state.ZoneGraveyard
		gs.Grave[card.Team] = append(gs.Grave[card.Team], chainTop.CardID)
	}

	// log chain item removal
	match.EventLog = append(match.EventLog, logs.RemoveSimItemLog{
		UpdateID:  match.NextUpdateID,
		Timestamp: wallTime,
		Type:      logs.RemoveChain,
		CardID:    chainTop.CardID,
	})
	match.NextUpdateID++

	// if chain is empty, exit sim pause; otherwise reschedule for next item
	if len(gs.Chain) == 0 {
		match.TransitionPauseState(-1, wallTime)
	} else {
		scheduleChainResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
