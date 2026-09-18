package state

// CensorForTeam deep-copies gs and hides the opponent's still-sealed Standoff bid from it, so a
// client never learns the other side's bid/choice before it's meant to be revealed. This only
// matters for a fresh "fullState" sync (initial connect or reconnect) — ongoing play is driven
// entirely by EventLog broadcasts, which are already designed to never carry a sealed bid's value
// (see logs.StandoffBidLog).
//
// Other zones (hand contents, deck order) are NOT filtered here; that's a separate, pre-existing
// gap this doesn't attempt to close.
func CensorForTeam(gs *GameState, viewingTeam int) *GameState {
	censored := DeepCopyGameState(gs)

	opponent := 1 - viewingTeam
	if opponent >= 0 && opponent < len(censored.PendingBidAmount) {
		censored.PendingBidAmount[opponent] = -1
		censored.PendingBidChoice[opponent] = -1
	}

	return censored
}
