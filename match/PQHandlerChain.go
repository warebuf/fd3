package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions/trigger"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
)

func (match *Match) handleChainResolution() {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	gs := match.GameState

	if len(gs.Chain) == 0 {
		return
	}

	top := gs.Chain[len(gs.Chain)-1]
	wallTime := top.ExecuteAtActiveTime + gs.FullPauseOffset

	// get prio team
	winnerTeam := -1
	if match.PendingChains[gs.Priority].Type != 0 {
		winnerTeam = gs.Priority
	} else if match.PendingChains[1-gs.Priority].Type != 0 {
		winnerTeam = 1 - gs.Priority
	}

	if winnerTeam >= 0 {
		p := match.PendingChains[winnerTeam]
		match.PendingChains = make([]PendingChain, len(match.PendingChains)) // clear both slots; the loser (if any) simply fails

		oldLogsSize := len(match.EventLog)

		match.EventLog = append(match.EventLog, logs.ClearPendingChainsLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: wallTime,
			Type:          logs.ClearPendingChains,
		})
		match.NextUpdateID++

		ok := false
		switch p.Type {
		case PendingChainCard:
			ok = chain.PushCardOnChain(gs, &match.EventLog, &match.NextUpdateID, winnerTeam, p.CardIndex, wallTime-constants.LEAD)
		case PendingChainAbility:
			ok = chain.PushAbilityOnChain(gs, &match.EventLog, &match.NextUpdateID, p.CardIndex, p.ActiveIndex, wallTime)
		case PendingChainTrap:
			ok = chain.PushTrapOnChain(gs, &match.EventLog, &match.NextUpdateID, p.CardIndex, p.ActiveIndex, wallTime)
		case PendingChainStructureDraw:
			chain.PushStructureDrawOnChain(gs, &match.EventLog, &match.NextUpdateID, winnerTeam, wallTime)
			ok = true
		}

		// state is frozen during an open window, so this should never fail; fall through to
		// resolving the top of the chain as if nobody had submitted if it somehow does.
		if ok {
			scheduleChainResolution(match)
			match.publish(wallTime, oldLogsSize)
			return
		}
	}

	// pop the top of the chain
	chainTop := gs.Chain[len(gs.Chain)-1]
	gs.Chain = gs.Chain[:len(gs.Chain)-1]

	oldLogsSize := len(match.EventLog)

	// apply effects
	if chainSourceStillValid(gs, chainTop) {
		for _, eff := range chainTop.Effects {
			sim.ApplyEffect(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, eff, chainTop.Effects, chainTop.CardIndex, chainTop.ActiveIndex, wallTime)
		}
	}

	if chainTop.Type != state.ChainStructureAbility {
		card := &gs.Cards[chainTop.CardIndex]
		switch chainTop.Type {
		case state.ChainSummon:
			card.Zone = state.ZoneField
			gs.Field[card.Team] = append(gs.Field[card.Team], chainTop.CardIndex)
			match.EventLog = append(match.EventLog, logs.SummonResolutionLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.SummonResolution,
				TeamIndex:     card.Team,
				CardIndex:     chainTop.CardIndex,
			})
			match.NextUpdateID++
			schedule.ScheduleAllOnTicks(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, chainTop.CardIndex, match.GameState.SimPauseStartedAt-gs.SimPauseOffset, wallTime)

		case state.ChainContinuousResolution:
			card.Zone = state.ZoneContinuous
			gs.Continuous[card.Team] = append(gs.Continuous[card.Team], chainTop.CardIndex)
			match.EventLog = append(match.EventLog, logs.PlayContinuousLog{
				UpdateID:      match.NextUpdateID,
				WallTimestamp: wallTime,
				Type:          logs.PlayContinuous,
				TeamIndex:     card.Team,
				CardIndex:     chainTop.CardIndex,
			})
			match.NextUpdateID++

		case state.ChainTrap: // if a trap successfully resolves, check for triggers
			if card.Zone == state.ZoneTrap {
				for i, ci := range gs.Traps[card.Team] {
					if ci == chainTop.CardIndex {
						gs.Traps[card.Team] = append(gs.Traps[card.Team][:i], gs.Traps[card.Team][i+1:]...)
						break
					}
				}
				card.Zone = state.ZoneGraveyard
				gs.Grave[card.Team] = append(gs.Grave[card.Team], chainTop.CardIndex)
				trigger.FireAllyTrigger(gs, &match.EventLog, &match.NextUpdateID, card.Team, state.TriggerAllyTrapResolve, wallTime)
			}

		case state.ChainContinuousAbility:
			card.PendingCharges--
			card.Charges--
			match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
				UpdateID:         match.NextUpdateID,
				WallTimestamp:    wallTime,
				Type:             logs.RemoveCharge,
				TeamIndex:        card.Team,
				CardIndex:        chainTop.CardIndex,
				ChargesRemaining: card.Charges,
			})
			match.NextUpdateID++
			if card.Charges <= 0 {
				for i, id := range gs.Continuous[card.Team] {
					if id == chainTop.CardIndex {
						gs.Continuous[card.Team] = append(gs.Continuous[card.Team][:i], gs.Continuous[card.Team][i+1:]...)
						break
					}
				}
				card.Zone = state.ZoneGraveyard
				gs.Grave[card.Team] = append(gs.Grave[card.Team], chainTop.CardIndex)
				match.EventLog = append(match.EventLog, logs.RemoveContinuousLog{
					UpdateID:      match.NextUpdateID,
					WallTimestamp: wallTime,
					Type:          logs.RemoveContinuous,
					TeamIndex:     card.Team,
					CardIndex:     chainTop.CardIndex,
				})
				match.NextUpdateID++
			}

		case state.ChainUnitAbility:
			if chainTop.ActiveIndex >= 0 { // only for Active abilities
				activeDef := defs.CardDefs[card.DefNo].Active[chainTop.ActiveIndex]
				card.PendingCharges--

				if !activeDef.Unlimited {
					wasAtMax := card.Charges == card.MaxCharges
					card.Charges--
					match.EventLog = append(match.EventLog, logs.RemoveChargeLog{
						UpdateID:         match.NextUpdateID,
						WallTimestamp:    wallTime,
						Type:             logs.RemoveCharge,
						TeamIndex:        card.Team,
						CardIndex:        chainTop.CardIndex,
						ChargesRemaining: card.Charges,
					})
					match.NextUpdateID++

					// if the card was at max charges, the gain-charge OnTick ability was paused — restart it now
					if wasAtMax {
						startAtSimTime := wallTime - gs.SimPauseOffset
						for tickIndex, r := range defs.CardDefs[card.DefNo].OnTick {
							for _, eff := range r.Effects {
								if eff.Type == state.EffectChargeSelf {
									schedule.ScheduleOnTickAbility(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, chainTop.CardIndex, tickIndex, startAtSimTime, wallTime)
									break
								}
							}
						}
					}
				}
			}
		}
	}

	match.EventLog = append(match.EventLog, logs.RemoveChainLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: wallTime,
		Type:          logs.RemoveChain,
		CardIndex:     chainTop.CardIndex,
	})
	match.NextUpdateID++

	if len(gs.Chain) == 0 {
		match.TransitionPauseState(state.PauseTransitionUnpause, wallTime, -1)
		match.processAutoSummonQueue(wallTime, -1)
	} else { // the new top gets a refreshed chain window
		newTopIndex := len(gs.Chain) - 1
		freshStart := wallTime - gs.FullPauseOffset
		gs.Chain[newTopIndex].StartAtActiveTime = freshStart
		gs.Chain[newTopIndex].PopAtActiveTime = freshStart + constants.CHAIN_WINDOW - constants.LEAD
		gs.Chain[newTopIndex].ExecuteAtActiveTime = freshStart + constants.CHAIN_WINDOW

		match.EventLog = append(match.EventLog, logs.RefreshChainLog{
			UpdateID:            match.NextUpdateID,
			WallTimestamp:       wallTime,
			Type:                logs.RefreshChain,
			CardIndex:           gs.Chain[newTopIndex].CardIndex,
			StartAtActiveTime:   gs.Chain[newTopIndex].StartAtActiveTime,
			PopAtActiveTime:     gs.Chain[newTopIndex].PopAtActiveTime,
			ExecuteAtActiveTime: gs.Chain[newTopIndex].ExecuteAtActiveTime,
		})
		match.NextUpdateID++

		scheduleChainResolution(match)
	}

	match.publish(wallTime, oldLogsSize)

}

// whether chainTop's source card is still where its ability expects it to be; false means it fizzles
func chainSourceStillValid(gs *state.GameState, chainTop state.ChainItem) bool {
	if chainTop.CardIndex < 0 {
		return true // no source card, e.g. a structure ability
	}
	if chainTop.Type == state.ChainSummon || chainTop.Type == state.ChainContinuousResolution {
		return true // the card hasn't reached its zone yet; that's expected, not a fizzle
	}
	card := gs.Cards[chainTop.CardIndex]
	switch card.Type {
	case state.CardTypeUnit:
		return card.Zone == state.ZoneField
	case state.CardTypeContinuous:
		return card.Zone == state.ZoneContinuous
	case state.CardTypeTrap:
		return card.Zone == state.ZoneTrap
	}
	return true
}
