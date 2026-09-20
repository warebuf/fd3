package effects

import (
	"math"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// deals eff.B damage to every unit on the source's team's field, destroying any that drop to 0 HP or below
func ApplyEffectDamageAllAllies(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	amount := eff.B
	if amount <= 0 {
		return
	}
	teamIndex := gs.Cards[sourceCardIndex].Team

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
			destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
		}
	}
}
