package logs

import "github.com/warebuf/fd3/engine/state"

type LogType int

const (
	PhaseLog LogType = iota
	DrawExec
	DeclareCard
	SummonUnit
	ScheduleChain
	RemoveChain
	DestroyCard
	GainMana
)

// honestly Entry is not the right name, it should be log

type PhaseLogEntry struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	Phase     int
}

type DrawExecEntry struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	NumCards  int
}

type DeclareCardEntry struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	CardIndex int
}

type SummonUnitEntry struct {
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

type DestroyCardEntry struct {
	Timestamp       int64
	UpdateID        int
	Type            LogType
	TeamIndex       int
	CardIndex       int
	TargetAll       bool
	TargetTeamIndex int
	TargetCardIndex int
}

type GainManaEntry struct {
	Timestamp int64
	UpdateID  int
	Type      LogType
	TeamIndex int
	Amount    int
}
