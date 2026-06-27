package match

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

const (
	DeclareErrBadInput      = 0
	DeclareChainStarted     = 1
	DeclareConditionsNotMet = 2
)

func ScheduleDeclareCard(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, wallTimeAtPop int64) int {

	if gs.Cards[cardIndex].Zone != state.ZoneHand {
		return DeclareErrBadInput
	}
	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return DeclareErrBadInput
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return DeclareErrBadInput
	}

	if gs.Phase != state.PhaseReveal && gs.Phase != state.PhasePlay {
		return DeclareErrBadInput
	}

	if defs.CardDefs[gs.Cards[cardIndex].DefNo].RequiresStackTarget && len(gs.Stack) == 0 {
		return DeclareErrBadInput
	}

	// mana check last
	cost := gs.Cards[cardIndex].Cost
	if gs.Mana[teamIndex] < cost {
		return DeclareConditionsNotMet
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
	if gs.Cards[cardIndex].Type == state.CardTypeUnit {
		gs.Cards[cardIndex].Status = state.CardStatusSummoning
	}
	gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
	addCardToStack(gs, logList, UpdateID, teamIndex, cardIndex, wallTimeAtPop+constants.LEAD)
	return DeclareChainStarted

}


func ScheduleDrawCard(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, executeAtWallTime int64) {
	executeAtSimTime := executeAtWallTime - gs.SimPauseOffset

	item := &state.SimulationItem{
		StartAtSimTime:   executeAtSimTime,
		PopAtSimTime:     executeAtSimTime - constants.LEAD,
		ExecuteAtSimTime: executeAtSimTime,
		AbilityID:        -1,
		CardID:           -1,
		Type:             state.SimItemDraw,
		Effects:          []state.Effect{{Type: state.EffectDraw, A: teamIndex, B: 1}},
	}
	heap.Push(pq, item)
}


func addCardToStack(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, declarationTime int64) {
	card := gs.Cards[cardIndex]

	stackItemType := state.StackItemSpell
	if card.Type == state.CardTypeUnit {
		stackItemType = state.StackItemSummon
	}

	abilities := defs.CardDefs[card.DefNo].OnPlay
	effs := make([]state.Effect, len(abilities))
	for i, ability := range abilities {
		effs[i] = sim.FulfillEffectTemplate(ability, teamIndex, cardIndex)
	}

	startAtActiveTime := declarationTime - gs.FullPauseOffset
	stackItem := &state.StackItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardID:              cardIndex,
		Type:                stackItemType,
		Effects:             effs,
	}
	gs.Stack = append(gs.Stack, *stackItem)

	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       declarationTime,
		Type:                logs.AddStack,
		StackType:           stackItem.Type,
		StartAtActiveTime:   stackItem.StartAtActiveTime,
		PopAtActiveTime:     stackItem.PopAtActiveTime,
		ExecuteAtActiveTime: stackItem.ExecuteAtActiveTime,
		CardID:              stackItem.CardID,
		Effects:             stackItem.Effects,
	})
	*UpdateID += 1
}
