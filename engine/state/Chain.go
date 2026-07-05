package state

// NOTE: we need to figure out the correct name for this stuff

// TODO: these are stupid, reduce
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
	CardIndex           int
	ActiveIndex         int // which Actives[] slot this activation came from, -1 if not Active-ability-sourced
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	Effects             []Effect
}
