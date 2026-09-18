package effects

import (
	"math"
	"math/rand"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// sums, across any equips attached to unitIndex, AttackBonusPerSummon * how many units unitIndex's
// own controller has ever summoned (still on the field, or already in the graveyard). Equips are
// searched across both teams' Equip lists since an equip can be aimed at the opponent's field, in
// which case it's filed under its caster's team, not its host's.
func equipAttackBonus(gs *state.GameState, unitIndex int) int {
	bonus := 0
	for team := 0; team < len(gs.Equip); team++ {
		for _, ci := range gs.Equip[team] {
			if gs.Cards[ci].EquipHostIndex != unitIndex {
				continue
			}
			perSummon := defs.CardDefs[gs.Cards[ci].DefNo].AttackBonusPerSummon
			if perSummon <= 0 {
				continue
			}
			bonus += perSummon * countUnitsSummoned(gs, gs.Cards[unitIndex].Team)
		}
	}
	return bonus
}

// units teamIndex has ever summoned: still alive on the field, plus ones already in the graveyard
func countUnitsSummoned(gs *state.GameState, teamIndex int) int {
	count := 0
	for _, ci := range gs.Field[teamIndex] {
		if gs.Cards[ci].Type == state.CardTypeUnit {
			count++
		}
	}
	for _, ci := range gs.Grave[teamIndex] {
		if gs.Cards[ci].Type == state.CardTypeUnit {
			count++
		}
	}
	return count
}

// hitStructure reports whether this attack actually landed on an enemy structure (as opposed to a
// unit, or fizzling with nothing to hit at all), so callers like ApplyEffectGraveyardBarrage can
// react to siege damage specifically. overkill reports how much damage exceeded a killed unit's HP
// (0 if the target survived, wasn't a unit, or the destruction was redirected to an equip), so
// callers like ApplyEffectScalingSpillover can carry the excess onto a second target.
func ApplyEffectAttack(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, activeIndex int, executeAtWallTime int64) (hitStructure bool, overkill int) {
	unitIndex := sourceCardIndex
	teamIndex := gs.Cards[unitIndex].Team
	enemyTeam := 1 - teamIndex
	damage := eff.B
	if activeIndex >= 0 {
		damage += gs.Cards[unitIndex].Actives[activeIndex].Bonus
	}
	damage += equipAttackBonus(gs, unitIndex)

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
		targetIndex = state.BodyguardRedirect(gs, targetIndex)
	}

	if targetIndex >= 0 {
		originalHP := gs.Cards[targetIndex].HP

		// allow negative HP but clamp to math.MinInt; wrap-around check avoids overflow in the guard itself
		newHP := originalHP - damage
		if damage > 0 && newHP >= originalHP {
			newHP = math.MinInt
		}

		if newHP <= 0 && destruction.RedirectDestructionToEquip(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime) {
			*logList = append(*logList, logs.AttackLog{
				UpdateID:             *UpdateID,
				WallTimestamp:        executeAtWallTime,
				Type:                 logs.Attack,
				AttackerIndex:        unitIndex,
				TargetTeamIndex:      enemyTeam,
				TargetCardIndex:      targetIndex,
				TargetStructureIndex: -1,
				Damage:               damage,
			})
			*UpdateID += 1
			return false, 0 // the equip absorbed it, no excess to speak of
		}

		gs.Cards[targetIndex].HP = newHP
		*logList = append(*logList, logs.AttackLog{
			UpdateID:             *UpdateID,
			WallTimestamp:        executeAtWallTime,
			Type:                 logs.Attack,
			AttackerIndex:        unitIndex,
			TargetTeamIndex:      enemyTeam,
			TargetCardIndex:      targetIndex,
			TargetStructureIndex: -1,
			Damage:               damage,
		})
		*UpdateID += 1

		if gs.Cards[targetIndex].HP <= 0 {
			destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
			excess := damage - originalHP
			if excess < 0 {
				excess = 0
			}
			return false, excess
		}
		return false, 0 // target survived, nothing excess to spill
	} else {
		// no unit target, so the attack lands on a random living enemy structure
		var living []int
		for i := range gs.Structures[enemyTeam] {
			if gs.Structures[enemyTeam][i].HP > 0 {
				living = append(living, i)
			}
		}
		if len(living) == 0 {
			return false, 0
		}
		structureIndex := living[rand.Intn(len(living))]

		if damage >= gs.Structures[enemyTeam][structureIndex].HP {
			gs.Structures[enemyTeam][structureIndex].HP = 0
		} else {
			gs.Structures[enemyTeam][structureIndex].HP -= damage
		}
		*logList = append(*logList, logs.AttackLog{
			UpdateID:             *UpdateID,
			WallTimestamp:        executeAtWallTime,
			Type:                 logs.Attack,
			AttackerIndex:        unitIndex,
			TargetTeamIndex:      enemyTeam,
			TargetCardIndex:      -1,
			TargetStructureIndex: structureIndex,
			Damage:               damage,
		})
		*UpdateID += 1

		if gs.AllStructuresDead(enemyTeam) {
			if gs.AllStructuresDead(teamIndex) {
				gs.Winner = -2 // tie
			} else {
				gs.Winner = teamIndex
			}
		}
		return true, 0 // structure hits don't cascade further
	}
}
