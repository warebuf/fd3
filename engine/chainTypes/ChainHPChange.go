package chainTypes

// Declaration -> Reaction -> Resolution

type HPChangeDeclaration struct {

	// parent
	Parent ChainNode // an interface is naturally a pointer

	// base info
	CasterIndex int
	Amount      int

	// type
	Heal       bool // attack if false
	Sequential bool // if not sequential, is concurrent

	// target related
	TargetType    string // custom, enemies, team, teamxself, all, allxself
	All           bool   // all units in the TargetType set
	TargetCount   int    // doesn't matter what this is if All is true
	TargetIndexes []int  // if custom, which unitIndex

	// reactions
	Reactions []ChainNode
	Cancelled bool
}

func (HPChangeDeclaration) Kind() string { return "HPChangeDeclaration" }

type HPChangeWindow struct {
	Parent      *HPChangeDeclaration
	TargetIndex int
}

func (HPChangeWindow) Kind() string { return "HPChangeWindow" }

type HPChangeResolution struct {
	Parent *HPChangeWindow
	Amount int

	Negated bool // individual level
}

func (HPChangeResolution) Kind() string { return "HPChangeResolution" }
