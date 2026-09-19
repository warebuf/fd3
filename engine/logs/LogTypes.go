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
	SkipChain
	UnskipChain
	GiveControl
	StandoffOpen
	StandoffBid
	StandoffResolution
	Discard
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
	StructureIndex      int // which structure activated, -1 if this item didn't come from a structure
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
	WallTimestamp        int64
	UpdateID             int
	Type                 LogType
	AttackerIndex        int
	TargetTeamIndex      int
	TargetCardIndex      int // -1 if targeting a structure, meaning a unit was not the target
	TargetStructureIndex int // which structure was hit, -1 if a unit was the target
	Damage               int
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

// a team waiving its reaction to the current chain top; once every team has, the window collapses
type SkipChainLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
}

type UnskipChainLog struct {
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
	WallTimestamp  int64
	UpdateID       int
	Type           LogType
	TeamIndex      int
	StructureIndex int
	On             bool
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

// a field unit changed controller; it moves between the two Field lists and its Team follows
type GiveControlLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardIndex     int
	FromTeamIndex int
	ToTeamIndex   int
}

// a Standoff-style trap opened a blind mana auction; the card leaves its trap zone to await both bids
type StandoffOpenLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int // the team that opened it
	CardIndex     int
}

// one team sealed their bid. Deliberately carries no Amount - that stays secret until
// StandoffResolutionLog reveals it, since this log is broadcast to both clients unfiltered
type StandoffBidLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
}

// both bids are in; reveals both amounts and the winner now that it's safe to. The loser's forced
// discard, if their hand wasn't empty, arrives as its own separate DiscardLog in the same batch
type StandoffResolutionLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	CardIndex     int
	Bids          [2]int
	WinnerTeam    int
}

// a card was forced out of a hand and into the graveyard without being played, e.g. by losing a
// Standoff
type DiscardLog struct {
	WallTimestamp int64
	UpdateID      int
	Type          LogType
	TeamIndex     int
	CardIndex     int
}
