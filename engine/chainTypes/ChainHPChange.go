package chainTypes

// Declaration -> Reaction -> Resolution

type HPChangeDeclaration struct {

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

	// status
	Cancelled bool // group level
}

func (HPChangeDeclaration) Kind() string { return "HPChangeDeclaration" }

type HPChangeWindow struct {
	Parent      *HPChangeWindow
	TargetIndex int
}

func (HPChangeWindow) Kind() string { return "HPChangeWindow" }

type HPChangeResolution struct {
	Parent      *HPChangeWindow
	DamageDealt int

	Negated bool // individual level
}

func (HPChangeResolution) Kind() string { return "HPChangeResolution" }
