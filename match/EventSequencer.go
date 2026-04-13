package match

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/randomFunctions"
	"github.com/warebuf/fd3/engine/state"
)

// so we have to balance between client inputs and autoevents
// every client input allows us to predict the next n+1 autoevent
// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// this one loop handles pushing inputs into a priority queue
// and taking out Events, Inputs, and AutoSchedulers from it

func (match *Match) EventScheduler() {

	for {
		// clear due PQ events
		match.handleDuePQEvents()

		// check if timer is set
		var timerC <-chan time.Time
		if match.EventPQTimer != nil {
			timerC = match.EventPQTimer.C
		}

		// wait on input or PQ timer
		select {
		case ie := <-match.InputEventChan:
			pushInputEvent(match, ie)

		case <-timerC:
			match.handleDuePQEvents()
		}
	}

}

func (match *Match) handleDuePQEvents() {
	for {

		// return if empty
		if match.EventPQueue.Len() == 0 {
			return
		}

		next := match.EventPQueue[0]

		// no more due events
		if time.Now().Before(next.Event.Timestamp()) {
			break
		}

		heap.Pop(&match.EventPQueue) // pop the event

		// handle the event
		switch next.Type {
		case "Effect":
			fmt.Println("received effect")

			// simulate
			oldLogsSize := len(match.EventLog)
			match.NextUpdateID += randomFunctions.ManageStackPQ(match.GameState, match.PQ, &match.EventLog, match.NextUpdateID)

			// update the published state
			match.GameState.UpdateID = match.NextUpdateID
			match.NextUpdateID += 1
			match.PublishedState = state.DeepCopyGameState(match.GameState)

			// publish the new logs
			match.BroadcastLogs(oldLogsSize)
		case "Input":
			ie := next.Event.(*InputEvent)
			fmt.Println("received Input", ie)

			// BIG NOTE; IM DOING THIS WRONG
			// INPUTS SHOULD BE APPLIED 200 MS ONCE RECEIVED
			// THEREFORE THE SERVER APPLIES THEM THE SAME TIME THE CLIENT APPLIES THEM
			// CORRECT THIS

			// TODO: notably there is little to no validation rn

			// if phase is not 2 or greater, do not accept inputs (maybe you can do this earlier, not sure)
			if match.GameState.Phase != 2 {
				continue
			}

			// activate the card and queue the effects
			oldLogsSize := len(match.EventLog)
			success := randomFunctions.ScheduleActivateCard(match.GameState, match.PQ, match.NextPQIndex, ie.TeamIdx, ie.CardIndex, ie.T.UnixMilli())

			if success {
				// if successful, increment the PQ node count
				match.NextPQIndex += 1

				// update published game state with any applied effects
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				// broadcast any new logs
				match.BroadcastLogs(oldLogsSize)
			}

		case "NewPhase":

			newPhase := next.Event.(*NewPhaseEvent)
			fmt.Println("received newPhase", newPhase)

			if newPhase.Phase == 1 { // entering draw phase

				// saving the old log count for broadcasting the new logs
				oldLogsSize := len(match.EventLog)

				// increment the phase
				match.GameState.Phase = 1
				match.EventLog = append(match.EventLog, logs.Log{
					UpdateID:  match.NextUpdateID,
					Timestamp: time.Now().UnixMilli(),
					Type:      logs.PhaseLog,
					A:         1, // A = phase
				})
				match.NextUpdateID += 1

				// schedule the draw event
				currentTime := time.Now().UnixMilli()
				randomFunctions.ScheduleDrawCard(match.PQ, match.NextPQIndex, 0, currentTime+1000) // 1.0 seconds from now
				match.NextPQIndex += 1
				randomFunctions.ScheduleDrawCard(match.PQ, match.NextPQIndex, 0, currentTime+1500) // 1.5 seconds from now
				match.NextPQIndex += 1
				randomFunctions.ScheduleDrawCard(match.PQ, match.NextPQIndex, 0, currentTime+2000) // 2.0 seconds from now
				match.NextPQIndex += 1
				randomFunctions.ScheduleDrawCard(match.PQ, match.NextPQIndex, 1, currentTime+2500) // 2.5 seconds from now
				match.NextPQIndex += 1
				randomFunctions.ScheduleDrawCard(match.PQ, match.NextPQIndex, 1, currentTime+3000) // 3.0 seconds from now
				match.NextPQIndex += 1
				randomFunctions.ScheduleDrawCard(match.PQ, match.NextPQIndex, 1, currentTime+3500) // 3.5 seconds from now
				match.NextPQIndex += 1

				// schedule phase 2 10s from now
				PushNewPhase(match, time.UnixMilli(currentTime).Add(10*time.Second), 2)

				// update the game state, published state
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				// broadcast new logs
				match.BroadcastLogs(oldLogsSize)

			} else if newPhase.Phase == 2 { // entering game start

				oldLogsSize := len(match.EventLog)

				// increment the phase
				match.GameState.Phase = 2
				match.EventLog = append(match.EventLog, logs.Log{
					UpdateID:  match.NextUpdateID,
					Timestamp: time.Now().UnixMilli(),
					Type:      logs.PhaseLog,
					A:         2, // A = phase
				})
				match.NextUpdateID += 1

				// schedule EventPQ effects
				randomFunctions.InitBoard(match.GameState, match.PQ)

				// update the game state, published state
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				// broadcast new logs
				match.BroadcastLogs(oldLogsSize)

			} else if newPhase.Phase == 3 { // entering game over
				// end game (not sure what this requires tbh)
				// update the game state, published state
				// broadcast new logs
			}
		}

		// set the event for the next stackPQ item
		scheduleStackPQ(match)

	}

	match.setPQTimer() // set timer for next event or set it to nil

}

