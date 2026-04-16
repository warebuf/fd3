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

		// get the earliest stack
		earliestStack := pq.Peek()
		if earliestStack == nil {
			return newUpdateIDCount
		}

		// if earliestStack is not due
		if earliestStack.PopTime > time.Now().UnixMilli() {
			return newUpdateIDCount
		}

		// handle the top effect
		oldLogCount := len(*eventLogs)
		earliestStack.Effects = SimulateStack(gs, earliestStack.Effects, eventLogs, earliestStack.PopTime, NextUpdateID+newUpdateIDCount)
		newLogCount := len(*eventLogs)
		newUpdateIDCount += newLogCount - oldLogCount
		if len(earliestStack.Effects) == 0 {
			heap.Pop(pq) // remove 0 index
		} else {
			pq.Update(earliestStack, time.Now().UnixMilli()+2000, 0, false) // do not update effect index
		}

	}

}

func SimulateStack(gs *state.GameState, stack []state.Effect, eventLogs *[]logs.Log, t int64, NextUpdateID int) []state.Effect {

	topEffect := stack[len(stack)-1]

	switch topEffect.Type {
	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, topEffect)

		fmt.Println("scheduling draw for client at", (t+200)%1000000)

		// create log
		*eventLogs = append(*eventLogs, logs.Log{
			UpdateID:  NextUpdateID,
			Timestamp: t + 200, // clients are allowed to receive logs 200ms early, but see them in 200ms
			Type:      logs.DrawExec,
			A:         topEffect.A, // A: teamIndex
			B:         1,           // B: num of cards to draw
		})

	case state.SummonResolution:

		fmt.Println("scheduling summon resolution for client at", (t+200)%1000000)

		success := false
		for i := 0; i < len(gs.Hand); i++ {
			if i == topEffect.A {
				for j := 0; j < len(gs.Hand[i]); j++ {
					if gs.Hand[i][j] == topEffect.B {
						gs.Cards[topEffect.B].Zone = state.ZoneGraveyard         // change card's state
						gs.Hand[i] = append(gs.Hand[i][:j], gs.Hand[i][j+1:]...) // remove from hand arr
						gs.Grave[i] = append(gs.Grave[i], topEffect.B)           // add to graveyard
						success = true
					}
				}
			}
		}

		// create permanent
		if success {
			// errr not sure what to do here
			perm := &state.Permanent{
				PermanentID: gs.LastPermanentID,
				SourceCard:  topEffect.B,
				Type:        1,
			}

			//

			//
		}

		// create log

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
