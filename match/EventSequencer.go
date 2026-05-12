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

		next := match.EventPQueue.Peek()

		// no more due events
		if time.Now().Before(next.Event.Timestamp()) {
			break
		}

		heap.Pop(&match.EventPQueue)

		// handle the event
		switch next.Type {

		case "Effect":

			effEv := next.Event.(*EffectEvent)

			// simulate
			oldLogsSize := len(match.EventLog)
			randomFunctions.ManageSimulationPQ(match.GameState, &match.EventLog, &match.NextUpdateID)

			// update the published state
			match.GameState.WallTimestamp = effEv.T.UnixMilli() + constants.LEAD
			match.GameState.UpdateID = match.NextUpdateID
			match.NextUpdateID += 1
			match.PublishedState = state.DeepCopyGameState(match.GameState)

			// publish the new logs
			match.BroadcastLogs(oldLogsSize)

		case "Input":
			ie := next.Event.(*InputEvent)

			switch ie.Type {
			case "FullPause":
				// set GameState
			}

		case "NewPhase":

			newPhase := next.Event.(*NewPhaseEvent)

			// skip stale phase events (e.g. board setup timer after early ready-up)
			if match.GameState.Phase >= newPhase.Phase {
				continue
			}

			switch newPhase.Phase {
			case 1: // entering draw phase
				oldLogsSize := len(match.EventLog)

				// change the phase
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
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+500)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+750)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+1000)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+1250)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 0, phaseChangeTimestamp+1500)

				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2000)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2250)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2500)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+2750)
				randomFunctions.ScheduleDrawCard(&match.GameState.SimulationPQ, &match.EventLog, &match.NextUpdateID, 1, phaseChangeTimestamp+3000)

				// schedule board setup phase after draws complete
				PushNewPhase(match, phaseChangeTimestamp+3500, 2)

				match.GameState.WallTimestamp = phaseChangeTimestamp
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				match.BroadcastLogs(oldLogsSize)

				// schedule draw cards
				scheduleEarliestSimItem(match)

			case 2: // entering board set up phase

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

				match.GameState.WallTimestamp = phaseChangeTimestamp
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
				declareTime := revealTime
				const declareDelay = int64(800) // ms between each auto-summon declaration
				for i := 0; i < maxStaged; i++ {
					for teamIdx, staged := range stagedByTeam {
						if i >= len(staged) {
							continue
						}
						cardIndex := staged[i]
						match.GameState.Cards[cardIndex].AutoSummon = 0
						randomFunctions.ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, teamIdx, cardIndex, declareTime)
						declareTime += declareDelay
					}
				}
				match.GameState.Ready = make([]bool, len(match.GameState.Ready))

				// schedule play phase after reveal effects resolve
				PushNewPhase(match, declareTime+int64(constants.REACTION_WINDOW)+2000, 4)

				match.GameState.WallTimestamp = time.Unix(0, phaseChangeTimestamp*int64(time.Millisecond))
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

				match.GameState.WallTimestamp = phaseChangeTimestamp
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				match.BroadcastLogs(oldLogsSize)

			case 5: // entering game over
				// TODO: end game logic

			}

		case "PauseEnd":

			pauseEndEvent := next.Event.(*PauseEndEvent)
			currentPauseMode := match.GameState.GetPauseMode()
			match.TransitionPauseState(currentPauseMode, true, pauseEndEvent.T.UnixMilli())

		}

	}

	match.setPQTimer() // set timer for next event or set it to nil

}