func PushNewPhase(match *Match, t time.Time, phase int) {
	match.NextSeq++
	newPhase := &NewPhaseEvent{
		T:     t.Add(-200 * time.Millisecond), // schedule 200 ms early
		Phase: phase,
	}
	item := &EventItem{
		Event: newPhase,
		Type:  "NewPhase",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

func pushInputEvent(match *Match, ev *InputEvent) {
	match.NextSeq++
	ev.T = time.Now() // relabel the timestamp when putting it into PQ
	item := &EventItem{
		Event: ev,
		Type:  "Input",
		seq:   int64(match.NextSeq),
	}
	heap.Push(&match.EventPQueue, item)
}

func (match *Match) setPQTimer() {

	// clear timer
	if match.EventPQTimer != nil {
		if !match.EventPQTimer.Stop() {
			select {
			case <-match.EventPQTimer.C: // drain the channel
			default:
			}
		}
	}

	// if no events, return
	first, ok := match.EventPQueue.Peek()
	if !ok {
		match.EventPQTimer = nil
		return
	}

	// set timer
	now := time.Now()
	fireTime := first.Event.Timestamp()
	delay := fireTime.Sub(now)
	if delay < 0 {
		delay = 0
	}
	match.EventPQTimer = time.NewTimer(delay)

}

func scheduleStackPQ(match *Match) {

	// get the timestamp of earliest in StackPQ
	earliestStack := match.PQ.Peek()
	fmt.Println("entered scheduleStackPQ", earliestStack)

	// if no effect events, remove the EffectEvent from the EventPQ
	if earliestStack == nil {
		if match.EarliestEffect != nil && match.EarliestEffect.Index > 0 { // EarliestEffect exists AND is in EventPQ (indicated by index > 0)
			heap.Remove(&match.EventPQueue, match.EarliestEffect.Index)
		}
		match.EarliestEffect = nil
		return
	}

	// if EffectEvent exists and is in the EventPQ, update and reheapify in EventPQ
	earliest := time.UnixMilli(earliestStack.PopTime)
	if match.EarliestEffect != nil && match.EarliestEffect.Index > 0 {
		// update and reheapify
		effEv, _ := match.EarliestEffect.Event.(*EffectEvent)
		effEv.T = earliest
		heap.Fix(&match.EventPQueue, match.EarliestEffect.Index)
		return
	}

	// otherwise add EffectEvent to EventPQ
	ev := &EffectEvent{
		T: earliest,
	}
	item := &EventItem{
		Event: ev,
		Type:  "Effect",
		seq:   int64(match.NextSeq),
	}
	match.NextSeq++
	heap.Push(&match.EventPQueue, item)
	match.EarliestEffect = item

}
