package defs

func NewBloodPriest() Card {
	return Card{
		OnPlay: []Effect{
			DealDamage(EnemyTarget, 2),
			Heal(SelfTarget, 2),
		},
		Triggers: []Trigger{
			OnAllyDeath(
				GainAttack(SelfTarget, 1),
			),
		},
	}
}
