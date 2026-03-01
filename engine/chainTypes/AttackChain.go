package chainTypes

// Declaration -> Reaction -> Resolution

type ConcurrentAttackDeclaration struct {
	AttackerIndex int
	Damage        int
	TargetIndexes []int

	Cancelled bool // group level
}

func (ConcurrentAttackDecalaration) Kind() string { return "ConcurrentAttackDeclaration" }

type ConcurrentAttackWindow struct {
	Parent      *ConcurrentAttackDeclaration
	TargetIndex int
}

func (ConcurrentAttackWindow) Kind() string { return "ConcurrentAttackWindow" }

type ConcurrentAttackResolution struct {
	Parent      *ConcurrentAttackWindow
	DamageDealt int

	Negated bool // individual level
}

func (ConcurrentAttackResolution) Kind() string { return "ConcurrentAttackResolution" }

type SequentialAttackDeclaration struct {
	AttackerIndex int
	Damage        int
	Hit           map[int]struct{}

	Cancelled bool
}

func (SequentialAttackDeclaration) Kind() string { return "SequentialAttackDeclaration" }

type SequentialAttackWindow struct {
	Parent      *SequentialAttackDeclaration
	TargetIndex int
}

func (SequentialAttackWindow) Kind() string { return "SequentialAttackWindow" }

type SequentialAttackResolution struct {
	Parent      *SequentialAttackWindow
	DamageDealt int

	Negated bool
}

func (SequentialAttackResolution) Kind() string { return "SequentialAttackResolution" }
