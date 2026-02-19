package chainTypes

// Declaration -> Reaction -> Resolution

type AttackDeclaration struct {
	AttackerIndex int
	Damage        int
	TargetIndexes []int

	Cancelled bool // group level
}

type AttackWindow struct {
	Parent      *AttackDeclaration
	TargetIndex int
}

type AttackResolution struct {
	Parent      *AttackWindow
	DamageDealt int

	Negated bool // individual level
}

type SequentialAttackDeclaration struct {
	AttackerIndex int
	Damage        int
	Hit           map[int]struct{}

	Cancelled bool
}

type SequentialAttackWindow struct {
	Parent      *SequentialAttackDeclaration
	TargetIndex int
}

type SequentialAttackResolution struct {
	Parent      *SequentialAttackWindow
	DamageDealt int

	Negated bool
}

// summon

// summon reaction

// summon resolution
