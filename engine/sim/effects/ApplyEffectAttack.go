package effects

import (
	"math"
	"math/rand"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectAttack(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, activeIndex int, executeAtWallTime int64) {
	unitIndex := sourceCardIndex
	teamIndex := gs.Cards[unitIndex].Team
	enemyTeam := 1 - teamIndex
	damage := eff.B
	if activeIndex >= 0 {
		damage += gs.Cards[unitIndex].Actives[activeIndex].Bonus
	}

	targetIndex := -1

	// A == 3: always hit the enemy structure directly, skip unit targeting entirely
	if eff.A != 3 {
		enemyField := gs.Field[enemyTeam]

		var tauntTargets []int
		for _, ci := range enemyField {
			if gs.Cards[ci].Taunt {
				tauntTargets = append(tauntTargets, ci)
			}
		}

		// user-targeted: use preselected target if still on the enemy field, but taunt redirects
		if eff.A == 1 {
			preselected := -1
			if activeIndex >= 0 {
				preselected = gs.Cards[unitIndex].Actives[activeIndex].Target
			}
			if preselected != -1 {
				for _, ci := range enemyField {
					if ci == preselected {
						targetIndex = preselected
						break
					}
				}
			}
			if len(tauntTargets) > 0 && (targetIndex == -1 || !gs.Cards[targetIndex].Taunt) {
				targetIndex = tauntTargets[rand.Intn(len(tauntTargets))]
			}
		}

		// random or priority targeting (also fallback for user-targeted with no valid preselect)
		if targetIndex == -1 && len(enemyField) > 0 {
			pool := enemyField
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
					targetIndex = priorityTargets[rand.Intn(len(priorityTargets))]
					damage *= 2
				}
			}
			if targetIndex == -1 {
				targetIndex = pool[rand.Intn(len(pool))]
			}
		}
	}

	if targetIndex >= 0 {
		// allow negative HP but clamp to math.MinInt; wrap-around check avoids overflow in the guard itself
		newHP := gs.Cards[targetIndex].HP - damage
		if damage > 0 && newHP >= gs.Cards[targetIndex].HP {
			newHP = math.MinInt
		}
		gs.Cards[targetIndex].HP = newHP
		*logList = append(*logList, logs.AttackLog{
			UpdateID:        *UpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.Attack,
			AttackerIndex:   unitIndex,
			TargetIsUnit:    true,
			TargetTeamIndex: enemyTeam,
			TargetCardIndex: targetIndex,
			Damage:          damage,
		})
		*UpdateID += 1

		if gs.Cards[targetIndex].HP <= 0 {
			HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, teamIndex, executeAtWallTime)
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
