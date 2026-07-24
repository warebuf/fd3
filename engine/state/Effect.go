package state

type EffectType int

const (
	EffectAttack EffectType = iota + 1
	EffectHeal
	EffectDraw
	EffectMana // A: team, B: amount
	EffectSilence
	EffectGrantAbility
	EffectDestroy    // A: Destroy Type (0: All, 1: team specific, 2: card specific), B: target (if team specific, team index, if card specific, card index)
	EffectCounter    // cancel the top chain item
	EffectGainCharge // A: active ability slot
	EffectSacrifice  // A: card index
	EffectConsume    // A: card index (from Card.OnPlay[abilityIndex].Target); destroys it. C: which of this card's Active slots gets its Charges as Bonus
	EffectSelfDestruct
)

type Effect struct {
	Type EffectType
	A    int
	B    int
	C    int
}

// abilityIndex indexes Card.Actives for EffectSacrifice or Card.OnPlay for EffectConsume — the two
// never overlap, since one is only ever built from an Active ability and the other only from OnPlay.
// Pass -1 when the calling site has neither (e.g. OnTick, OnTrigger).
func FulfillEffectTemplate(gs *GameState, templ Effect, teamIndex int, cardIndex int, abilityIndex int) Effect {
	switch templ.Type {
	case EffectDestroy:
		templ.C = cardIndex
		return templ
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
	}
	return templ
}
