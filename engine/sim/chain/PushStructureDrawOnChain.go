package chain

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// caller is responsible for checking the team can afford the structure's DrawCost before calling
func PushStructureDrawOnChain(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, executeAtWallTime int64) {
	structDef := defs.StructureDefs[gs.Structures[teamIndex].DefNo]

	gs.Mana[teamIndex] -= structDef.DrawCost
	*logList = append(*logList, logs.SpendManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        structDef.DrawCost,
	})
	*UpdateID++

	eff := state.FulfillEffectTemplate(gs, state.Effect{Type: state.EffectDraw, B: 1}, teamIndex, -1, -1)

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	chainItem := state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           -1,
		ActiveIndex:         -1,
		TeamIndex:           teamIndex,
		Type:                state.ChainStructureAbility,
		Effects:             []state.Effect{eff},
	}
	gs.Chain = append(gs.Chain, chainItem)
	*logList = append(*logList, logs.AddChainLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddChain,
		ChainType:           int(chainItem.Type),
		StartAtActiveTime:   chainItem.StartAtActiveTime,
		PopAtActiveTime:     chainItem.PopAtActiveTime,
		ExecuteAtActiveTime: chainItem.ExecuteAtActiveTime,
		CardIndex:           chainItem.CardIndex,
		ActiveIndex:         chainItem.ActiveIndex,
		TeamIndex:           chainItem.TeamIndex,
		Effects:             chainItem.Effects,
	})
	*UpdateID++

	gs.Priority = 1 - teamIndex
}
