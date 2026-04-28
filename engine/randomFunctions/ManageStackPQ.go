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
func ManageStackPQ(gs *state.GameState, eventLogs *[]logs.Log, UpdateID *int) {

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
		SimulateStack(gs, eventLogs, UpdateID)
	}

}

// handling it as a yugioh chain, could be a mtg stack though
func SimulateStack(gs *state.GameState, eventLogs *[]logs.Log, UpdateID *int) {

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
				*eventLogs = append(*eventLogs, logs.Log{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + 200, // clients are allowed to receive logs 200ms early, but see them in 200ms
					Type:      logs.DrawExec,
					A:         currentEffect.A, // A: teamIndex
					B:         1,               // B: num of cards to draw
				})
				*UpdateID += 1

			case state.SummonResolution:

				// complete the summon
				gs.Cards[currentEffect.B].Pending = false

				// TODO: handle any effects here?
				// TODO: handle any triggers here?

				// create log
				*eventLogs = append(*eventLogs, logs.Log{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + 200, // clients are allowed to receive logs 200ms early, but see them in 200ms
					Type:      logs.SummonUnit,
					A:         currentEffect.A, // A: teamIndex
					B:         currentEffect.B, // B: card index
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

	heap.Pop(&gs.ReactionPQ) // remove the stack header

}
