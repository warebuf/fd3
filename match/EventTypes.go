package match

import "time"

type Event interface {
	Timestamp() time.Time
}

// Event types

type InputEvent struct {
	T       time.Time
	Type    string
	TeamIdx int
	UnitIdx int
	FuncIdx int
	Target1 int
	Target2 int
}

func (e *InputEvent) Timestamp() time.Time { return e.T }

type NewPhaseEvent struct {
	T     time.Time
	Phase int
}

func (e *NewPhaseEvent) Timestamp() time.Time { return e.T }

type EffectEvent struct {
	T time.Time
}

func (e *EffectEvent) Timestamp() time.Time { return e.T }
