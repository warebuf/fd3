package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleStructureDraw(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(gs.Structures) {
		return
	}

	structDef := defs.StructureDefs[gs.Structures[teamIndex].DefNo]
	if gs.Mana[teamIndex] < structDef.DrawCost {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	chain.PushStructureDrawOnChain(gs, &match.EventLog, &match.NextUpdateID, teamIndex, executeAtWallTime)

	match.clearChainSkips(executeAtWallTime)
	match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
	scheduleChainResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)
}
