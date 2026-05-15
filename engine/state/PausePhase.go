package state

type PausePhase int

const (
	PausePhaseIndefinite PausePhase = 0 // full stop, player interaction
	PausePhaseSimPause   PausePhase = 1 // chain active, sim time stopped
	PausePhaseSimulating PausePhase = 2 // real-time sim running
)
