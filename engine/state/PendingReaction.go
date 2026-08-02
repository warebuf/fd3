package state

type PendingReactionType int

const (
	PendingReactionCard PendingReactionType = iota + 1
	PendingReactionAbility
	PendingReactionStructureDraw
)

type PendingReaction struct {
	Type        PendingReactionType
	CardIndex   int
	ActiveIndex int // -1 for card plays
}
