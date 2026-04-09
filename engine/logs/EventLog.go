package logs

type LogType int

const (
	PhaseLog LogType = iota // A: phase
	DrawExec                // A: team index, B: card draw
)

type Log struct {
	Timestamp int64
	UpdateID  int

	Type LogType
	A    int
	B    int
	C    int
}
