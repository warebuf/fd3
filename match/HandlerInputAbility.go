package match

import (
	"fmt"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleUseContinuous(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneContinuous || card.Team != teamIndex {
		return
	}

	// input events don't yet carry which Active slot to use, so this always targets slot 0
	// (no current card has more than one Active ability)
	const activeIndex = 0
	def := defs.CardDefs[card.DefNo].Active[activeIndex]

	if !def.Unlimited && card.PendingCharges >= card.Charges {
		return
	}

	if gs.Mana[teamIndex] < def.Cost {
		return
	}

	effs := []state.Effect{}
	for _, eff := range def.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, activeIndex))
	}
	if len(effs) == 0 {
		return // no real Use ability on this slot — don't spend mana or touch Charges
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	gs.Mana[teamIndex] -= def.Cost
	match.EventLog = append(match.EventLog, logs.SpendManaLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        def.Cost,
	})
	match.NextUpdateID++

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		Type:                state.ReactionContinuousAbility,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	gs.Cards[cardIndex].PendingCharges++
	match.EventLog = append(match.EventLog, logs.AddStackLog{
		UpdateID:            match.NextUpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	match.NextUpdateID++

	match.TransitionPauseState(1, executeAtWallTime, -1)
	scheduleReactionStackResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleUnitAbility(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneField || card.Team != teamIndex {
		return
	}

	// HARDCODE: ALWAYS TARGETS SLOT 0
	const activeIndex = 0
	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]
	if !activeDef.Unlimited && card.PendingCharges >= card.Charges {
		return
	}

	// dynamically populate the effect objects
	effs := []state.Effect{}
	for _, eff := range activeDef.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, activeIndex))
	}
	if len(effs) == 0 {
		return
	}

	// validating the effects
	for i := 0; i < len(effs); i++ {
		if effs[i].Type != state.EffectSacrifice {
			continue
		}
		fmt.Println("target index", effs[i].A)
		validTarget := false
		for j := 0; j < len(gs.Field[teamIndex]); j++ {
			if gs.Field[teamIndex][j] == effs[i].A {
				validTarget = true
				break
			}
		}
		if !validTarget {
			return
		}
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		Type:                state.ReactionUnitAbility,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	gs.Cards[cardIndex].PendingCharges++
	match.EventLog = append(match.EventLog, logs.AddStackLog{
		UpdateID:            match.NextUpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	match.NextUpdateID++

	match.TransitionPauseState(1, executeAtWallTime, -1)
	scheduleReactionStackResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}
