package match

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

// ScheduleDeclareCard returns true if a chain was started (caller should enter sim pause).
func ScheduleDeclareCard(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, wallTimeAtPop int64) bool {

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

	cost := gs.Cards[cardIndex].Cost
	if gs.Mana[teamIndex] < cost {
		return false
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return false
	} else if gs.PausePhase == state.PausePhaseSimPause && gs.Phase != state.PhaseReveal && gs.Phase != state.PhasePlay { // cards cannot be declared in Sim Pause unless they are counters
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

	gs.Cards[cardIndex].AutoSummon = 0

	if gs.Cards[cardIndex].Type == state.CardTypeUnit {
		gs.Cards[cardIndex].Zone = state.ZoneField
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
		gs.Field[teamIndex] = append(gs.Field[teamIndex], cardIndex)
		scheduleSummon(gs, pq, abilityIDCounter, logList, UpdateID, teamIndex, cardIndex, wallTimeAtPop)
	} else if gs.Cards[cardIndex].Type == state.CardTypeSpell {
		gs.Cards[cardIndex].Zone = state.ZoneChain
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
		scheduleChain(gs, logList, UpdateID, teamIndex, cardIndex, wallTimeAtPop+constants.LEAD)
		return true
	}

	return false
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

func scheduleSummon(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, wallTimeAtPop int64) {
	eff := state.Effect{
		Type: state.SummonResolution,
		A:    teamIndex,
		B:    cardIndex,
	}

	simTime := wallTimeAtPop - gs.SimPauseOffset

	abilityID := *abilityIDCounter
	*abilityIDCounter++

	simItem := &state.SimulationItem{
		StartAtSimTime:   simTime + constants.LEAD,
		PopAtSimTime:     simTime + constants.REACTION_WINDOW,
		ExecuteAtSimTime: simTime + constants.LEAD + constants.REACTION_WINDOW,

		AbilityID: abilityID,
		CardID:    cardIndex,
		Type:      state.SimItemSummonResolve,
		Effects:   []state.Effect{eff},
	}
	heap.Push(pq, simItem)

	gs.Cards[cardIndex].AbilityTimers = append(gs.Cards[cardIndex].AbilityTimers, state.AbilityTimer{
		AbilityID:        abilityID,
		AbilityIdx:       -1,
		Type:             state.AbilityTimerSummon,
		StartAtSimTime:   simItem.StartAtSimTime,
		ExecuteAtSimTime: simItem.ExecuteAtSimTime,
	})
	*logList = append(*logList, logs.AddAbilityTimerLog{
		UpdateID:         *UpdateID,
		WallTimestamp:    wallTimeAtPop + constants.LEAD,
		Type:             logs.AddAbilityTimer,
		CardIndex:        cardIndex,
		AbilityID:        abilityID,
		AbilityIdx:       -1,
		TimerType:        state.AbilityTimerSummon,
		StartAtSimTime:   simItem.StartAtSimTime,
		ExecuteAtSimTime: simItem.ExecuteAtSimTime,
	})
	*UpdateID++
}

func scheduleChain(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, declarationTime int64) {
	card := gs.Cards[cardIndex]

	abilities := defs.CardDefs[card.DefNo].OnPlay
	effs := make([]state.Effect, len(abilities))
	for i, ability := range abilities {
		effs[i] = sim.FulfillEffectTemplate(ability, teamIndex, cardIndex)
	}

	startAtActiveTime := declarationTime - gs.FullPauseOffset
	chainItem := &state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardID:              cardIndex,
		Type:                state.ChainItemSpell,
		Effects:             effs,
	}
	gs.Chain = append(gs.Chain, *chainItem)

	*logList = append(*logList, logs.AddChainLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       declarationTime,
		Type:                logs.AddChain,
		ChainType:           chainItem.Type,
		StartAtActiveTime:   chainItem.StartAtActiveTime,
		PopAtActiveTime:     chainItem.PopAtActiveTime,
		ExecuteAtActiveTime: chainItem.ExecuteAtActiveTime,
		CardID:              chainItem.CardID,
		Effects:             chainItem.Effects,
	})
	*UpdateID += 1
}
