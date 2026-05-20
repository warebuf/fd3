package state

const (
	ChainItemCard    = 0
	ChainItemAbility = 1
	ChainItemSpell   = 2
)

type ChainItem struct {
	Type      int
	CardID    int   // which card this originates from
	StartAtActiveTime int64 // when it was scheduled (denominated in ActiveTime ms)
	PopAtActiveTime   int64 // when the server pops it (early, to broadcast in time)
	ExecuteAtActiveTime int64 // when the effect actually fires (denominated in ActiveTime ms)

	Effects []Effect
}
