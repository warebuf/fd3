package logs

type LogType int

const (
	PhaseLog    LogType = iota // A: phase
	DrawExec                   // A: team index, B: card draw
	DeclareCard                // A: team index, B: card index
	SummonUnit                 // A: team index, B: model no
)

type Log struct {
	Timestamp int64
	UpdateID  int

	Type LogType
	A    int
	B    int
	C    int
}
