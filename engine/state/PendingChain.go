package state

type PendingChainType int

const (
	PendingChainCard PendingChainType = iota + 1
	PendingChainAbility
	PendingChainStructureDraw
)

type PendingChain struct {
	Type        PendingChainType
	CardIndex   int
	ActiveIndex int // -1 for card plays
}
