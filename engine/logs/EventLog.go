package logs

type LogType int

const (
	PhaseLog LogType = iota
)

type Log struct {
	Timestamp int64
	Type      LogType
	Value     int
}
