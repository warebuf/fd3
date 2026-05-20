package match

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

// only handles chain size 1, down the line will change to handle multiple
func (match *Match) handleChainResolution() {

	gs := match.GameState

	if len(gs.Chain) == 0 {
		return
	}

	// pop the top of the chain
	chainTop := gs.Chain[len(gs.Chain)-1]
	gs.Chain = gs.Chain[:len(gs.Chain)-1]

	wallTime := chainTop.ExecuteAtActiveTime + gs.FullPauseOffset
	oldLogsSize := len(match.EventLog)

	// apply effects
	for _, eff := range chainTop.Effects {
		sim.ApplyEffect(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, eff, chainTop.CardID, wallTime)
	}

	// move spell to graveyard
	card := &gs.Cards[chainTop.CardID]
	if card.Zone == state.ZoneChain {
		card.Zone = state.ZoneGraveyard
		gs.Grave[card.Team] = append(gs.Grave[card.Team], chainTop.CardID)
	}

	// log chain item removal
	match.EventLog = append(match.EventLog, logs.RemoveChainLog{
		UpdateID:  match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:      logs.RemoveChain,
		CardID:    chainTop.CardID,
	})
	match.NextUpdateID++

	// if chain is empty, exit sim pause (or advance reveal if in phase 3); otherwise reschedule for next item
	if len(gs.Chain) == 0 {
		if gs.Phase == 3 {
			scheduleReveal(match, wallTime+800)
		} else {
			match.TransitionPauseState(-1, wallTime)
		}
	} else {
		scheduleChainResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
