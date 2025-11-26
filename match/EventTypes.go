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
}

func (e *InputEvent) Timestamp() time.Time { return e.T }

type AutoScheduler struct {
	T time.Time
}

func (e *AutoScheduler) Timestamp() time.Time { return e.T }

type CurrentEvent struct {
	T time.Time
}

func (e *CurrentEvent) Timestamp() time.Time { return e.T }
