package gameState

const (
	EffectAttack = iota + 1
	EffectHeal
	EffectDraw
	EffectMana
	EffectSilence
	EffectGrantAbility
)

type Effect struct {
	Type       int
	TargetRule int
	Amount     int
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
