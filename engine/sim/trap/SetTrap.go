package trap

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
)

// setting never touches the chain, so there's no window to open and no mana to spend yet;
// the trap's cost is paid when it's activated
func SetTrap(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, executeAtWallTime int64) bool {

	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return false
	}

	card := gs.Cards[cardIndex]
	if card.Type != state.CardTypeTrap {
		return false
	}
	if card.Zone != state.ZoneHand || card.Team != teamIndex {
		return false
	}

	if gs.Phase != state.PhasePlay {
		return false
	}

	// state is frozen while a chain window is open, so a set has to wait for it to close
	if gs.PausePhase != state.PausePhaseSimulating {
		return false
	}

	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return false
	}

	gs.Cards[cardIndex].AutoSummon = 0

	gs.Cards[cardIndex].Zone = state.ZoneTrap
	gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
	gs.Traps[teamIndex] = append(gs.Traps[teamIndex], cardIndex)

	schedule.ScheduleTrapReady(gs, pq, cardIndex, executeAtWallTime-gs.SimPauseOffset)

	*logList = append(*logList, logs.SetTrapLog{
		UpdateID:          *UpdateID,
		WallTimestamp:     executeAtWallTime,
		Type:              logs.SetTrap,
		TeamIndex:         teamIndex,
		CardIndex:         cardIndex,
		SetStartAtSimTime: gs.Cards[cardIndex].SetStartAtSimTime,
		ReadyAtSimTime:    gs.Cards[cardIndex].ReadyAtSimTime,
	})
	*UpdateID++

	return true
}
