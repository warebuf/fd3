package match

type SerializedFunc struct {
	Timestamp int

	CasterIdx int
	FuncIdx   int
	// FuncNo int ?
}

type SerializedFuncResults struct {
	TargetIdx int
	Damage    int
	Heal      int
}
