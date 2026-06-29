package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) handleStackResolution() {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	gs := match.GameState

	if len(gs.Stack) == 0 {
		return
	}

	// pop the top of the stack
	stackTop := gs.Stack[len(gs.Stack)-1]
	gs.Stack = gs.Stack[:len(gs.Stack)-1]

	wallTime := stackTop.ExecuteAtActiveTime + gs.FullPauseOffset
	oldLogsSize := len(match.EventLog)

	// apply effects
	for _, eff := range stackTop.Effects {
		sim.ApplyEffect(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, eff, stackTop.CardID, wallTime)
	}

	card := &gs.Cards[stackTop.CardID]
	if card.Zone == state.ZoneStack {
		if card.Type == state.CardTypeUnit {
			card.Zone = state.ZoneField
			card.Status = state.CardStatusUnitActive
			gs.Field[card.Team] = append(gs.Field[card.Team], stackTop.CardID)
			match.EventLog = append(match.EventLog, logs.SummonResolutionLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.SummonResolution,
				TeamIndex:     card.Team,
				CardIndex:     stackTop.CardID,
			})
			match.NextUpdateID++
			sim.ScheduleRecurringAbilities(gs, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, stackTop.CardID, match.GameState.SimPauseStartedAt-gs.SimPauseOffset, wallTime)
		} else if card.Type == state.CardTypeContinuous {
			card.Zone = state.ZoneContinuous
			card.Status = state.CardStatusContinuousActive
			gs.Continuous[card.Team] = append(gs.Continuous[card.Team], stackTop.CardID)
			match.EventLog = append(match.EventLog, logs.PlayContinuousLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.PlayContinuous,
				TeamIndex:     card.Team,
				CardID:        stackTop.CardID,
				Charges:       card.Charges,
			})
			match.NextUpdateID++
		} else {
			card.Zone = state.ZoneGraveyard
			gs.Grave[card.Team] = append(gs.Grave[card.Team], stackTop.CardID)
		}
	} else if stackTop.Type == state.StackItemContinuousActivation {
		card.Charges--
		match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
			UpdateID:         match.NextUpdateID,
			WallTimestamp:    wallTime,
			Type:             logs.RemoveCharge,
			TeamIndex:        card.Team,
			CardID:           stackTop.CardID,
			ChargesRemaining: card.Charges,
		})
		match.NextUpdateID++
		if card.Charges <= 0 {
			for i, id := range gs.Continuous[card.Team] {
				if id == stackTop.CardID {
					gs.Continuous[card.Team] = append(gs.Continuous[card.Team][:i], gs.Continuous[card.Team][i+1:]...)
					break
				}
			}
			card.Zone = state.ZoneGraveyard
			card.Status = state.CardStatusDefault
			gs.Grave[card.Team] = append(gs.Grave[card.Team], stackTop.CardID)
			match.EventLog = append(match.EventLog, logs.RemoveContinuousLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.RemoveContinuous,
				TeamIndex:     card.Team,
				CardID:        stackTop.CardID,
			})
			match.NextUpdateID++
		}
	}

	match.EventLog = append(match.EventLog, logs.RemoveStackLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:          logs.RemoveStack,
		CardID:        stackTop.CardID,
	})
	match.NextUpdateID++

	if len(gs.Stack) == 0 {
		if gs.Phase == state.PhaseReveal {
			pushRevealEvent(match, wallTime-gs.FullPauseOffset+800)
		} else {
			match.TransitionPauseState(-1, wallTime, -1)
			match.processAutoSummonQueue(wallTime, -1)
		}
	} else { // the new top gets a refreshed reaction window
		newTopIdx := len(gs.Stack) - 1
		freshStart := wallTime - gs.FullPauseOffset
		gs.Stack[newTopIdx].StartAtActiveTime = freshStart
		gs.Stack[newTopIdx].PopAtActiveTime = freshStart + constants.REACTION_WINDOW - constants.LEAD
		gs.Stack[newTopIdx].ExecuteAtActiveTime = freshStart + constants.REACTION_WINDOW

		match.EventLog = append(match.EventLog, logs.RefreshStackLog{
			UpdateID:            match.NextUpdateID,
			WallTimestamp:       wallTime,
			Type:                logs.RefreshStack,
			CardID:              gs.Stack[newTopIdx].CardID,
			StartAtActiveTime:   gs.Stack[newTopIdx].StartAtActiveTime,
			PopAtActiveTime:     gs.Stack[newTopIdx].PopAtActiveTime,
			ExecuteAtActiveTime: gs.Stack[newTopIdx].ExecuteAtActiveTime,
		})
		match.NextUpdateID++

		scheduleStackResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
