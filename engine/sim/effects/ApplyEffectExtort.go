package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions/autoability"
	"github.com/warebuf/fd3/engine/state"
)

// eff.A is the team that just paid mana to activate something (auto-filled by FulfillEffectTemplate),
// eff.B is the toll. If the paying team can afford the toll it's taken from them; either way the
// source card's controller collects the same amount, so they profit whether or not the payment lands
func ApplyEffectExtort(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	payingTeam := eff.A
	amount := eff.B
	if amount <= 0 {
		return
	}
	controllerTeam := gs.Cards[sourceCardIndex].Team

	if gs.Mana[payingTeam] >= amount {
		gs.Mana[payingTeam] -= amount
		*logList = append(*logList, logs.SpendManaLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.SpendMana,
			TeamIndex:     payingTeam,
			Amount:        amount,
		})
		*UpdateID++
	}

	gs.Mana[controllerTeam] += amount
	*logList = append(*logList, logs.GainManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GainMana,
		TeamIndex:     controllerTeam,
		Amount:        amount,
	})
	*UpdateID++

	cardIndexes, activeIndexes := autoability.AutoAbilityManaCheck(gs, controllerTeam)
	for i := 0; i < len(cardIndexes); i++ {
		chain.PushAbilityOnChain(gs, logList, UpdateID, cardIndexes[i], activeIndexes[i], executeAtWallTime)
	}

	structureIndexes := autoability.AutoAbilityStructureDrawCheck(gs, controllerTeam)
	for i := 0; i < len(structureIndexes); i++ {
		chain.PushStructureDrawOnChain(gs, logList, UpdateID, controllerTeam, structureIndexes[i], executeAtWallTime)
	}
}
