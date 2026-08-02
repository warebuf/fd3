package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/sim/conditions/trigger"
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) handleReactionStackResolution() {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	gs := match.GameState

	if len(gs.ReactionStack) == 0 {
		return
	}

	top := gs.ReactionStack[len(gs.ReactionStack)-1]
	wallTime := top.ExecuteAtActiveTime + gs.FullPauseOffset

	// the window's timer just fired: reveal a contested reaction if either team submitted one,
	// instead of resolving the current top of the stack. the priority holder's submission always
	// wins if they made one; otherwise the other team's submission wins by default.
	winnerTeam := -1
	if gs.PendingReactions[gs.Priority].Type != 0 {
		winnerTeam = gs.Priority
	} else if gs.PendingReactions[1-gs.Priority].Type != 0 {
		winnerTeam = 1 - gs.Priority
	}

	if winnerTeam >= 0 {
		p := gs.PendingReactions[winnerTeam]
		gs.PendingReactions = make([]state.PendingReaction, len(gs.PendingReactions)) // clear both slots; the loser (if any) simply fails

		oldLogsSize := len(match.EventLog)
		ok := false
		switch p.Type {
		case state.PendingReactionCard:
			ok = reaction.PushCardOnStack(gs, &match.EventLog, &match.NextUpdateID, winnerTeam, p.CardIndex, wallTime-constants.LEAD)
		case state.PendingReactionAbility:
			ok = reaction.PushAbilityOnStack(gs, &match.EventLog, &match.NextUpdateID, p.CardIndex, p.ActiveIndex, wallTime)
		case state.PendingReactionStructureDraw:
			reaction.PushStructureDrawOnStack(gs, &match.EventLog, &match.NextUpdateID, winnerTeam, wallTime)
			ok = true
		}

		// state is frozen during an open window, so this should never fail; fall through to
		// resolving the top of the stack as if nobody had submitted if it somehow does.
		if ok {
			scheduleReactionStackResolution(match)
			match.publish(wallTime, oldLogsSize)
			return
		}
	}

	// pop the top of the reaction stack
	reactionTop := gs.ReactionStack[len(gs.ReactionStack)-1]
	gs.ReactionStack = gs.ReactionStack[:len(gs.ReactionStack)-1]

	oldLogsSize := len(match.EventLog)

	// apply effects
	for _, eff := range reactionTop.Effects {
		sim.ApplyEffect(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, eff, reactionTop.Effects, reactionTop.CardIndex, reactionTop.ActiveIndex, wallTime)
	}

	if reactionTop.Type != state.ReactionStructureAbility {
		card := &gs.Cards[reactionTop.CardIndex]
		if card.Zone == state.ZoneStack {
			if card.Type == state.CardTypeUnit {
				card.Zone = state.ZoneField
				gs.Field[card.Team] = append(gs.Field[card.Team], reactionTop.CardIndex)
				match.EventLog = append(match.EventLog, logs.SummonResolutionLog{
					UpdateID:      match.NextUpdateID,
					WallTimestamp: wallTime,
					Type:          logs.SummonResolution,
					TeamIndex:     card.Team,
					CardIndex:     reactionTop.CardIndex,
				})
				match.NextUpdateID++
				schedule.ScheduleAllOnTicks(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, reactionTop.CardIndex, match.GameState.SimPauseStartedAt-gs.SimPauseOffset, wallTime)
			} else if card.Type == state.CardTypeContinuous {
				card.Zone = state.ZoneContinuous
				gs.Continuous[card.Team] = append(gs.Continuous[card.Team], reactionTop.CardIndex)
				match.EventLog = append(match.EventLog, logs.PlayContinuousLog{
					UpdateID:      match.NextUpdateID,
					WallTimestamp: wallTime,
					Type:          logs.PlayContinuous,
					TeamIndex:     card.Team,
					CardIndex:     reactionTop.CardIndex,
				})
				match.NextUpdateID++
			} else if card.Type == state.CardTypeSpell { // if a spell is successfully summoned, check for triggers
				card.Zone = state.ZoneGraveyard
				gs.Grave[card.Team] = append(gs.Grave[card.Team], reactionTop.CardIndex)
				trigger.FireAllyTrigger(gs, &match.EventLog, &match.NextUpdateID, card.Team, state.TriggerAllySpellResolve, wallTime)
			} else {
				card.Zone = state.ZoneGraveyard
				gs.Grave[card.Team] = append(gs.Grave[card.Team], reactionTop.CardIndex)
			}
		} else if reactionTop.Type == state.ReactionContinuousAbility {
			card.PendingCharges--
			card.Charges--
			match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
				UpdateID:         match.NextUpdateID,
				WallTimestamp:    wallTime,
				Type:             logs.RemoveCharge,
				TeamIndex:        card.Team,
				CardIndex:        reactionTop.CardIndex,
				ChargesRemaining: card.Charges,
			})
			match.NextUpdateID++
			if card.Charges <= 0 {
				for i, id := range gs.Continuous[card.Team] {
					if id == reactionTop.CardIndex {
						gs.Continuous[card.Team] = append(gs.Continuous[card.Team][:i], gs.Continuous[card.Team][i+1:]...)
						break
					}
				}
				card.Zone = state.ZoneGraveyard
				gs.Grave[card.Team] = append(gs.Grave[card.Team], reactionTop.CardIndex)
				match.EventLog = append(match.EventLog, logs.RemoveContinuousLog{
					UpdateID:      match.NextUpdateID,
					WallTimestamp: wallTime,
					Type:          logs.RemoveContinuous,
					TeamIndex:     card.Team,
					CardIndex:     reactionTop.CardIndex,
				})
				match.NextUpdateID++
			}
		} else if reactionTop.Type == state.ReactionUnitAbility && reactionTop.ActiveIndex >= 0 { // only for Active abilities
			activeDef := defs.CardDefs[card.DefNo].Active[reactionTop.ActiveIndex]
			card.PendingCharges--

			if !activeDef.Unlimited {
				wasAtMax := card.Charges == card.MaxCharges
				card.Charges--
				match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
					UpdateID:         match.NextUpdateID,
					WallTimestamp:    wallTime,
					Type:             logs.RemoveCharge,
					TeamIndex:        card.Team,
					CardIndex:        reactionTop.CardIndex,
					ChargesRemaining: card.Charges,
				})
				match.NextUpdateID++

				// if the card was at max charges, the gain-charge OnTick ability was paused — restart it now
				if wasAtMax {
					startAtSimTime := wallTime - gs.SimPauseOffset
					for tickIndex, r := range defs.CardDefs[card.DefNo].OnTick {
						for _, eff := range r.Effects {
							if eff.Type == state.EffectChargeSelf {
								schedule.ScheduleOnTickAbility(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, reactionTop.CardIndex, tickIndex, startAtSimTime, wallTime)
								break
							}
						}
					}
				}
			}
		}
	}

	match.EventLog = append(match.EventLog, logs.RemoveStackLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:          logs.RemoveStack,
		CardIndex:     reactionTop.CardIndex,
	})
	match.NextUpdateID++

	if len(gs.ReactionStack) == 0 {
		match.TransitionPauseState(state.PauseTransitionUnpause, wallTime, -1)
		match.processAutoSummonQueue(wallTime, -1)
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

		scheduleReactionStackResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}
