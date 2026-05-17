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
func ScheduleDeclareCard(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, wallTimeAtPop int64) bool {

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

	if gs.PausePhase == 0 {
		return false
	} else if gs.PausePhase == 1 { // sim pause — TODO: handle counter-spells

	} else if gs.PausePhase == 2 { // simulating

		gs.Mana[teamIndex] -= cost
		*logList = append(*logList, logs.SpendManaLog{
			UpdateID:  *UpdateID,
			Timestamp: wallTimeAtPop + constants.LEAD,
			Type:      logs.SpendMana,
			TeamIndex: teamIndex,
			Amount:    cost,
		})
		*UpdateID += 1

		*logList = append(*logList, logs.DeclareCardLog{
			UpdateID:  *UpdateID,
			Timestamp: wallTimeAtPop + constants.LEAD,
			Type:      logs.DeclareCard,
			TeamIndex: teamIndex,
			CardIndex: cardIndex,
		})
		*UpdateID += 1

		if gs.Cards[cardIndex].Type == state.CardTypeUnit {
			gs.Cards[cardIndex].Zone = state.ZoneField
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
			scheduleSummon(gs, logList, UpdateID, teamIndex, cardIndex, wallTimeAtPop)

		} else if gs.Cards[cardIndex].Type == state.CardTypeSpell {
			gs.Cards[cardIndex].Zone = state.ZoneChain
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)
			scheduleChain(gs, logList, UpdateID, teamIndex, cardIndex, wallTimeAtPop+constants.LEAD)
			return true
		}
	}

	return false
}

func ScheduleDrawCard(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, executeAtWallTime int64) {

	// convert the 'execute at wall time' to sim time
	executeAtSimTime := executeAtWallTime - gs.SimPauseOffset

	item := &state.SimulationItem{
		SimTimeAtStart: executeAtSimTime,
		SimTimeAtPop:   executeAtSimTime - constants.LEAD,
		SimTimeAtOccur: executeAtSimTime,
		CardID:         -1,
		Type:           state.SimItemDraw,
		Effects:        []state.Effect{{Type: state.EffectDraw, A: teamIndex, B: 1}},
	}
	heap.Push(pq, item)

	*logList = append(*logList, logs.AddSimItemLog{
		UpdateID:  *UpdateID,
		Timestamp: executeAtWallTime,
		Type:      logs.AddSimItem,

		SeqID:       item.SeqID,
		StartTime:   item.SimTimeAtStart,
		OccurTime:   item.SimTimeAtOccur,
		CardID:      item.CardID,
		SimItemType: int(state.SimItemDraw),
	})
	*UpdateID += 1
}

func scheduleSummon(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	eff := state.Effect{
		Type: state.SummonResolution,
		A:    teamIndex,
		B:    cardIndex,
	}

	simTime := popTime - gs.SimPauseOffset

	simItem := &state.SimulationItem{
		SimTimeAtStart: simTime + constants.LEAD,
		SimTimeAtPop:   simTime + constants.REACTION_WINDOW,
		SimTimeAtOccur: simTime + constants.REACTION_WINDOW + constants.LEAD,

		CardID:  cardIndex,
		Type:    state.SimItemSummon,
		Effects: []state.Effect{eff},
	}
	heap.Push(&gs.SimulationPQ, simItem)

	*logList = append(*logList, logs.AddSimItemLog{
		UpdateID:  *UpdateID,
		Timestamp: popTime + constants.LEAD,
		Type:      logs.AddSimItem,

		SeqID:       simItem.SeqID,
		StartTime:   simItem.SimTimeAtStart,
		OccurTime:   simItem.SimTimeAtOccur,
		CardID:      simItem.CardID,
		SimItemType: int(state.SimItemSummon),
	})
	*UpdateID += 1
}

func scheduleChain(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, declarationTime int64) {
	card := gs.Cards[cardIndex]

	abilities := defs.CardDefs[card.DefNo].OnPlay
	for i := 0; i < len(abilities); i++ {

		newEff := sim.FulfillEffectTemplate(abilities[i], teamIndex, cardIndex)
		chainTime := declarationTime - gs.FullPauseOffset

		chainItem := &state.ChainItem{
			StartTime: chainTime,
			PopTime:   chainTime + constants.REACTION_WINDOW - constants.LEAD,
			OccurTime: chainTime + constants.REACTION_WINDOW,

			CardID:  cardIndex,
			Type:    state.ChainItemSpell,
			Effects: []state.Effect{newEff},
		}
		gs.Chain = append(gs.Chain, *chainItem)

		*logList = append(*logList, logs.AddChainLog{
			UpdateID:  *UpdateID,
			Timestamp: declarationTime,
			Type:      logs.AddChain,
			StartTime: chainItem.StartTime,
			OccurTime: chainItem.OccurTime,
			CardID:    chainItem.CardID,
		})
		*UpdateID += 1
	}
}
