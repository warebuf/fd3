package randomFunctions

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/defs/effects"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// returns the number of new logs created
func ManageStackPQ(gs *state.GameState, eventLogs *[]logs.Log, NextUpdateID int) int {

	newUpdateIDCount := 0

	for {

		// get the earliest stack
		earliestStack := gs.ReactionPQ.Peek()
		if earliestStack == nil {
			return newUpdateIDCount
		}

		// if earliestStack is not due
		if earliestStack.PopTime > time.Now().UnixMilli() {
			return newUpdateIDCount
		}

		// handle the top effect
		oldLogCount := len(*eventLogs)
		earliestStack.Effects = SimulateStack(gs, eventLogs, NextUpdateID+newUpdateIDCount)
		newLogCount := len(*eventLogs)
		newUpdateIDCount += newLogCount - oldLogCount
		if len(earliestStack.Effects) == 0 {
			heap.Pop(pq) // remove 0 index
		} else {
			pq.Update(earliestStack, time.Now().UnixMilli()+2000, 0, false) // do not update effect index
		}

	}

}

func SimulateStack(gs *state.GameState, eventLogs *[]logs.Log, NextUpdateID int) []state.Effect {

	stack := gs.ReactionPQ.Peek().StackEntries
	topEffect := stack[len(stack)-1]

	switch topEffect.Type {
	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, topEffect)

		// create log
		*eventLogs = append(*eventLogs, logs.Log{
			UpdateID:  NextUpdateID,
			Timestamp: t + 200, // clients are allowed to receive logs 200ms early, but see them in 200ms
			Type:      logs.DrawExec,
			A:         topEffect.A, // A: teamIndex
			B:         1,           // B: num of cards to draw
		})

	case state.SummonResolution:

		// complete the summon
		gs.Cards[topEffect.B].Pending = false

		// TODO: handle any effects here?
		// TODO: handle any triggers here?

		// create log
		*eventLogs = append(*eventLogs, logs.Log{
			UpdateID:  NextUpdateID,
			Timestamp: t + 200, // clients are allowed to receive logs 200ms early, but see them in 200ms
			Type:      logs.SummonUnit,
			A:         topEffect.A, // A: teamIndex
			B:         topEffect.B, // B: card index
		})

	case state.EffectAttack:
		// placeholder
	case state.EffectHeal:
		// placeholder
	case state.EffectMana:
		// placeholder
	case state.EffectSilence:
		// placeholder
	case state.EffectGrantAbility:
		// placeholder
	case state.EffectDestroy:
		fmt.Println("TODO: received EffectDestroy", topEffect)

	}

	return stack[:len(stack)-1] // remove the last element

}
