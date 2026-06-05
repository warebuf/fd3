package match

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) handleChainResolution() {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	gs := match.GameState

	if len(gs.Chain) == 0 {
		return
	}

	// pop the top of the chain
	chainTop := gs.Chain[len(gs.Chain)-1]
	gs.Chain = gs.Chain[:len(gs.Chain)-1]

	wallTime := chainTop.ExecuteAtActiveTime + gs.FullPauseOffset
	oldLogsSize := len(match.EventLog)

	now := time.Now()
	fmt.Printf(
		"[chain pop] %s cardID=%d name=%q executeAt=%d wallTime=%d\n",
		now.Format("15:04:05.000"), chainTop.CardID, gs.Cards[chainTop.CardID].Name, chainTop.ExecuteAtActiveTime%1000000, wallTime%1000000)

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
	fmt.Println("removeChainLog", match.NextUpdateID, wallTime, chainTop.CardID)
	match.EventLog = append(match.EventLog, logs.RemoveChainLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:          logs.RemoveChain,
		CardID:        chainTop.CardID,
	})
	match.NextUpdateID++

	if len(gs.Chain) == 0 {
		if gs.Phase == state.PhaseReveal {
			pushRevealEvent(match, wallTime-gs.FullPauseOffset+800)
		} else {
			match.TransitionPauseState(-1, wallTime, -1)
		}
	} else {
		scheduleChainResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
