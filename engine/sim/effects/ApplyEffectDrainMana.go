package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions/autoability"
	"github.com/warebuf/fd3/engine/state"
)

// eff.A is the draining team (auto-filled), eff.B is the amount requested. Takes at most what the
// opponent actually has, so this never pushes their mana below zero, then grants that (possibly
// smaller) amount to the draining team.
func ApplyEffectDrainMana(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	drainingTeam := eff.A
	opponentTeam := 1 - drainingTeam

	amount := eff.B
	if amount > gs.Mana[opponentTeam] {
		amount = gs.Mana[opponentTeam]
	}
	if amount <= 0 {
		return
	}

	gs.Mana[opponentTeam] -= amount
	*logList = append(*logList, logs.SpendManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     opponentTeam,
		Amount:        amount,
	})
	*UpdateID++

	gs.Mana[drainingTeam] += amount
	*logList = append(*logList, logs.GainManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GainMana,
		TeamIndex:     drainingTeam,
		Amount:        amount,
	})
	*UpdateID++

	cardIndexes, activeIndexes := autoability.AutoAbilityManaCheck(gs, drainingTeam)
	for i := 0; i < len(cardIndexes); i++ {
		chain.PushAbilityOnChain(gs, logList, UpdateID, cardIndexes[i], activeIndexes[i], executeAtWallTime)
	}

	structureIndexes := autoability.AutoAbilityStructureDrawCheck(gs, drainingTeam)
	for i := 0; i < len(structureIndexes); i++ {
		chain.PushStructureDrawOnChain(gs, logList, UpdateID, drainingTeam, structureIndexes[i], executeAtWallTime)
	}
}
