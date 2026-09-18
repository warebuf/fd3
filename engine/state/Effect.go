package state

type EffectType int

const (
	EffectAttack EffectType = iota + 1
	EffectHeal
	EffectDraw
	EffectMana       // A: team, B: amount
	EffectDestroy    // A: Destroy Type (0: All, 1: team specific, 2: card specific), B: target (if team specific, team index, if card specific, card index)
	EffectCounter    // cancel the top chain item
	EffectChargeSelf // +1 charge on the source card itself, capped at its MaxCharges; no fields used
	EffectSacrifice  // A: card index
	EffectConsume    // A: card index (from Card.OnPlay[abilityIndex].Target); destroys it. C: which of this card's Active slots gets its Charges as Bonus
	EffectSelfDestruct
	EffectConvertToTokens   // A: card index (from Card.Actives[abilityIndex].Target); destroys it and converts its Charges into tokens. B: DefNo of the token CardDef to spawn, one per 2 Charges
	EffectChargeTarget      // A: card index (from Card.Actives[abilityIndex].Target); gains B charges, capped at its MaxCharges. C: DefNo of the source card, so the applier can switch on which card's specific fizzle condition applies
	EffectAbsorbDestruction // only used in an equip card's OnTrigger[TriggerHostWouldBeDestroyed]; destroys the source card instead of the card that would've been destroyed
	EffectGiveControl       // A: card index (from Card.Actives[abilityIndex].Target); hands that friendly field unit to the other team
	EffectHealNonDamaging   // B: amount; every unit on the field (both sides) whose CardDef has no EffectAttack in any ability slot gains this much HP
	EffectExtort            // only used in OnTrigger[TriggerEnemyPaidMana]; A: the team that just paid mana (auto-filled), B: toll amount taken from them if they can afford it, granted to the source card's controller either way
)

type Effect struct {
	Type EffectType
	A    int
	B    int
	C    int
}

// abilityIndex indexes Card.Actives for the effects built from an Active slot, or Card.OnPlay for
// EffectConsume, which is the only one still built from an OnPlay ability. Pass -1 when the calling
// site has neither (e.g. OnTick, OnTrigger).
func FulfillEffectTemplate(gs *GameState, templ Effect, teamIndex int, cardIndex int, abilityIndex int) Effect {
	switch templ.Type {
	case EffectMana:
		templ.A = teamIndex
		return templ
	case EffectDraw:
		templ.A = teamIndex
		return templ
	case EffectAttack:
		return templ
	case EffectHeal:
		return templ
	case EffectSacrifice:
		if abilityIndex >= 0 {
			templ.A = gs.Cards[cardIndex].Actives[abilityIndex].Target
		}
		return templ
	case EffectConsume:
		if abilityIndex >= 0 {
			templ.A = gs.Cards[cardIndex].OnPlay[abilityIndex].Target
		}
		return templ
	case EffectConvertToTokens:
		if abilityIndex >= 0 {
			templ.A = gs.Cards[cardIndex].Actives[abilityIndex].Target
		}
		return templ
	case EffectChargeTarget:
		if abilityIndex >= 0 {
			templ.A = gs.Cards[cardIndex].Actives[abilityIndex].Target
		}
		templ.C = gs.Cards[cardIndex].DefNo
		return templ
	case EffectGiveControl:
		if abilityIndex >= 0 {
			templ.A = gs.Cards[cardIndex].Actives[abilityIndex].Target
		}
		return templ
	case EffectExtort:
		templ.A = teamIndex
		return templ
	}
	return templ
}
