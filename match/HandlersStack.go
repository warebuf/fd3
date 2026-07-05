package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) handleStackResolution() {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	gs := match.GameState

	if len(gs.ReactionStack) == 0 {
		return
	}

	// pop the top of the reaction stack
	stackTop := gs.ReactionStack[len(gs.ReactionStack)-1]
	gs.ReactionStack = gs.ReactionStack[:len(gs.ReactionStack)-1]

	wallTime := stackTop.ExecuteAtActiveTime + gs.FullPauseOffset
	oldLogsSize := len(match.EventLog)

	// apply effects
	for _, eff := range stackTop.Effects {
		sim.ApplyEffect(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, eff, stackTop.CardIndex, stackTop.ActiveIndex, wallTime)
	}

	card := &gs.Cards[stackTop.CardIndex]
	if card.Zone == state.ZoneStack {
		if card.Type == state.CardTypeUnit {
			card.Zone = state.ZoneField
			card.Status = state.CardStatusUnitActive
			gs.Field[card.Team] = append(gs.Field[card.Team], stackTop.CardIndex)
			match.EventLog = append(match.EventLog, logs.SummonResolutionLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.SummonResolution,
				TeamIndex:     card.Team,
				CardIndex:     stackTop.CardIndex,
			})
			match.NextUpdateID++
			sim.ScheduleAllOnTicks(gs, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, stackTop.CardIndex, match.GameState.SimPauseStartedAt-gs.SimPauseOffset, wallTime)
		} else if card.Type == state.CardTypeContinuous {
			card.Zone = state.ZoneContinuous
			card.Status = state.CardStatusContinuousActive
			gs.Continuous[card.Team] = append(gs.Continuous[card.Team], stackTop.CardIndex)
			match.EventLog = append(match.EventLog, logs.PlayContinuousLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.PlayContinuous,
				TeamIndex:     card.Team,
				CardIndex:     stackTop.CardIndex,
				Charges:       card.Actives[0].Charges,
			})
			match.NextUpdateID++
		} else {
			card.Zone = state.ZoneGraveyard
			gs.Grave[card.Team] = append(gs.Grave[card.Team], stackTop.CardIndex)
		}
	} else if stackTop.Type == state.ReactionContinuousAbility {
		card.Actives[stackTop.ActiveIndex].Pending--
		card.Actives[stackTop.ActiveIndex].Charges--
		match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
			UpdateID:         match.NextUpdateID,
			WallTimestamp:    wallTime,
			Type:             logs.RemoveCharge,
			TeamIndex:        card.Team,
			CardIndex:        stackTop.CardIndex,
			ChargesRemaining: card.Actives[stackTop.ActiveIndex].Charges,
		})
		match.NextUpdateID++
		if card.Actives[stackTop.ActiveIndex].Charges <= 0 {
			for i, id := range gs.Continuous[card.Team] {
				if id == stackTop.CardIndex {
					gs.Continuous[card.Team] = append(gs.Continuous[card.Team][:i], gs.Continuous[card.Team][i+1:]...)
					break
				}
			}
			card.Zone = state.ZoneGraveyard
			card.Status = state.CardStatusDefault
			gs.Grave[card.Team] = append(gs.Grave[card.Team], stackTop.CardIndex)
			match.EventLog = append(match.EventLog, logs.RemoveContinuousLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.RemoveContinuous,
				TeamIndex:     card.Team,
				CardIndex:     stackTop.CardIndex,
			})
			match.NextUpdateID++
		}
	} else if stackTop.Type == state.ReactionUnitAbility {
		wasAtMax := card.Actives[stackTop.ActiveIndex].MaxCharges > 0 && card.Actives[stackTop.ActiveIndex].Charges == card.Actives[stackTop.ActiveIndex].MaxCharges
		card.Actives[stackTop.ActiveIndex].Pending--
		card.Actives[stackTop.ActiveIndex].Charges--
		match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
			UpdateID:         match.NextUpdateID,
			WallTimestamp:    wallTime,
			Type:             logs.RemoveCharge,
			TeamIndex:        card.Team,
			CardIndex:        stackTop.CardIndex,
			ChargesRemaining: card.Actives[stackTop.ActiveIndex].Charges,
		})
		match.NextUpdateID++

		// if the card was at max charges, the gain-charge OnTick ability was paused — restart it now
		if wasAtMax {
			startAtSimTime := wallTime - gs.SimPauseOffset
			def := defs.CardDefs[card.DefNo]
			for tickIndex, r := range def.OnTick {
				for _, eff := range r.Effects {
					if eff.Type == state.EffectGainCharge && eff.A == stackTop.ActiveIndex { // NOTE: not sure about the second part
						sim.ScheduleOnTickAbility(gs, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, stackTop.CardIndex, tickIndex, startAtSimTime, wallTime)
						break
					}
				}
			}
		}
	}

	match.EventLog = append(match.EventLog, logs.RemoveStackLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:          logs.RemoveStack,
		CardIndex:     stackTop.CardIndex,
	})
	match.NextUpdateID++

	if len(gs.ReactionStack) == 0 {
		if gs.Phase == state.PhaseReveal {
			pushRevealEvent(match, wallTime-gs.FullPauseOffset+800)
		} else {
			match.TransitionPauseState(-1, wallTime, -1)
			match.processAutoSummonQueue(wallTime, -1)
		}
	} else { // the new top gets a refreshed reaction window
		newTopIndex := len(gs.ReactionStack) - 1
		freshStart := wallTime - gs.FullPauseOffset
		gs.ReactionStack[newTopIndex].StartAtActiveTime = freshStart
		gs.ReactionStack[newTopIndex].PopAtActiveTime = freshStart + constants.REACTION_WINDOW - constants.LEAD
		gs.ReactionStack[newTopIndex].ExecuteAtActiveTime = freshStart + constants.REACTION_WINDOW

		match.EventLog = append(match.EventLog, logs.RefreshStackLog{
			UpdateID:            match.NextUpdateID,
			WallTimestamp:       wallTime,
			Type:                logs.RefreshStack,
			CardIndex:           gs.ReactionStack[newTopIndex].CardIndex,
			StartAtActiveTime:   gs.ReactionStack[newTopIndex].StartAtActiveTime,
			PopAtActiveTime:     gs.ReactionStack[newTopIndex].PopAtActiveTime,
			ExecuteAtActiveTime: gs.ReactionStack[newTopIndex].ExecuteAtActiveTime,
		})
		match.NextUpdateID++

		scheduleStackResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
