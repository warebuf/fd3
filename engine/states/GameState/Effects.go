package gameState

type EffectType int

const (
	EffectAttack EffectType = iota + 1
	EffectHeal
	EffectDraw
	EffectMana
	EffectSilence
	EffectGrantAbility
)

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