// call when you want to transition the pause state. isEnd is whether pauseState is the 'ending' state.
func (match *Match) TransitionPauseState(pauseState int, isEnd bool, endTime int64) {

	gs := match.GameState

	// determine startPauseState and endPauseState
	var startPauseState int
	var endPauseState int
	if isEnd {
		startPauseState = pauseState

		// find the highest-priority pause still active after removing this one
		endPauseState = 3 // default: return to simulating
		for i := startPauseState + 1; i < 3; i++ {
			if gs.PauseStack[i] > 0 {
				endPauseState = i
				break
			}
		}
	} else {
		startPauseState = gs.GetPauseMode()
		endPauseState = pauseState
	}

	switch startPauseState {
	case 3: // simulating
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}

	case 2: // sim pause
		if gs.PauseStack[2] != 0 {
			gs.SimPauseOffset += endTime - gs.PauseStack[2]
			gs.PauseStack[2] = 0
		}
		if match.ChainItem != nil {
			match.EventPQueue.Remove(match.ChainItem)
			match.ChainItem = nil
		}

	case 1: // animation pause
		if gs.PauseStack[1] != 0 {
			gs.SimPauseOffset += endTime - gs.PauseStack[1]
			gs.FullPauseOffset += endTime - gs.PauseStack[1]
			gs.PauseStack[1] = 0
		}
		if match.PauseEndItem != nil {
			match.EventPQueue.Remove(match.PauseEndItem)
			match.PauseEndItem = nil
		}

	case 0: // full pause
		if gs.PauseStack[0] != 0 {
			gs.SimPauseOffset += endTime - gs.PauseStack[0]
			gs.FullPauseOffset += endTime - gs.PauseStack[0]
			gs.PauseStack[0] = 0
		}
		if match.PauseEndItem != nil {
			match.EventPQueue.Remove(match.PauseEndItem)
			match.PauseEndItem = nil
		}
	}

	switch endPauseState {
	case 3: // simulating
		scheduleSimulationItem(match)

	case 2: // sim pause
		gs.PauseStack[2] = endTime
		scheduleChainItem(match)

	case 1: // animation pause
		gs.PauseStack[1] = endTime
		ev := &PauseEndEvent{T: time.UnixMilli(endTime)}
		match.PauseEndItem = match.EventPQueue.Enqueue(ev, "PauseEnd")

	case 0: // full pause
		gs.PauseStack[0] = endTime
		ev := &PauseEndEvent{T: time.UnixMilli(endTime)}
		match.PauseEndItem = match.EventPQueue.Enqueue(ev, "PauseEnd")
	}

}

func PushNewPhase(match *Match, t int64, phase int) {
	newPhase := &NewPhaseEvent{
		T:     time.UnixMilli(t - constants.LEAD), // schedule 200 ms early
		Phase: phase,
	}
	match.EventPQueue.Enqueue(newPhase, "NewPhase")
}

func pushInputEvent(match *Match, ev *InputEvent) {
	ev.T = time.Now() // relabel the timestamp when putting it into PQ
	match.EventPQueue.Enqueue(ev, "Input")
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
	first := match.EventPQueue.Peek()
	if first == nil {
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

func scheduleEarliestSimItem(match *Match) {

	// get the timestamp of earliest in StackPQ
	earliestSimItem := match.GameState.SimulationPQ.Peek()

	// if no effect events, remove the EffectEvent from the EventPQ
	if earliestSimItem == nil {
		if match.EarliestSimItem != nil && match.EarliestSimItem.Index > 0 { // EarliestEffect exists AND is in EventPQ (indicated by index > 0)
			match.EventPQueue.Remove(match.EarliestSimItem)
		}
		match.EarliestSimItem = nil
		return
	}

	// popTime is denominated in sim time, so we convert sim time to wall time by adding accumulated pause offset
	earliest := time.UnixMilli(earliestSimItem.PopTime + match.GameState.SimPauseOffset)

	// if EffectEvent exists and is in the EventPQ, update and reheapify in EventPQ
	if match.EarliestSimItem != nil && match.EarliestSimItem.Index > 0 {
		// update and reheapify
		effEv, _ := match.EarliestSimItem.Event.(*EffectEvent)
		effEv.T = earliest
		match.EventPQueue.Fix(match.EarliestSimItem)
		return
	}

	// otherwise add EffectEvent to EventPQ
	ev := &EffectEvent{T: earliest}
	match.EarliestSimItem = match.EventPQueue.Enqueue(ev, "Effect")

}
