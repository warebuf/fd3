package destruction

import (
	"math"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// fires cardIndex's own TriggerSelfDestroyed OnTrigger ability, if it has one. Applied immediately
// rather than through the chain: by this point the source card has already reached the graveyard,
// which would fail chainSourceStillValid's CardTypeUnit check (it requires Zone == ZoneField), so a
// self-death reaction can't be modeled as a normal reactable chain item the way other triggers are.
func fireSelfDestroyedTrigger(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	card := gs.Cards[cardIndex]
	def := defs.CardDefs[card.DefNo]

	for _, ability := range def.OnTrigger {
		if ability.Trigger != state.TriggerSelfDestroyed || ability.TriggerZones&defs.TriggerZoneGraveyard == 0 {
			continue
		}
		for _, eff := range ability.Effects {
			if eff.Type == state.EffectDamageAllAllies {
				damageAllAllies(gs, pq, logList, UpdateID, card.Team, cardIndex, eff.B, executeAtWallTime)
			}
		}
	}
}

// deals amount damage to every unit currently on teamIndex's field, recursing into
// HandleUnitDestruction for any that drop to 0 HP or below
func damageAllAllies(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, sourceCardIndex int, amount int, executeAtWallTime int64) {
	if amount <= 0 {
		return
	}

	targets := append([]int{}, gs.Field[teamIndex]...) // snapshot: HandleUnitDestruction below mutates gs.Field[teamIndex]
	for _, targetIndex := range targets {
		target := &gs.Cards[targetIndex]

		newHP := target.HP - amount
		if newHP >= target.HP { // guards against int overflow wraparound
			newHP = math.MinInt
		}
		target.HP = newHP

		*logList = append(*logList, logs.AttackLog{
			UpdateID:             *UpdateID,
			WallTimestamp:        executeAtWallTime,
			Type:                 logs.Attack,
			AttackerIndex:        sourceCardIndex,
			TargetTeamIndex:      teamIndex,
			TargetCardIndex:      targetIndex,
			TargetStructureIndex: -1,
			Damage:               amount,
		})
		*UpdateID++

		if target.HP <= 0 {
			HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
		}
	}
}
