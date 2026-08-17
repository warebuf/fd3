package match

type PendingChainType int

const (
	PendingChainCard PendingChainType = iota + 1
	PendingChainAbility
	PendingChainStructureDraw
	PendingChainTrap
)

type PendingChain struct {
	Type        PendingChainType
	CardIndex   int
	ActiveIndex int // -1 for card plays
}
