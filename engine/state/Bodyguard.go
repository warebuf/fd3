package state

import "math/rand"

// odds a Bodyguard unit intercepts targeting meant for one of its teammates, rerolled
// independently every time a target actually gets resolved (against friendly or enemy sources
// alike, since "when any effect selects a target" doesn't distinguish who selected it)
const BodyguardChance = 0.5

// if targetIndex is a field unit belonging to a team with an eligible Bodyguard on the field (any
// Bodyguard other than targetIndex itself), there's a BodyguardChance chance it steps in and
// becomes the target instead. Returns targetIndex unchanged otherwise (no eligible Bodyguard, off
// the roll, targetIndex is itself the Bodyguard, or targetIndex isn't a field card at all).
func BodyguardRedirect(gs *GameState, targetIndex int) int {
	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return targetIndex
	}
	target := gs.Cards[targetIndex]
	if target.Zone != ZoneField {
		return targetIndex
	}

	var candidates []int
	for _, ci := range gs.Field[target.Team] {
		if ci != targetIndex && gs.Cards[ci].Bodyguard {
			candidates = append(candidates, ci)
		}
	}
	if len(candidates) == 0 || rand.Float64() >= BodyguardChance {
		return targetIndex
	}

	return candidates[rand.Intn(len(candidates))]
}
