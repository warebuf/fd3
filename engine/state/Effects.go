package state

type EffectType int

const (
	EffectAttack EffectType = iota + 1
	EffectActivate
	EffectHeal
	EffectDraw
	EffectMana
	EffectSilence
	EffectGrantAbility
	SummonResolution
	EffectDestroy // A: Destroy Type (0: All, 1: team specific, 2: card specific), B: target (if team specific, team index, if card specific, card index)
)

// delete later
func (e EffectType) String() string {
	switch e {
	case EffectAttack:
		return "EffectAttack"
	case EffectActivate:
		return "EffectActivate"
	case EffectHeal:
		return "EffectHeal"
	case EffectDraw:
		return "EffectDraw"
	case EffectMana:
		return "EffectMana"
	case EffectSilence:
		return "EffectSilence"
	case EffectGrantAbility:
		return "EffectGrantAbility"
	case SummonResolution:
		return "SummonResolution"
	case EffectDestroy:
		return "EffectDestroy"
	default:
		return "Unknown"
	}
}

type Effect struct {
	Type EffectType
	A    int
	B    int
	C    int
}

type TriggerEffect struct {
	Condition int
	Effect    Effect
}

type RecurringEffect struct {
	Effect   Effect
	Interval int // milliseconds
}

type Passive struct {
	Effect Effect
	Scope  int // friendly, enemy, all
}
