package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleStructureDraw(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	teamIndex := ie.TeamIndex
	if teamIndex < 0 || teamIndex >= len(gs.Structures) {
		return
	}

	structDef := defs.StructureDefs[gs.Structures[teamIndex].DefNo]
	if gs.Mana[teamIndex] < structDef.DrawCost {
		return
	}

	eff := state.FulfillEffectTemplate(gs, state.Effect{Type: state.EffectDraw, B: 1}, teamIndex, -1, -1)

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	gs.Mana[teamIndex] -= structDef.DrawCost
	match.EventLog = append(match.EventLog, logs.SpendManaLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        structDef.DrawCost,
	})
	match.NextUpdateID++

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
	match.EventLog = append(match.EventLog, logs.AddStackLog{
		UpdateID:            match.NextUpdateID,
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
	match.NextUpdateID++

	match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
	scheduleReactionStackResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)
}
