package effects

import (
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
	if len(enemyField) > 0 {
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
		gs.Cards[targetIdx].HP -= damage
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
			for i, ci := range gs.Field[enemyTeam] {
				if ci == targetIdx {
					gs.Field[enemyTeam] = append(gs.Field[enemyTeam][:i], gs.Field[enemyTeam][i+1:]...)
					break
				}
			}
			gs.Cards[targetIdx].Zone = state.ZoneGraveyard
			gs.Grave[enemyTeam] = append(gs.Grave[enemyTeam], targetIdx)

			pq.RemoveSimItemByCardID(targetIdx)
			gs.Cards[targetIdx].AbilityTimers = []state.AbilityTimer{}

			*logList = append(*logList, logs.DestroyCardLog{
				UpdateID:        *UpdateID,
				WallTimestamp:   executeAtWallTime,
				Type:            logs.DestroyCard,
				TeamIndex:       teamIndex,
				CardIndex:       targetIdx,
				TargetAll:       false,
				TargetTeamIndex: enemyTeam,
				TargetCardIndex: targetIdx,
			})
			*UpdateID += 1
		}
	} else {
		gs.Structures[enemyTeam].HP -= damage
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
