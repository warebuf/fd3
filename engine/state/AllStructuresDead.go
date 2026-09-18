package state

// true when every one of the team's structures is destroyed
func (gs *GameState) AllStructuresDead(teamIndex int) bool {
	for i := range gs.Structures[teamIndex] {
		if gs.Structures[teamIndex][i].HP > 0 {
			return false
		}
	}
	return true
}
