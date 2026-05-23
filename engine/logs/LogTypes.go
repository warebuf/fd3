package logs

type LogType int

const (
	Phase LogType = iota
	Draw
	DeclareCard
	SummonResolution
	AddChain
	RemoveChain
	DestroyCard
	GainMana
	SpendMana
	Attack
	Ready
	AutoSummonIndex
	Pause
	Unpause
	AddAbilityTimer
	ClearAbilityTimer
)

type PhaseLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	Phase         int
}

type DrawLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	NumCards      int
}

type DeclareCardLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}

type SummonResolutionLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}

type RemoveChainLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardID        int
}

type AddChainLog struct {
	WallTimestamp       int64
	UpdateID            int
	Type                LogType
	StartAtActiveTime   int64
	ExecuteAtActiveTime int64
	CardID              int
}

type DestroyCardLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	TeamIndex       int
	CardIndex       int
	TargetAll       bool
	TargetTeamIndex int
	TargetCardIndex int
}

type GainManaLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	Amount        int
}

type SpendManaLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	Amount        int
}

type AttackLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	AttackerIndex   int
	TargetIsUnit    bool
	TargetTeamIndex int
	TargetCardIndex int // -1 if targeting a structure
	Damage          int
}

type ReadyLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
}

type AutoSummonLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	AutoSummonIdx int // 0 when cleared, assigned order when set
}

type ClearAbilityTimerLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardIndex     int
	AbilityID     int
}

type AddAbilityTimerLog struct {
	WallTimestamp    int64
	UpdateID         int
	Type             LogType
	CardIndex        int
	AbilityID        int
	AbilityIdx       int
	TimerType        int
	StartAtSimTime   int64
	ExecuteAtSimTime int64
}

type PauseLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	PausePhase    int   // which pause level started: PausePhaseFullPause or PausePhaseSimPause
	TeamIndex     int   // -1 for system-initiated pauses
	TimeStart     int64 // wallTime the pause started
	TimeRemaining int64 // ms remaining in this team's time bank; 0 for system-initiated pauses
}

type UnpauseLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TimeRemaining int64 // ms remaining in this team's time bank; 0 for system-initiated pauses
}
