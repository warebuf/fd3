package chain

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// duplicates conditions/trigger.CheckAllyTriggers' zone-flag switch locally: that package already
// imports chain (for PushTriggerEffectsOnChain), so chain can't import it back without a cycle
func triggerZoneFlag(zone int) int {
	switch zone {
	case state.ZoneField:
		return defs.TriggerZoneField
	case state.ZoneGraveyard:
		return defs.TriggerZoneGraveyard
	case state.ZoneHand:
		return defs.TriggerZoneHand
	default:
		return 0
	}
}

// fireOpponentTrigger fires OnTrigger abilities sitting on actingTeam's opponent's board that react
// to actingTeam's own activation (e.g. a mana-tax effect like Extortionist). Unlike an ally trigger,
// the team passed down to the effect template is actingTeam, not the reacting card's own team, so
// effects like EffectExtort know who to tax rather than who's reacting.
func fireOpponentTrigger(gs *state.GameState, logList *[]any, UpdateID *int, actingTeam int, triggerType state.TriggerType, executeAtWallTime int64) {
	opponentTeam := 1 - actingTeam

	var candidates []int
	candidates = append(candidates, gs.Field[opponentTeam]...)
	candidates = append(candidates, gs.Grave[opponentTeam]...)
	candidates = append(candidates, gs.Hand[opponentTeam]...)

	for _, cardIndex := range candidates {
		card := gs.Cards[cardIndex]
		zoneFlag := triggerZoneFlag(card.Zone)
		def := defs.CardDefs[card.DefNo]

		for triggerIndex, ability := range def.OnTrigger {
			if ability.Trigger != triggerType || ability.TriggerZones&zoneFlag == 0 {
				continue
			}
			PushTriggerEffectsOnChain(gs, logList, UpdateID, actingTeam, cardIndex, triggerIndex, executeAtWallTime)
		}
	}
}
