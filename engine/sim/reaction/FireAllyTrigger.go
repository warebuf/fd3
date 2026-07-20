package reaction

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/state"
)

func FireAllyTrigger(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, trigger state.TriggerType, executeAtWallTime int64) {
	cardIndexes, triggerIndexes := conditions.CheckAllyTriggers(gs, teamIndex, trigger)

	for i := range cardIndexes {
		cardIndex := cardIndexes[i]
		ability := defs.CardDefs[gs.Cards[cardIndex].DefNo].OnTrigger[triggerIndexes[i]]

		effs := []state.Effect{}
		for _, eff := range ability.Effects {
			if eff.Type == 0 {
				continue
			}
			effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, -1))
		}
		PushEffectsOnStack(gs, logList, UpdateID, cardIndex, effs, executeAtWallTime)
	}
}
