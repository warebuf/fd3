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
)

type Effect struct {
	Type EffectType
	A    int
	B    int
	C    int
}

func FulfillEffectTemplate(gs *GameState, templ Effect, teamIndex int, cardIndex int, activeIndex int) Effect {
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
		if activeIndex >= 0 {
			templ.A = gs.Cards[cardIndex].Actives[activeIndex].Target
		}
		return templ
	}
	return templ
}
