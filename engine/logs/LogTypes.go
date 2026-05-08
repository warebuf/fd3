package logs

import "github.com/warebuf/fd3/engine/state"

type LogType int

const (
	Phase LogType = iota
	Draw
	DeclareCard
	Summon
	ScheduleChain
	RemoveChain
	DestroyCard
	GainMana
	SpendMana
	Attack
	Ready
	SetAutoSummon
	ClearAutoSummon
)

type PhaseLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	Phase     int
}

type DrawLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	NumCards  int
}

type DeclareCardLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	CardIndex int
}

type SummonLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	CardIndex int
}

type ScheduleChainLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	ChainID   int
	Step      state.ChainStep
}

type RemoveChainLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	ChainID   int
}

type DestroyCardLog struct {
	Timestamp       int64
	UpdateID        int
	Type            LogType
	TeamIndex       int
	CardIndex       int
	TargetAll       bool
	TargetTeamIndex int
	TargetCardIndex int
}

type GainManaLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	Amount    int
}

type SpendManaLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	Amount    int
}

type AttackLog struct {
	Timestamp       int64
	UpdateID        int
	Type            LogType
	AttackerIndex   int
	TargetIsUnit    bool
	TargetTeamIndex int
	TargetCardIndex int // -1 if targeting a structure
	Damage          int
}

type ReadyLog struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
}

type AutoSummonLog struct {
	Timestamp     int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	AutoSummonIdx int // 0 when cleared, assigned order when set
}
