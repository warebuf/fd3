package sim

import (
	"fmt"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/sim/effects"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffect(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, activeIndex int, executeAtWallTime int64) {
	switch eff.Type {

	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectDestroy:
		destroyedIndexes := map[int]bool{}
		for a := 0; a < len(gs.Field); a++ {
			for b := 0; b < len(gs.Field[a]); b++ {
				cardIndex := gs.Field[a][b]
				gs.Cards[cardIndex].Zone = state.ZoneGraveyard
				gs.Grave[a] = append(gs.Grave[a], cardIndex)
				destroyedIndexes[cardIndex] = true
			}
			gs.Field[a] = []int{}
		}
		pq.RemoveSimItemsByCardIndexes(destroyedIndexes)
		for cardIndexes := range destroyedIndexes {
			gs.Cards[cardIndexes].OnTick = [3]state.TickAbility{}
		}
		// clear any preselected targets pointing at destroyed cards
		for i := range gs.Cards {
			for j := range gs.Cards[i].Actives {
				if destroyedIndexes[gs.Cards[i].Actives[j].Target] {
					gs.Cards[i].Actives[j].Target = -1
				}
			}
		}
		*logList = append(*logList, logs.DestroyCardLog{
			UpdateID:        *UpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.DestroyCard,
			TeamIndex:       gs.Cards[eff.C].Team,
			CardIndex:       eff.C,
			TargetAll:       eff.A == 0,
			TargetTeamIndex: -1,
			TargetCardIndex: -1,
		})
		*UpdateID += 1

	case state.EffectMana:
		gs.Mana[eff.A] += eff.B
		*logList = append(*logList, logs.GainManaLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.GainMana,
			TeamIndex:     eff.A,
			Amount:        eff.B,
		})
		*UpdateID += 1

		cardIndexes, activeIndexes := conditions.AutoAbilityManaCheck(gs, eff.A)
		for i := 0; i < len(cardIndexes); i++ {
			pushAbilityOnStack(gs, logList, UpdateID, cardIndexes[i], activeIndexes[i], executeAtWallTime)
		}

	case state.EffectAttack:
		effects.ApplyEffectAttack(gs, pq, logList, UpdateID, eff, sourceCardIndex, activeIndex, executeAtWallTime)

	case state.EffectHeal:
		effects.ApplyEffectHeal(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectCounter:
		effects.ApplyEffectCounter(gs, logList, UpdateID, sourceCardIndex, executeAtWallTime)

	case state.EffectGainCharge:
		effects.ApplyEffectGainCharge(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

		if conditions.AutoAbilityChargeCheck(gs, sourceCardIndex, eff.A) {
			pushAbilityOnStack(gs, logList, UpdateID, sourceCardIndex, eff.A, executeAtWallTime)
		}

	default:
		fmt.Println("unhandled effect type:", eff.Type)
	}
}

func FulfillEffectTemplate(templ state.Effect, teamIndex int, cardIndex int) state.Effect {
	switch templ.Type {
	case state.EffectDestroy:
		templ.C = cardIndex
		return templ
	case state.EffectMana:
		templ.A = teamIndex
		return templ
	case state.EffectDraw:
		templ.A = teamIndex
		return templ
	case state.EffectAttack:
		return templ
	case state.EffectHeal:
		return templ
	}
	return templ
}
