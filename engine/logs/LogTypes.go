package logs

type LogType int

const (
	Phase LogType = iota
	Draw
	DeclareCard
	SummonResolution
	AddSimItem
	RemoveSimItem
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
	Type      LogType
	TeamIndex int
	NumCards  int
}

type DeclareCardLog struct {
	WallTimestamp int64
	UpdateID      int
	Type      LogType
	TeamIndex int
	CardIndex int
}

type SummonResolutionLog struct {
	WallTimestamp int64
	UpdateID      int
	Type      LogType
	TeamIndex int
	CardIndex int
}

type AddSimItemLog struct {
	WallTimestamp int64
	UpdateID      int
	Type        LogType
	SeqID       int
	StartTime   int64
	OccurTime   int64
	CardID      int
	SimItemType int
}

type RemoveSimItemLog struct {
	WallTimestamp int64
	UpdateID      int
	Type      LogType
	CardID    int
}

type RemoveChainLog struct {
	WallTimestamp int64
	UpdateID      int
	Type      LogType
	CardID    int
}

type AddChainLog struct {
	WallTimestamp int64
	UpdateID      int
	Type      LogType
	StartTime int64
	OccurTime int64
	CardID    int
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
	Type      LogType
	TeamIndex int
	Amount    int
}

type SpendManaLog struct {
	WallTimestamp int64
	UpdateID      int
	Type      LogType
	TeamIndex int
	Amount    int
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
	Type      LogType
	TeamIndex int
}

type AutoSummonLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	AutoSummonIdx int // 0 when cleared, assigned order when set
}

type PauseLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	TimeRemaining int64 // ms remaining in this team's time bank at pause start
}

type UnpauseLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	TimeRemaining int64 // ms remaining in this team's time bank after unpause
}
