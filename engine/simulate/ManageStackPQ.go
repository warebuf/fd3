package simulate

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine/defs/effDef/effects"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/state"
)

func ManageStackPQ(gs *state.GameState, pq *stackPQ.PriorityQueue, eventLogs *[]logs.Log) {

	for {

		// get the top stack
		topStack := pq.Peek()
		if topStack == nil {
			return
		}

		// if topStack is not due
		if topStack.ResolveTime > time.Now().UnixMilli() {
			return
		}

		// handle the top effect
		topStack.Effects = SimulateStack(gs, topStack.Effects, eventLogs)
		if len(topStack.Effects) == 0 {
			heap.Pop(pq) // remove 0 index
		} else {
			topStack.DeclareTime = time.Now().UnixMilli()
			topStack.ResolveTime = topStack.DeclareTime + 2000         // 2 seconds later
			pq.Update(topStack, time.Now().UnixMilli()+2000, 0, false) // do not update effect index
		}

	}

}

func SimulateStack(gs *state.GameState, stack []state.Effect, eventLogs *[]logs.Log) []state.Effect {

	topEffect := stack[len(stack)-1]

	switch topEffect.Type {
	case state.EffectDraw:
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

	return stack[:len(stack)-1]

}
