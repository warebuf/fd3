package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectDestroyAll(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	destroyedIndexes := map[int]bool{}
	for team := 0; team < len(gs.Field); team++ {
		for _, cardIndex := range gs.Field[team] {
			gs.Cards[cardIndex].Zone = state.ZoneGraveyard
			if !gs.Cards[cardIndex].IsToken {
				gs.Grave[team] = append(gs.Grave[team], cardIndex)
			}
			destroyedIndexes[cardIndex] = true
		}
		gs.Field[team] = []int{}
	}
	pq.RemoveSimItemsByCardIndexes(destroyedIndexes)
	for cardIndex := range destroyedIndexes {
		gs.Cards[cardIndex].OnTick = [3]state.TickAbility{}
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
		TargetAll:       eff.A == 0,
		TargetTeamIndex: -1,
		TargetCardIndex: -1,
	})
	*UpdateID++
}
