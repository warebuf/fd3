package logs

import "github.com/warebuf/fd3/engine/state"

type LogType int

const (
	PhaseLog LogType = iota
	DrawExec
	DeclareCard
	SummonUnit
	ScheduleStack
)

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

type ScheduleStackLog struct {
	Timestamp  int64
	UpdateID   int
	Type       LogType
	StackID    int
	StackEntry state.StackEntry
}
