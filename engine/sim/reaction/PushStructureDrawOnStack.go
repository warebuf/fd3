package reaction

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// caller is responsible for checking the team can afford the structure's DrawCost before calling
func PushStructureDrawOnStack(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, executeAtWallTime int64) {
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
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           -1,
		ActiveIndex:         -1,
		TeamIndex:           teamIndex,
		Type:                state.ReactionStructureAbility,
		Effects:             []state.Effect{eff},
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		ActiveIndex:         reactionItem.ActiveIndex,
		Effects:             reactionItem.Effects,
	})
	*UpdateID++
}
