package match

import (
	"github.com/warebuf/fd3/engine/logs"
)

type PendingChainType int

const (
	PendingChainCard PendingChainType = iota + 1
	PendingChainAbility
	PendingChainStructureDraw
	PendingChainTrap
)

type PendingChain struct {
	Type        PendingChainType
	CardIndex   int // -1 for a structure draw, which has no card
	ActiveIndex int // -1 for card plays
}

// TODO: broadcast for now - this tells the opponent a reaction is waiting, and needs
// per-client censoring before it's correct
func (match *Match) submitPendingChain(teamIndex int, p PendingChain, executeAtWallTime int64) {
	oldLogsSize := len(match.EventLog)

	match.PendingChains[teamIndex] = p

	match.EventLog = append(match.EventLog, logs.SubmitChainLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SubmitChain,
		TeamIndex:     teamIndex,
		CardIndex:     p.CardIndex,
		PendingType:   int(p.Type),
	})
	match.NextUpdateID++

	match.publish(executeAtWallTime, oldLogsSize)
}
