package match

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDeclareCard(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, wallTimeAtPop int64) bool {

	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return false
	}

	if gs.Cards[cardIndex].Zone != state.ZoneHand {
		return false
	}
	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return false
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return false
	}

	if gs.Phase != state.PhaseReveal && gs.Phase != state.PhasePlay {
		return false
	}

	if defs.CardDefs[gs.Cards[cardIndex].DefNo].RequiresStackTarget && len(gs.ReactionStack) == 0 {
		return false
	}

	// mana check last
	cost := gs.Cards[cardIndex].Cost
	if gs.Mana[teamIndex] < cost {
		return false
	}

	gs.Mana[teamIndex] -= cost
	*logList = append(*logList, logs.SpendManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: wallTimeAtPop + constants.LEAD,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        cost,
	})
	*UpdateID += 1

	*logList = append(*logList, logs.DeclareCardLog{
		UpdateID:      *UpdateID,
		WallTimestamp: wallTimeAtPop + constants.LEAD,
		Type:          logs.DeclareCard,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
	})
	*UpdateID += 1

	gs.Cards[cardIndex].StageOrder = 0
	gs.Cards[cardIndex].AutoSummon = 0

	gs.Cards[cardIndex].Zone = state.ZoneStack
	gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
	addCardToStack(gs, logList, UpdateID, teamIndex, cardIndex, wallTimeAtPop+constants.LEAD)
	return true

}

func ScheduleSystemDraw(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, executeAtWallTime int64) {
	executeAtSimTime := executeAtWallTime - gs.SimPauseOffset

	item := &state.SimulationItem{
		StartAtSimTime:   executeAtSimTime,
		PopAtSimTime:     executeAtSimTime - constants.LEAD,
		ExecuteAtSimTime: executeAtSimTime,
		AbilityIndex:     -1,
		CardIndex:        -1,
		TeamIndex:        teamIndex,
		Type:             state.SimItemSystemDraw,
		Effects:          []state.Effect{{Type: state.EffectDraw, A: teamIndex, B: 1}},
	}
	heap.Push(pq, item)
}

func addCardToStack(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, declarationTime int64) {
	card := gs.Cards[cardIndex]

	reactionType := state.ReactionSpell
	if card.Type == state.CardTypeUnit {
		reactionType = state.ReactionSummon
	}

	var effs []state.Effect
	for _, ability := range defs.CardDefs[card.DefNo].OnPlay {
		for _, eff := range ability.Effects {
			if eff.Type == 0 {
				continue // unused Effect slot
			}
			effs = append(effs, state.FulfillEffectTemplate(eff, teamIndex, cardIndex))
		}
	}

	startAtActiveTime := declarationTime - gs.FullPauseOffset
	reactionItem := &state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         -1, // OnPlay effects aren't Active-ability-sourced
		Type:                reactionType,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, *reactionItem)

	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       declarationTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	*UpdateID += 1
}
