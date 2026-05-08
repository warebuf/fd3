package match

import (
	"container/heap"
	"sort"
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

				case "hand":
					cardIndex := ie.CardIndex
					teamIdx := ie.TeamIdx

					if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
						continue
					}
					card := match.GameState.Cards[cardIndex]
					if card.Zone != state.ZoneHand || card.Team != teamIdx || card.AutoSummon != 0 {
						continue
					}

					nextAutoSummon := 1
					for _, c := range match.GameState.Cards {
						if c.Team == teamIdx && c.AutoSummon >= nextAutoSummon {
							nextAutoSummon = c.AutoSummon + 1
						}
					}
					match.GameState.Cards[cardIndex].AutoSummon = nextAutoSummon

					oldLogsSize := len(match.EventLog)
					match.EventLog = append(match.EventLog, logs.AutoSummonLog{
						Timestamp:     ie.T.UnixMilli(),
						UpdateID:      match.NextUpdateID,
						Type:          logs.AutoSummonIndex,
						TeamIndex:     teamIdx,
						CardIndex:     cardIndex,
						AutoSummonIdx: nextAutoSummon,
					})
					match.NextUpdateID++
					match.PublishedState = state.DeepCopyGameState(match.GameState)
					match.BroadcastLogs(oldLogsSize)

				case "clearAutoSummon":
					cardIndex := ie.CardIndex
					teamIdx := ie.TeamIdx

					if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
						continue
					}
					card := match.GameState.Cards[cardIndex]
					if card.Zone != state.ZoneHand || card.Team != teamIdx || card.AutoSummon == 0 {
						continue
					}

					cleared := match.GameState.Cards[cardIndex].AutoSummon
					match.GameState.Cards[cardIndex].AutoSummon = 0
					for i := range match.GameState.Cards {
						if match.GameState.Cards[i].Team == teamIdx && match.GameState.Cards[i].AutoSummon > cleared {
							match.GameState.Cards[i].AutoSummon--
						}
					}

					oldLogsSize := len(match.EventLog)
					match.EventLog = append(match.EventLog, logs.AutoSummonLog{
						Timestamp:     ie.T.UnixMilli(),
						UpdateID:      match.NextUpdateID,
						Type:          logs.AutoSummonIndex,
						TeamIndex:     teamIdx,
						CardIndex:     cardIndex,
						AutoSummonIdx: 0,
					})
					match.NextUpdateID++
					match.PublishedState = state.DeepCopyGameState(match.GameState)
					match.BroadcastLogs(oldLogsSize)

				case "ready":
					teamIdx := ie.TeamIdx
					if teamIdx < 0 || teamIdx >= len(match.GameState.Ready) {
						continue
					}
					match.GameState.Ready[teamIdx] = true

					oldLogsSize := len(match.EventLog)
					match.EventLog = append(match.EventLog, logs.ReadyLog{
						Timestamp: ie.T.UnixMilli(),
						UpdateID:  match.NextUpdateID,
						Type:      logs.Ready,
						TeamIndex: teamIdx,
					})
					match.NextUpdateID++
					match.BroadcastLogs(oldLogsSize)

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

				// enter board setup after draws complete
				PushNewPhase(match, phaseChangeTimestamp+3500, 2)

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

				// initialize ready flags
				numTeams := len(match.GameState.Hand)
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

				// collect staged cards per team, sorted by AutoSummon order
				revealTime := phaseChangeTimestamp + 500
				numTeams := len(match.GameState.Hand)
				stagedByTeam := make([][]int, numTeams)
				for i, c := range match.GameState.Cards {
					if c.AutoSummon > 0 {
						stagedByTeam[c.Team] = append(stagedByTeam[c.Team], i)
					}
				}
				for teamIdx := range stagedByTeam {
					sort.Slice(stagedByTeam[teamIdx], func(a, b int) bool {
						return match.GameState.Cards[stagedByTeam[teamIdx][a]].AutoSummon < match.GameState.Cards[stagedByTeam[teamIdx][b]].AutoSummon
					})
				}

				// declare in AutoSummon order, alternating between teams
				maxStaged := 0
				for _, staged := range stagedByTeam {
					if len(staged) > maxStaged {
						maxStaged = len(staged)
					}
				}
				for i := 0; i < maxStaged; i++ {
					for teamIdx, staged := range stagedByTeam {
						if i >= len(staged) {
							continue
						}
						cardIndex := staged[i]
						match.GameState.Cards[cardIndex].AutoSummon = 0
						randomFunctions.ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, teamIdx, cardIndex, revealTime)
					}
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
