package effects

import (
	"math"
	"math/rand"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectAttack(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardID int, executeAtWallTime int64) {
	unitIndex := sourceCardID
	teamIndex := gs.Cards[unitIndex].Team
	enemyTeam := 1 - teamIndex
	damage := eff.B

	enemyField := gs.Field[enemyTeam]

	targetIdx := -1

	// user-targeted: use preselected target if still on the enemy field
	if eff.A == 1 {
		preselected := gs.Cards[unitIndex].Targets[0]
		if preselected != -1 {
			for _, ci := range enemyField {
				if ci == preselected {
					targetIdx = preselected
					break
				}
			}
		}
	}

	// random or priority targeting (also fallback for user-targeted with no valid preselect)
	if targetIdx == -1 && len(enemyField) > 0 {
		pool := enemyField
		var tauntTargets []int
		for _, ci := range enemyField {
			if gs.Cards[ci].Taunt {
				tauntTargets = append(tauntTargets, ci)
			}
		}
		if len(tauntTargets) > 0 {
			pool = tauntTargets
		}

		if eff.A == 2 {
			var priorityTargets []int
			for _, ci := range pool {
				if gs.Cards[ci].Archetype == eff.C {
					priorityTargets = append(priorityTargets, ci)
				}
			}
			if len(priorityTargets) > 0 {
				targetIdx = priorityTargets[rand.Intn(len(priorityTargets))]
				damage *= 2
			}
		}
		if targetIdx == -1 {
			targetIdx = pool[rand.Intn(len(pool))]
		}
	}

	if targetIdx >= 0 {
		if damage >= gs.Cards[targetIdx].HP {
			gs.Cards[targetIdx].HP = 0
		} else {
			gs.Cards[targetIdx].HP -= damage
		}
		*logList = append(*logList, logs.AttackLog{
			UpdateID:        *UpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.Attack,
			AttackerIndex:   unitIndex,
			TargetIsUnit:    true,
			TargetTeamIndex: enemyTeam,
			TargetCardIndex: targetIdx,
			Damage:          damage,
		})
		*UpdateID += 1

		if gs.Cards[targetIdx].HP <= 0 {
			HandleUnitDestruction(gs, pq, logList, UpdateID, targetIdx, teamIndex, executeAtWallTime)
		}
	} else {
		if damage >= gs.Structures[enemyTeam].HP {
			gs.Structures[enemyTeam].HP = 0
		} else {
			gs.Structures[enemyTeam].HP -= damage
		}
		*logList = append(*logList, logs.AttackLog{
			UpdateID:        *UpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.Attack,
			AttackerIndex:   unitIndex,
			TargetIsUnit:    false,
			TargetTeamIndex: enemyTeam,
			TargetCardIndex: -1,
			Damage:          damage,
		})
		*UpdateID += 1

		if gs.Structures[enemyTeam].HP <= 0 {
			if gs.Structures[teamIndex].HP <= 0 {
				gs.Winner = -2 // tie
			} else {
				gs.Winner = teamIndex
			}
		}
	}
}
