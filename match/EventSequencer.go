package match

import (
	"container/heap"
	"fmt"
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

			effEvent := next.Event.(*EffectEvent)

			// simulate
			oldLogsSize := len(match.EventLog)
			randomFunctions.ManageSimulationPQ(match.GameState, &match.EventLog, &match.NextUpdateID)

			// update the published state
			match.GameState.WallTimestamp = effEvent.T.UnixMilli() + constants.LEAD
			match.GameState.UpdateID = match.NextUpdateID
			match.NextUpdateID += 1
			match.PublishedState = state.DeepCopyGameState(match.GameState)

			// publish the new logs
			match.BroadcastLogs(oldLogsSize)

		case "Input":
			ie := next.Event.(*InputEvent)

			switch ie.Type {
			case "fullPause":
				// set GameState
			case "hand":

				// do not accept if we are not in phase 2 (board set up) or 4 (play)
				if match.GameState.Phase != 2 || match.GameState.Phase != 4 {
					continue
				}

				cardIndex := ie.CardIndex
				teamIdx := ie.TeamIdx
				if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
					continue
				}

				card := match.GameState.Cards[cardIndex]
				if card.Zone != state.ZoneHand || card.Team != teamIdx {
					continue
				}

				// if we are in phase 2 (board set up)
				if match.GameState.Phase == 2 {

					if card.AutoSummon != 0 {
						continue
					}

					// stage the card
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
				}

				// if we are in phase 4 (play)
				if match.GameState.Phase == 4 {
					oldLogsSize := len(match.EventLog)
					randomFunctions.ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, ie.TeamIdx, ie.CardIndex, ie.T.UnixMilli())
					if len(match.EventLog) > oldLogsSize {
						match.GameState.WallTimestamp = ie.T.UnixMilli() + constants.LEAD
						match.GameState.UpdateID = match.NextUpdateID
						match.NextUpdateID += 1
						match.PublishedState = state.DeepCopyGameState(match.GameState)
						match.BroadcastLogs(oldLogsSize)
					}
				}

			case "undoHand":
				// only accept if we are in phase 2 (board set up)
				if match.GameState.Phase != 2 {
					continue
				}

				cardIndex := ie.CardIndex
				teamIdx := ie.TeamIdx

				// do not accept a card that does not exist
				if cardIndex < 0 || cardIndex >= len(match.GameState.Cards) {
					continue
				}

				// card must be in hand, must match the user's team, must be staged
				card := match.GameState.Cards[cardIndex]
				if card.Zone != state.ZoneHand || card.Team != teamIdx || card.AutoSummon == 0 {
					continue
				}

				// unstage the card, shift all subsequent staged cards
				cleared := match.GameState.Cards[cardIndex].AutoSummon
				match.GameState.Cards[cardIndex].AutoSummon = 0
				for i := range match.GameState.Cards {
					if match.GameState.Cards[i].Team == teamIdx && match.GameState.Cards[i].AutoSummon > cleared {
						match.GameState.Cards[i].AutoSummon--
					}
				}

				// create the log, update game state, broadcast the log
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
				// only accept if we are in stage 2 (board set up)
				if match.GameState.Phase != 2 {
					continue
				}

				// do not accept if not acceptable team index
				teamIdx := ie.TeamIdx
				if teamIdx < 0 || teamIdx >= len(match.GameState.Ready) {
					continue
				}

				// if the player is already ready, do not accept
				if match.GameState.Ready[teamIdx] {
					continue
				}

				// update ready arr
				match.GameState.Ready[teamIdx] = true

				// create the log and broadcast
				oldLogsSize := len(match.EventLog)
				match.EventLog = append(match.EventLog, logs.ReadyLog{
					Timestamp: ie.T.UnixMilli(),
					UpdateID:  match.NextUpdateID,
					Type:      logs.Ready,
					TeamIndex: teamIdx,
				})
				match.NextUpdateID++
				match.BroadcastLogs(oldLogsSize)

				// if all ready, increment the phase to board set up
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

				// TODO: enter animation pause phase ? maybe not ? it can be something specific to phase 2

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

				// TODO: also animation pause phase. can also be something specific to phase 3

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

				match.GameState.WallTimestamp = phaseChangeTimestamp
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

				// enter simulating state
				match.TransitionPauseState(2, phaseChangeTimestamp)

				// updating the published state
				match.GameState.WallTimestamp = phaseChangeTimestamp
				match.GameState.UpdateID = match.NextUpdateID
				match.NextUpdateID += 1
				match.PublishedState = state.DeepCopyGameState(match.GameState)

				// broadcasting to clients
				match.BroadcastLogs(oldLogsSize)

			case 5: // entering game over
				// TODO: end game logic

			}

		case "PauseEnd":
			pauseEndEvent := next.Event.(*PauseEndEvent)
			match.TransitionPauseState(-1, pauseEndEvent.T.UnixMilli())
		}

	}

	match.setPQTimer() // set timer for next event or set it to nil

}

// TODO: there are probably more events that need to be handled tbh
func (match *Match) TransitionPauseState(newPauseState int, transitionTime int64) {

	// 0: indefinite pause
	// 1: sim pause
	// 2: simulating

	gameState := match.GameState
	currentPauseState := match.GameState.GetPauseMode()

	// if newPauseState < 0, end the current pause state
	if newPauseState < 0 {
		switch currentPauseState {
		case 2: // already in simulating, do nothing
			return
		case 1: // end sim pause by incrementing SimPauseOffset, schedule simEvents
			gameState.SimPauseOffset += transitionTime - gameState.PauseStack[1]
			gameState.PauseStack[1] = 0
			scheduleSimulationItem(match)
			return
		case 0: // end indefinite pause
			// entering sim pause
			if gameState.PauseStack[1] > 0 {
				// some notes on why how we increment in this case:
				// say simPause at t=3, indefinitePause at t=4, unpause indefinitePause at t=5, unpause simPause at t=6
				// SimPauseOffset +3, FullPauseOffsest at +1
				gameState.FullPauseOffset += transitionTime - gameState.PauseStack[0]
				gameState.PauseStack[0] = 0
				scheduleChainItem(match)
				return
			}
			// entering simulating
			gameState.SimPauseOffset += transitionTime - gameState.PauseStack[0]
			gameState.FullPauseOffset += transitionTime - gameState.PauseStack[0]
			gameState.PauseStack[0] = 0
			scheduleChainItem(match)
			scheduleSimulationItem(match)
			return
		}
	}

	// if newPause >= 0, entering a pause state
	switch newPauseState {
	case 0: // entering indefinite pause
		gameState.PauseStack[0] = transitionTime
		if match.EarliestSimItem != nil {
			match.EventPQueue.Remove(match.EarliestSimItem)
			match.EarliestSimItem = nil
		}
		if match.ChainItem != nil {
			match.EventPQueue.Remove(match.ChainItem)
			match.ChainItem = nil
		}
	case 1: // entering sim pause
		gameState.PauseStack[1] = transitionTime
		match.EventPQueue.Remove(match.EarliestSimItem)
		match.EarliestSimItem = nil

	case 2: // entering simulating. do nothing, should not be using the function this way
		fmt.Println("please do not use the function this way. just unpause by passing -1 into newPauseState")
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
