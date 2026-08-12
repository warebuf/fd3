package state

type ChainType int

const (
	ChainUnitAbility ChainType = iota + 1
	ChainSpell
	ChainContinuousAbility
	ChainSummon
	ChainContinuousResolution
	ChainStructureAbility
)

type ChainEntry int

const (
	ChainDual     ChainEntry = iota // can be an action or a reaction, the default so nothing is silently refused
	ChainAction                     // can only start a chain
	ChainReaction                   // can only join an existing chain
)

// whether this may go on the chain as it currently stands
func (e ChainEntry) CanEnter(gs *GameState) bool {
	switch e {
	case ChainAction:
		return len(gs.Chain) == 0
	case ChainReaction:
		return len(gs.Chain) > 0
	}
	return true
}

type ChainItem struct {
	Type                ChainType
	CardIndex           int
	ActiveIndex         int // which Actives[] slot this activation came from, -1 if not an Active ability
	TeamIndex           int
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	Effects             []Effect
}
