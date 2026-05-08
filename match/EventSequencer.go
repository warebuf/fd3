package match

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/randomFunctions"
	"github.com/warebuf/fd3/engine/state"
)

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

			// simulate
			oldLogsSize := len(match.EventLog)
			randomFunctions.ManageStackPQ(match.GameState, &match.EventLog, &match.NextUpdateID)

			// update the published state
			match.GameState.Timestamp = next.Event.Timestamp().Add(constants.LEAD * time.Millisecond)
			match.GameState.UpdateID = match.NextUpdateID
			match.NextUpdateID += 1
			match.PublishedState = state.DeepCopyGameState(match.GameState)

			// publish the new logs
			match.BroadcastLogs(oldLogsSize)

		case "Input":
			ie := next.Event.(*InputEvent)

			switch match.GameState.Phase {

			case 2: // board setup
				switch ie.Type {

				case "stageCard":
					cardIndex := ie.CardIndex
					teamIdx := ie.TeamIdx

					if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
						continue
					}
					card := match.GameState.Cards[cardIndex]
					if card.Zone != state.ZoneHand || card.Team != teamIdx {
						continue
					}

					handIndex := -1
					for i, ci := range match.GameState.Hand[teamIdx] {
						if ci == cardIndex {
							handIndex = i
							break
						}
					}
					if handIndex < 0 {
						continue
					}

					// move from hand to staged
					match.GameState.Hand[teamIdx] = append(match.GameState.Hand[teamIdx][:handIndex], match.GameState.Hand[teamIdx][handIndex+1:]...)
					match.GameState.StagedCards[teamIdx] = append(match.GameState.StagedCards[teamIdx], cardIndex)
					match.GameState.Cards[cardIndex].Zone = state.ZoneSetup

					// publish updated state
					match.GameState.UpdateID = match.NextUpdateID
					match.NextUpdateID++
					match.PublishedState = state.DeepCopyGameState(match.GameState)
					match.BroadcastPublishedState()

				case "unstageCard":
					cardIndex := ie.CardIndex
					teamIdx := ie.TeamIdx

					if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
						continue
					}
					card := match.GameState.Cards[cardIndex]
					if card.Zone != state.ZoneSetup || card.Team != teamIdx {
						continue
					}

					stageIndex := -1
					for i, ci := range match.GameState.StagedCards[teamIdx] {
						if ci == cardIndex {
							stageIndex = i
							break
						}
					}
					if stageIndex < 0 {
						continue
					}

					// move from staged back to hand
					match.GameState.StagedCards[teamIdx] = append(match.GameState.StagedCards[teamIdx][:stageIndex], match.GameState.StagedCards[teamIdx][stageIndex+1:]...)
					match.GameState.Hand[teamIdx] = append(match.GameState.Hand[teamIdx], cardIndex)
					match.GameState.Cards[cardIndex].Zone = state.ZoneHand

					// publish updated state
					match.GameState.UpdateID = match.NextUpdateID
					match.NextUpdateID++
					match.PublishedState = state.DeepCopyGameState(match.GameState)
					match.BroadcastPublishedState()

				case "ready":
					teamIdx := ie.TeamIdx
					if teamIdx < 0 || teamIdx >= len(match.GameState.Ready) {
						continue
					}
					match.GameState.Ready[teamIdx] = true

					allReady := true
					for _, r := range match.GameState.Ready {
						if !r {
							allReady = false
							break
						}
					}
					if allReady {
						PushNewPhase(match, time.Now().UnixMilli(), 3)
					}
				}

			case 4: // play
				if ie.Type != "hand" {
					continue
				}

				oldLogsSize := len(match.EventLog)
				randomFunctions.ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, ie.TeamIdx, ie.CardIndex, ie.T.UnixMilli())

				if len(match.EventLog) > oldLogsSize {
					match.NextUpdateID += 1

					match.GameState.Timestamp = next.Event.Timestamp().Add(constants.LEAD * time.Millisecond)
					match.GameState.UpdateID = match.NextUpdateID
					match.NextUpdateID += 1
					match.PublishedState = state.DeepCopyGameState(match.GameState)

					match.BroadcastLogs(oldLogsSize)
				}

			default:
				continue
			}

		case "NewPhase":

			newPhase := next.Event.(*NewPhaseEvent)

			// skip stale phase events (e.g. 90s board setup timer after early ready-up)
			if match.GameState.Phase >= newPhase.Phase {
				continue
			}

			switch newPhase.Phase {

			case 1: // entering draw phase

				oldLogsSize := len(match.EventLog)

				phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
				match.GameState.Phase = 1
				match.EventLog = append(match.EventLog, logs.PhaseLog{
					UpdateID:  match.NextUpdateID,
					Timestamp: phaseChangeTimestamp,
					Type:      logs.Phase,
					Phase:     1,
				})
				match.NextUpdateID += 1

				// schedule draws: 5 cards per team, staggered 250ms apart
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+500)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+750)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+1000)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+1250)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+1500)

				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2000)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2250)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2500)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2750)
				randomFunctions.ScheduleDrawCard(&match.GameState.ReactionPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+3000)

				// enter board setup 5 seconds after draw starts
				PushNewPhase(match, phaseChangeTimestamp+5000, 2)

				match.GameState.Timestamp = time.Unix(0, phaseChangeTimestamp*int64(time.Millisecond))
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				match.BroadcastLogs(oldLogsSize)

			case 2: // entering board setup

				oldLogsSize := len(match.EventLog)

				phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
				match.GameState.Phase = 2
				match.EventLog = append(match.EventLog, logs.PhaseLog{
					UpdateID:  match.NextUpdateID,
					Timestamp: phaseChangeTimestamp,
					Type:      logs.Phase,
					Phase:     2,
				})
				match.NextUpdateID += 1

				// initialize staged cards and ready flags
				numTeams := len(match.GameState.Hand)
				match.GameState.StagedCards = make([][]int, numTeams)
				for i := range match.GameState.StagedCards {
					match.GameState.StagedCards[i] = []int{}
				}
				match.GameState.Ready = make([]bool, numTeams)

				// schedule reveal in 90 seconds; can be triggered early by both players readying up
				PushNewPhase(match, phaseChangeTimestamp+90000, 3)

				match.GameState.Timestamp = time.Unix(0, phaseChangeTimestamp*int64(time.Millisecond))
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				match.BroadcastLogs(oldLogsSize)

			case 3: // entering reveal

				oldLogsSize := len(match.EventLog)

				phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
				match.GameState.Phase = 3
				match.EventLog = append(match.EventLog, logs.PhaseLog{
					UpdateID:  match.NextUpdateID,
					Timestamp: phaseChangeTimestamp,
					Type:      logs.Phase,
					Phase:     3,
				})
				match.NextUpdateID += 1

				// move all staged cards back to hand, then declare them simultaneously
				revealTime := phaseChangeTimestamp + 500
				for teamIdx := range match.GameState.StagedCards {
					for _, cardIndex := range match.GameState.StagedCards[teamIdx] {
						match.GameState.Cards[cardIndex].Zone = state.ZoneHand
						match.GameState.Hand[teamIdx] = append(match.GameState.Hand[teamIdx], cardIndex)
					}
				}
				for teamIdx := range match.GameState.StagedCards {
					for _, cardIndex := range match.GameState.StagedCards[teamIdx] {
						randomFunctions.ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, teamIdx, cardIndex, revealTime)
					}
					match.GameState.StagedCards[teamIdx] = nil
				}
				match.GameState.Ready = make([]bool, len(match.GameState.Ready))

				// schedule play phase after reveal effects resolve
				PushNewPhase(match, revealTime+int64(constants.REACTION_WINDOW)+2000, 4)

				match.GameState.Timestamp = time.Unix(0, phaseChangeTimestamp*int64(time.Millisecond))
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				match.BroadcastLogs(oldLogsSize)

			case 4: // entering play

				oldLogsSize := len(match.EventLog)

				phaseChangeTimestamp := newPhase.T.UnixMilli() + constants.LEAD
				match.GameState.Phase = 4
				match.EventLog = append(match.EventLog, logs.PhaseLog{
					UpdateID:  match.NextUpdateID,
					Timestamp: phaseChangeTimestamp,
					Type:      logs.Phase,
					Phase:     4,
				})
				match.NextUpdateID += 1

				randomFunctions.InitBoard(match.GameState)

				match.GameState.Timestamp = newPhase.T.Add(constants.LEAD * time.Millisecond)
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				match.BroadcastLogs(oldLogsSize)

			case 5: // entering game over
				// TODO: end game logic
			}
		}

		// set the event for the next stackPQ item
		scheduleStackPQ(match)

	}

	match.setPQTimer() // set timer for next event or set it to nil

}

func PushNewPhase(match *Match, t int64, phase int) {
	match.NextSeq++
	newPhase := &NewPhaseEvent{
		T:     time.UnixMilli(t - constants.LEAD), // schedule 200 ms early
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
	earliestStack := match.GameState.ReactionPQ.Peek()

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
