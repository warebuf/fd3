package state

type ReactionType int

const (
	ReactionUnitAbility ReactionType = iota + 1
	ReactionSpell
	ReactionContinuousAbility
	ReactionSummon
	ReactionContinuousResolution
)

type ReactionItem struct {
	Type                ReactionType
	CardIndex           int
	ActiveIndex         int // which Actives[] slot this activation came from, -1 if not an Active ability
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	Effects             []Effect
}
