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
