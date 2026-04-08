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

func ManageStackPQ(gs *state.GameState, pq *stackPQ.PriorityQueue, eventLogs *[]logs.Log) {

	for {
		fmt.Println("stackPQ length", pq.Len())

		// get the earliest stack
		earliestStack := pq.Peek()
		if earliestStack == nil {
			return
		}

		// if earliestStack is not due
		if earliestStack.ResolveTime > time.Now().UnixMilli() {
			return
		}

		// handle the top effect
		earliestStack.Effects = SimulateStack(gs, earliestStack.Effects, eventLogs)
		fmt.Println("stack length", len(earliestStack.Effects))
		if len(earliestStack.Effects) == 0 {
			heap.Pop(pq) // remove 0 index
		} else {
			earliestStack.DeclareTime = time.Now().UnixMilli()
			earliestStack.ResolveTime = earliestStack.DeclareTime + 2000    // 2 seconds later
			pq.Update(earliestStack, time.Now().UnixMilli()+2000, 0, false) // do not update effect index
		}

	}

}

func SimulateStack(gs *state.GameState, stack []state.Effect, eventLogs *[]logs.Log) []state.Effect {

	topEffect := stack[len(stack)-1]

	switch topEffect.Type {
	case state.EffectDraw:
		fmt.Println("draw eff", topEffect)
		effects.ApplyEffectDraw(gs, topEffect)

		*eventLogs = append(*eventLogs, logs.Log{
			Timestamp: time.Now().UnixMilli(),
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
