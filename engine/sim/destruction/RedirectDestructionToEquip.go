package destruction

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// if hostIndex has an equip whose TriggerHostWouldBeDestroyed ability declares EffectAbsorbDestruction,
// destroys that equip instead and reports true; the caller should return without destroying hostIndex itself
func RedirectDestructionToEquip(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, hostIndex int, executeAtWallTime int64) bool {
	hostTeam := gs.Cards[hostIndex].Team
	for _, ci := range gs.Equip[hostTeam] {
		if gs.Cards[ci].EquipHostIndex != hostIndex {
			continue
		}
		if !absorbsHostDestruction(gs.Cards[ci].DefNo) {
			continue
		}
		HandleEquipDestruction(gs, pq, logList, UpdateID, ci, executeAtWallTime)
		return true
	}
	return false
}

func absorbsHostDestruction(defNo int) bool {
	for _, ability := range defs.CardDefs[defNo].OnTrigger {
		if ability.Trigger != state.TriggerHostWouldBeDestroyed {
			continue
		}
		for _, eff := range ability.Effects {
			if eff.Type == state.EffectAbsorbDestruction {
				return true
			}
		}
	}
	return false
}
