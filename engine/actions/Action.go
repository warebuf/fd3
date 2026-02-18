package actions

import "time"

type Action struct {
	Timestamp time.Time
	Type      string
	UnitIndex int

	// optional fields
	FunctionIndex int
	FunctionNo    int
	TargetIndex   []int
}
