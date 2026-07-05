package logs

import "github.com/warebuf/fd3/engine/state"

type LogType int

const (
	Phase LogType = iota
	Draw
	DeclareCard
	SummonResolution
	AddStack
	RemoveStack
	DestroyCard
	GainMana
	SpendMana
	Attack
	Heal
	Ready
	Unready
	StageIndex
	AutoSummonIndex
	Pause
	Unpause
	AddTickAbility
	ClearTickAbility
	GameOver
	CounterStack
	RefreshStack
	PlayContinuous
	RemoveContinuous
	RemoveCharge
	AddCharge
	Preselect
	ToggleAutoAbility
)

type PhaseLog struct {
	WallTimestamp        int64
	UpdateID             int
	Type                 LogType
	Phase                int
	PhaseEndInActiveTime int64 // active time (ms) when this phase ends; 0 if no fixed end
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

type RemoveStackLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardID        int
}

type AddStackLog struct {
	WallTimestamp       int64
	UpdateID            int
	Type                LogType
	StackType           int
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	CardID              int
	Effects             []state.Effect
}

type RefreshStackLog struct {
	WallTimestamp       int64
	UpdateID            int
	Type                LogType
	CardID              int
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
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

type HealLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	HealerIndex     int
	TargetCardIndex int
	Amount          int
}

type ReadyLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
}

type UnreadyLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
}

type StageLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	StageOrder    int // 0 when cleared, assigned order when set
}

type AutoSummonLog struct {
	WallTimestamp      int64
	UpdateID           int
	Type               LogType
	TeamIndex          int
	CardIndex          int
	AutoSummonPriority int // 0 when cleared, assigned priority when set
}

type ClearTickAbilityLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardIndex     int
	AbilityID     int
}

type AddTickAbilityLog struct {
	WallTimestamp    int64
	UpdateID         int
	Type             LogType
	CardIndex        int
	AbilityID        int
	AbilityIndex     int
	Condition        state.TickCondition
	ActiveIndex      int
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

type GameOverLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	Winner        int // winning team index, -2 = tie
}

type CounterStackLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	SourceCardID  int // the counterspell card
	TargetCardID  int // the spell being countered
}

type PlayContinuousLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardID        int
	Charges       int
}

type RemoveContinuousLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardID        int
}

type RemoveChargeLog struct {
	WallTimestamp    int64
	UpdateID         int
	Type             LogType
	TeamIndex        int
	CardID           int
	ChargesRemaining int
}

type AddChargeLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardID        int
	Charges       int
}

type PreselectLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	TeamIndex       int
	CardID          int
	TargetCardIndex int
}

type ToggleAutoAbilityLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	AbilityIndex  int
	On            bool
}
