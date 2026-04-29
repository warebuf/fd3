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
func ManageStackPQ(gs *state.GameState, logList *[]any, UpdateID *int) {

	for {
		// get the earliest stack
		earliestStack := gs.ReactionPQ.Peek()
		if earliestStack == nil {
			return
		}

		// if earliestStack is not due
		if earliestStack.PopTime > time.Now().UnixMilli() {
			return
		}

		// handle the top effect
		SimulateStack(gs, logList, UpdateID)
	}

}

// handling it as a yugioh chain, could be a mtg stack though
func SimulateStack(gs *state.GameState, logList *[]any, UpdateID *int) {

	stackHeader := gs.ReactionPQ.Peek()

	for len(stackHeader.StackEntries) > 0 {
		topEntry := stackHeader.StackEntries[len(stackHeader.StackEntries)-1]
		for i := 0; i < len(topEntry.Effects); i++ {

			currentEffect := topEntry.Effects[i]

			// apply effect
			switch currentEffect.Type {

			case state.EffectDraw:
				effects.ApplyEffectDraw(gs, currentEffect)

				// create log
				*logList = append(*logList, logs.DrawExecEntry{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + 200,
					Type:      logs.DrawExec,
					TeamIndex: currentEffect.A,
					NumCards:  1,
				})
				*UpdateID += 1

			case state.SummonResolution:

				// complete the summon
				gs.Cards[currentEffect.B].Pending = false

				// TODO: handle any effects here?
				// TODO: handle any triggers here?

				// create log
				*logList = append(*logList, logs.SummonUnitEntry{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + 200,
					Type:      logs.SummonUnit,
					TeamIndex: currentEffect.A,
					CardIndex: currentEffect.B,
				})
				*UpdateID += 1

			case state.EffectDestroy:
				fmt.Println("TODO: received EffectDestroy", currentEffect)

			default:
				fmt.Println("this effect isnt handled")
			}

		}

		// remove top entry of the stack
		stackHeader.StackEntries = stackHeader.StackEntries[:len(stackHeader.StackEntries)-1]
	}

	heap.Pop(&gs.ReactionPQ)
	*logList = append(*logList, logs.RemoveStackLog{
		UpdateID:  *UpdateID,
		Timestamp: stackHeader.PopTime,
		Type:      logs.RemoveStack,
		StackID:   stackHeader.StackID,
	})
	*UpdateID += 1

}
