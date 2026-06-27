package state

const (
	StackItemCard                = 0
	StackItemAbility             = 1
	StackItemSpell               = 2
	StackItemContinuousActivation = 3
	StackItemSummon              = 4
)

type StackItem struct {
	Type                int
	CardID              int
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	Effects             []Effect
}
