package state

type ReactionType int

const (
	ReactionUnitAbility ReactionType = iota + 1
	ReactionSpell
	ReactionContinuousAbility
	ReactionSummon
)

type ReactionItem struct {
	Type                ReactionType
	CardIndex           int
	ActiveIndex         int // which Actives[] slot this activation came from, -1 if not Active-ability-sourced
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	Effects             []Effect
}
