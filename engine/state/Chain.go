package state

const (
	StackItemCard                 = 0
	StackItemAbility              = 1
	StackItemSpell                = 2
	StackItemContinuousActivation = 3
	StackItemSummon               = 4
	StackItemUnitEffect           = 5
)

type StackItem struct {
	Type                int
	CardID              int
	ActiveIndex         int // which Actives[] slot this activation came from, -1 if not Active-ability-sourced
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	Effects             []Effect
}
