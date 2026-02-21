package chainTypes

// Declaration -> Reaction -> Resolution

type ConcurrentAttackDeclaration struct {
	AttackerIndex int
	Damage        int
	TargetIndexes []int

	Cancelled bool // group level
}

type ConcurrentAttackWindow struct {
	Parent      *ConcurrentAttackDeclaration
	TargetIndex int
}

type ConcurrentAttackResolution struct {
	Parent      *ConcurrentAttackWindow
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
