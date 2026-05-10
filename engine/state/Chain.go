package state

const (
	ChainItemCard    = 0
	ChainItemAbility = 1
	ChainItemSpell   = 2
)

type ChainItem struct {
	Type      int
	CardID    int   // which card this originates from
	StartTime int64 // when it was scheduled (denominated in ActiveTime ms)
	PopTime   int64 // when the server pops it (early, to broadcast in time)
	OccurTime int64 // when the effect actually fires (denominated in ActiveTime ms)

	Effects []Effect
}
