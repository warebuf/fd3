package destruction

import "github.com/warebuf/fd3/engine/state"

// if hostIndex has a card equipped to it, destroys that equip instead and reports true;
// the caller should return without destroying hostIndex itself
func RedirectDestructionToEquip(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, hostIndex int, executeAtWallTime int64) bool {
	hostTeam := gs.Cards[hostIndex].Team
	for _, ci := range gs.Equip[hostTeam] {
		if gs.Cards[ci].EquipHostIndex == hostIndex {
			HandleEquipDestruction(gs, pq, logList, UpdateID, ci, executeAtWallTime)
			return true
		}
	}
	return false
}
