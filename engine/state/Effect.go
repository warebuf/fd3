package state

type EffectType int

const (
	EffectAttack EffectType = iota + 1
	EffectHeal
	EffectDraw
	EffectMana // A: team, B: amount
	EffectSilence
	EffectGrantAbility
	EffectDestroy // A: Destroy Type (0: All, 1: team specific, 2: card specific), B: target (if team specific, team index, if card specific, card index)
	EffectCounter    // cancel the top chain item
	EffectGainCharge // gain 1 charge on the source card, up to MaxCharges
)

type Effect struct {
	Type EffectType
	A    int
	B    int
	C    int
}
