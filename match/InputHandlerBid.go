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
		gs.PendingBidChoice[i] = -1
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

// ie.CardIndex carries the sealed bid Amount and ie.TargetCardIndex carries the ChoiceIndex (0 or
// 1); InputEvent has no dedicated fields for either, so a "bid" event reuses these two positionally
func (match *Match) handleBid(ie *eventpq.InputEvent) {
	gs := match.GameState
	teamIndex := ie.TeamIndex
	amount := ie.CardIndex
	choice := ie.TargetCardIndex

	if gs.StandoffCardIndex < 0 || gs.StandoffCardIndex >= len(gs.Cards) {
		return
	}
	if gs.PendingBidAmount[teamIndex] != -1 {
		return // already sealed a bid for this auction
	}
	if amount < 0 || amount > gs.Mana[teamIndex] {
		return
	}
	if choice != 0 && choice != 1 {
		return
	}

	gs.PendingBidAmount[teamIndex] = amount
	gs.PendingBidChoice[teamIndex] = choice

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	// safe to broadcast right away: carries no amount or choice, so it can't leak the secret
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

// both bids are in: spend them both, decide the winner, and apply their chosen effect
func (match *Match) resolveStandoff(executeAtWallTime int64) {
	gs := match.GameState
	cardIndex := gs.StandoffCardIndex
	card := &gs.Cards[cardIndex]
	activatorTeam := card.Team

	bids := [2]int{gs.PendingBidAmount[0], gs.PendingBidAmount[1]}
	choices := [2]int{gs.PendingBidChoice[0], gs.PendingBidChoice[1]}

	// both bids are spent regardless of who wins
	for team := 0; team < len(gs.Mana); team++ {
		if bids[team] <= 0 {
			continue
		}
		gs.Mana[team] -= bids[team]
		match.EventLog = append(match.EventLog, logs.SpendManaLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.SpendMana,
			TeamIndex:     team,
			Amount:        bids[team],
		})
		match.NextUpdateID++
	}

	// higher bid wins; a tie favors whoever set this trap
	winner := activatorTeam
	loser := 1 - activatorTeam
	if bids[loser] > bids[winner] {
		winner, loser = loser, winner
	}
	chosenEffectIndex := choices[winner]

	switch chosenEffectIndex {
	case 0: // gain 40 mana
		const amount = 40
		gs.Mana[winner] += amount
		match.EventLog = append(match.EventLog, logs.GainManaLog{
			UpdateID:      match.NextUpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.GainMana,
			TeamIndex:     winner,
			Amount:        amount,
		})
		match.NextUpdateID++

	case 1: // deal 8 damage to a random enemy structure
		const damage = 8
		var living []int
		for i := range gs.Structures[loser] {
			if gs.Structures[loser][i].HP > 0 {
				living = append(living, i)
			}
		}
		if len(living) > 0 {
			structureIndex := living[rand.Intn(len(living))]
			if damage >= gs.Structures[loser][structureIndex].HP {
				gs.Structures[loser][structureIndex].HP = 0
			} else {
				gs.Structures[loser][structureIndex].HP -= damage
			}
			match.EventLog = append(match.EventLog, logs.AttackLog{
				UpdateID:             match.NextUpdateID,
				WallTimestamp:        executeAtWallTime,
				Type:                 logs.Attack,
				AttackerIndex:        cardIndex,
				TargetTeamIndex:      loser,
				TargetCardIndex:      -1,
				TargetStructureIndex: structureIndex,
				Damage:               damage,
			})
			match.NextUpdateID++

			if gs.AllStructuresDead(loser) {
				if gs.AllStructuresDead(winner) {
					gs.Winner = -2
				} else {
					gs.Winner = winner
				}
			}
		}
	}

	match.EventLog = append(match.EventLog, logs.StandoffResolutionLog{
		UpdateID:          match.NextUpdateID,
		WallTimestamp:     executeAtWallTime,
		Type:              logs.StandoffResolution,
		CardIndex:         cardIndex,
		Bids:              bids,
		Choices:           choices,
		WinnerTeam:        winner,
		ChosenEffectIndex: chosenEffectIndex,
	})
	match.NextUpdateID++

	card.Zone = state.ZoneGraveyard
	if !card.IsToken {
		gs.Grave[activatorTeam] = append(gs.Grave[activatorTeam], cardIndex)
	}

	gs.StandoffCardIndex = -1
	for i := range gs.PendingBidAmount {
		gs.PendingBidAmount[i] = -1
		gs.PendingBidChoice[i] = -1
	}
}
