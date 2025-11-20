package match

import "time"

type Event interface {
	Timestamp() time.Time
}

type InputEvent struct {
	T    time.Time
	Type string
}

func (e *InputEvent) Timestamp() time.Time { return e.T }

type AutoEvent struct {
	T time.Time
}

func (e *AutoEvent) Timestamp() time.Time { return e.T }

type AutoScheduler struct {
	T time.Time
}

func (e *AutoScheduler) Timestamp() time.Time { return e.T }
