package actions

import "time"

type Action struct {
	Timestamp time.Time
	Type      string
}
