package match

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

// opens a blind mana auction: the trap leaves its trap zone (it isn't destroyed until the auction
// resolves) and both teams now have a bid outstanding for it, tracked on gs.StandoffCardIndex
func (match *Match) openStandoffAuction(cardIndex int, teamIndex int, executeAtWallTime int64) {
	gs := match.GameState
	card := &gs.Cards[cardIndex]
	cost := card.Cost

	for i, ci := range gs.Traps[teamIndex] {
		if ci == cardIndex {
			gs.Traps[teamIndex] = append(gs.Traps[teamIndex][:i], gs.Traps[teamIndex][i+1:]...)
			break
		}
	}
	card.Ready = false
	card.SetStartAtSimTime = 0
	card.ReadyAtSimTime = 0
	card.Zone = state.ZonePending

	oldLogsSize := len(match.EventLog)

	if cost > 0 {
		gs.Mana[teamIndex] -= cost
		match.EventLog = append(match.EventLog, logs.SpendManaLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.SpendMana,
			TeamIndex:     teamIndex,
			Amount:        cost,
		})
		match.NextUpdateID++
	}

	gs.StandoffCardIndex = cardIndex
	for i := range gs.PendingBidAmount {
		gs.PendingBidAmount[i] = -1
	}

	match.EventLog = append(match.EventLog, logs.StandoffOpenLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.StandoffOpen,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
	})
	match.NextUpdateID++

	match.publish(executeAtWallTime, oldLogsSize)
}

// ie.CardIndex carries the sealed bid Amount; InputEvent has no dedicated field for it, so a "bid"
// event reuses this one positionally
func (match *Match) handleBid(ie *eventpq.InputEvent) {
	gs := match.GameState
	teamIndex := ie.TeamIndex
	amount := ie.CardIndex

	if gs.StandoffCardIndex < 0 || gs.StandoffCardIndex >= len(gs.Cards) {
		return
	}
	if gs.PendingBidAmount[teamIndex] != -1 {
		return // already sealed a bid for this auction
	}
	if amount < 0 || amount > gs.Mana[teamIndex] {
		return
	}

	gs.PendingBidAmount[teamIndex] = amount

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	// safe to broadcast right away: carries no amount, so it can't leak the secret
	match.EventLog = append(match.EventLog, logs.StandoffBidLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.StandoffBid,
		TeamIndex:     teamIndex,
	})
	match.NextUpdateID++

	if gs.PendingBidAmount[0] != -1 && gs.PendingBidAmount[1] != -1 {
		match.resolveStandoff(executeAtWallTime)
	}

	match.publish(executeAtWallTime, oldLogsSize)
}

// both bids are in: the higher bidder pays their bid, and if that's the activator, the opponent discards
func (match *Match) resolveStandoff(executeAtWallTime int64) {
	gs := match.GameState
	cardIndex := gs.StandoffCardIndex
	card := &gs.Cards[cardIndex]
	activatorTeam := card.Team
	opponent := 1 - activatorTeam

	bids := [2]int{gs.PendingBidAmount[0], gs.PendingBidAmount[1]}

	// higher bid wins; a tie favors whoever set this trap
	winner := activatorTeam
	if bids[opponent] > bids[activatorTeam] {
		winner = opponent
	}

	// only the winning bid is spent; the lower bidder keeps their mana
	if bids[winner] > 0 {
		gs.Mana[winner] -= bids[winner]
		match.EventLog = append(match.EventLog, logs.SpendManaLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.SpendMana,
			TeamIndex:     winner,
			Amount:        bids[winner],
		})
		match.NextUpdateID++
	}

	// only the opponent ever discards: outbidding the activator just saves their hand
	if winner == activatorTeam && len(gs.Hand[opponent]) > 0 {
		discardIndex := gs.Hand[opponent][rand.Intn(len(gs.Hand[opponent]))]
		for i, ci := range gs.Hand[opponent] {
			if ci == discardIndex {
				gs.Hand[opponent] = append(gs.Hand[opponent][:i], gs.Hand[opponent][i+1:]...)
				break
			}
		}
		gs.Cards[discardIndex].Zone = state.ZoneGraveyard
		if !gs.Cards[discardIndex].IsToken {
			gs.Grave[opponent] = append(gs.Grave[opponent], discardIndex)
		}
		match.EventLog = append(match.EventLog, logs.DiscardLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.Discard,
			TeamIndex:     opponent,
			CardIndex:     discardIndex,
		})
		match.NextUpdateID++
	}

	match.EventLog = append(match.EventLog, logs.StandoffResolutionLog{
		UpdateID:      match.NextUpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.StandoffResolution,
		CardIndex:     cardIndex,
		Bids:          bids,
		WinnerTeam:    winner,
	})
	match.NextUpdateID++

	card.Zone = state.ZoneGraveyard
	if !card.IsToken {
		gs.Grave[activatorTeam] = append(gs.Grave[activatorTeam], cardIndex)
	}

	gs.StandoffCardIndex = -1
	for i := range gs.PendingBidAmount {
		gs.PendingBidAmount[i] = -1
	}
}
