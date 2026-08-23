package destruction

import "github.com/warebuf/fd3/engine/state"

// destroys any equips still attached to hostIndex, e.g. after it's actually destroyed
func DestroyEquipsForHost(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, hostIndex int, executeAtWallTime int64) {
	hostTeam := gs.Cards[hostIndex].Team
	snapshot := append([]int{}, gs.Equip[hostTeam]...)
	for _, ci := range snapshot {
		if gs.Cards[ci].EquipHostIndex == hostIndex {
			HandleEquipDestruction(gs, pq, logList, UpdateID, ci, executeAtWallTime)
		}
	}
}
