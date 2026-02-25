package chainTypes

// Declaration -> Reaction -> Resolution

type SummonDeclaration struct {

	// parent
	Parent ChainNode

	SummonedIndex int

	Cancelled bool // used if the condition no longer exists

	// reactions
	Reactions []ChainNode
}

func (SummonDeclaration) Kind() string { return "SummonDeclaration" }

type SummonWindow struct {
	Parent *SummonDeclaration
}

func (SummonWindow) Kind() string { return "SummonWindow" }

type SummonResolution struct {
	Parent *SummonWindow

	Negated    bool
	Bounced    bool
	HPChanged  int
	EXPChanged int
}

func (SummonResolution) Kind() string { return "SummonResolution" }
