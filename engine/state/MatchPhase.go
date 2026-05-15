package state

type MatchPhase int

const (
	MatchPhasePreMatch   MatchPhase = 0
	MatchPhaseDraw       MatchPhase = 1
	MatchPhaseBoardSetup MatchPhase = 2
	MatchPhaseReveal     MatchPhase = 3
	MatchPhasePlay       MatchPhase = 4
	MatchPhaseGameEnd    MatchPhase = 5
)
