package trigger

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// converts the zone int into bitmask version
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

func CheckAllyTriggers(gs *state.GameState, teamIndex int, trigger state.TriggerType) ([]int, []int) {
	var cardIndexes []int
	var triggerIndexes []int

	var candidates []int
	candidates = append(candidates, gs.Field[teamIndex]...)
	candidates = append(candidates, gs.Grave[teamIndex]...)
	candidates = append(candidates, gs.Hand[teamIndex]...)

	for _, cardIndex := range candidates {
		card := gs.Cards[cardIndex]
		zoneFlag := triggerZoneFlag(card.Zone)
		def := defs.CardDefs[card.DefNo]

		for triggerIndex, ability := range def.OnTrigger {
			if ability.Trigger != trigger || ability.TriggerZones&zoneFlag == 0 {
				continue
			}
			cardIndexes = append(cardIndexes, cardIndex)
			triggerIndexes = append(triggerIndexes, triggerIndex)
		}
	}

	return cardIndexes, triggerIndexes
}
