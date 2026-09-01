package logs

import "github.com/warebuf/fd3/engine/state"

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
	Heal
	Ready
	Unready
	AutoSummonIndex
	Pause
	Unpause
	AddTickAbility
	ClearTickAbility
	GameOver
	CounterChain
	RefreshChain
	PlayContinuous
	RemoveContinuous
	RemoveCharge
	AddCharge
	Preselect
	ToggleAutoAbility
	Consume
	SpawnToken
	ToggleStructureAutoAbility
	SetTrap
	TrapReady
	PurgeChain
	EquipResolution
	TrapResolution
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

type EquipResolutionLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	HostIndex     int // -1 if the preselected host was gone by the time this resolved
}

type RemoveChainLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardIndex     int
}

type AddChainLog struct {
	WallTimestamp       int64
	UpdateID            int
	Type                LogType
	ChainType           int
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
	CardIndex           int
	ActiveIndex         int // which Actives[] slot this item's activation came from, -1 if not an Active ability
	TeamIndex           int
	Effects             []state.Effect
}

type RefreshChainLog struct {
	WallTimestamp       int64
	UpdateID            int
	Type                LogType
	CardIndex           int
	StartAtActiveTime   int64
	PopAtActiveTime     int64
	ExecuteAtActiveTime int64
}

type DestroyCardLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
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
	TargetTeamIndex int
	TargetCardIndex int // -1 if targeting a structure, meaning a unit was not the target
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
	AbilityIndex  int
}

type AddTickAbilityLog struct {
	WallTimestamp    int64
	UpdateID         int
	Type             LogType
	CardIndex        int
	AbilityIndex     int
	Condition        state.TickCondition
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

type CounterChainLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	SourceCardIndex int // the countering card
	TargetCardIndex int // the chain item's source card being countered
}

// a trap leaving the hand for the trap zone, where it sets until it's ready
type SetTrapLog struct {
	WallTimestamp     int64
	UpdateID          int
	Type              LogType
	TeamIndex         int
	CardIndex         int
	SetStartAtSimTime int64
	ReadyAtSimTime    int64
}

// a set trap finishing its setting time, so it may now be activated
type TrapReadyLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}

type PlayContinuousLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}

type RemoveContinuousLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}

type RemoveChargeLog struct {
	WallTimestamp    int64
	UpdateID         int
	Type             LogType
	TeamIndex        int
	CardIndex        int
	ChargesRemaining int
}

type AddChargeLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	Charges       int
}

type PreselectLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	TeamIndex       int
	CardIndex       int
	AbilityIndex    int
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

type ToggleStructureAutoAbilityLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	On            bool
}

type ConsumeLog struct {
	WallTimestamp   int64
	UpdateID        int
	Type            LogType
	SourceCardIndex int
	TargetCardIndex int
	AbilityIndex    int // which Active slot's Bonus was credited
	Bonus           int // resulting total Bonus on that slot
}

type SpawnTokenLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
	DefNo         int
	Name          string
	Description   string
	HP            int
	Cost          int
	Archetype     int
	CardType      int
	IsToken       bool
	ChainEntry    state.ChainEntry
}

type PurgeChainLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardIndex     int
}

// a set trap finishing its resolution, leaving the trap zone for the graveyard
type TrapResolutionLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}
