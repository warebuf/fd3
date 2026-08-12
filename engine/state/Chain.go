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
