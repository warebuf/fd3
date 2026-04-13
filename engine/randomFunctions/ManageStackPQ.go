package randomFunctions

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/defs/effects"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/state"
)

// returns the number of new logs created
func ManageStackPQ(gs *state.GameState, pq *stackPQ.PriorityQueue, eventLogs *[]logs.Log, NextUpdateID int) int {

	newUpdateIDCount := 0

	for {
		fmt.Println("stackPQ length", pq.Len())

		// get the earliest stack
		earliestStack := pq.Peek()
		if earliestStack == nil {
			return newUpdateIDCount
		}

		// if earliestStack is not due
		if earliestStack.ResolveTime > time.Now().UnixMilli() {
			return newUpdateIDCount
		}

		// handle the top effect
		oldLogCount := len(*eventLogs)
		earliestStack.Effects = SimulateStack(gs, earliestStack.Effects, eventLogs, earliestStack.ResolveTime, NextUpdateID+newUpdateIDCount)
		newLogCount := len(*eventLogs)
		newUpdateIDCount += newLogCount - oldLogCount
		fmt.Println("stack length", len(earliestStack.Effects))
		if len(earliestStack.Effects) == 0 {
			heap.Pop(pq) // remove 0 index
		} else {
			pq.Update(earliestStack, time.Now().UnixMilli()+2000-200, 0, false) // do not update effect index
		}

	}

}

func SimulateStack(gs *state.GameState, stack []state.Effect, eventLogs *[]logs.Log, t int64, NextUpdateID int) []state.Effect {

	topEffect := stack[len(stack)-1]

	switch topEffect.Type {
	case state.EffectDraw:
		fmt.Println("draw eff", topEffect)
		effects.ApplyEffectDraw(gs, topEffect)

		*eventLogs = append(*eventLogs, logs.Log{
			UpdateID:  NextUpdateID,
			Timestamp: t + 200,
			Type:      logs.DrawExec,
			A:         topEffect.A, // A: teamIndex
			B:         1,           // B: num of cards to draw
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
	}

	return stack[:len(stack)-1] // remove the last element

}
